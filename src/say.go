package say

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/signal"
	"say/src/encryption"
	"say/src/forwarding"
	"strings"
	"syscall"
)

///TODO: (bool)Read Key Value Pair from cert/file
type Config struct {
	Name            string `json:"name"`
	BroadcastName   bool   `json:"broadcast_name"`
	IsLocal         bool   `json:"is_local"`
	IsHost          bool   `json:"is_host"`
	Port            uint16 `json:"port"`
	PortDescription string `json:"port_description"`
}

type chatapp struct {
	Device    *forwarding.Device
	AppConfig *Config
	KeyPair   *encryption.KeyPair
	AESKey    *encryption.AESKey
	IP        string
}

type HandshakeMessage1 struct {
	Name      string        `json:"name"`
	PublicKey rsa.PublicKey `json:"public_key"`
}

type HandshakeMessage2 struct {
	Name            string        `json:"name"`
	PublicKey       rsa.PublicKey `json:"public_key"`
	EncryptedAESKey []byte        `json:"encrypted_aes_key"`
}

func CreateChatApp(config *Config) *chatapp {
	var port = config.Port
	var description = config.PortDescription
	var device *forwarding.Device = nil
	//Forward Port by UPnP if not running in Local and this is the host app
	if !config.IsLocal && config.IsHost {
		var createdDevice, err = forwarding.CreateDevice(port, description)
		//Run in local is Port Forwarding failed
		if err != nil {
			log.Printf("[Warning: %s]: Couldn't forward port, running in local mode\n", err.Error())
			config.IsLocal = true
			//Close the port if it was already open
			if createdDevice != nil {
				createdDevice.Close()
			}
		} else {
			device = createdDevice
		}
	}
	keyPair, err := encryption.GenerateKeyPair()
	if err != nil {
		log.Panicf("[Error: %s]: Error generating RSA Key Pair\n", err.Error())
	}
	var AESKey *encryption.AESKey = nil
	if config.IsHost {
		var err error
		AESKey, err = encryption.GenerateAESKey()
		if err != nil {
			log.Panicf("[Error: %s]: Error generating AES Key\n", err.Error())
		}
	}

	logFile, err := os.OpenFile("say.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("[Warning: %s]: Failed to create/open Log File\n", err.Error())
	} else {
		log.SetOutput(logFile)
	}

	var app = &chatapp{
		Device:    device,
		AppConfig: config,
		KeyPair:   keyPair,
		AESKey:    AESKey,
	}

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-termChan
		log.Print("[Info]: Ctrl-C shutting down...\n")
		app.Clean()
		os.Exit(0)
	}()

	return app
}

func getCode(IP net.IP, port uint16) string {
	var ip2int = func(ip net.IP) uint32 {
		if len(ip) == 16 {
			return binary.BigEndian.Uint32(ip[12:16])
		}
		return binary.BigEndian.Uint32(ip)
	}

	return big.NewInt(int64(ip2int(IP))).Text(62) + "-" + big.NewInt(int64(port)).Text(62)
}

func (c *chatapp) runHost() {
	//Open TCP Socket
	var listener, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", c.IP, c.AppConfig.Port))
	if err != nil {
		log.Panicf("[Error: %s]: Error opening TCP socket\n", err.Error())
	}
	defer listener.Close()

	//Accept Client Connection
	conn, err := listener.Accept()
	if err != nil {
		log.Panicf("[Error: %s]: Error accepting client connection\n", err.Error())
	}
	defer conn.Close()

	var connWriter = json.NewEncoder(conn)
	var connReader = json.NewDecoder(conn)

	var clientData HandshakeMessage1
	err = connReader.Decode(&clientData)
	if err != nil {
		log.Panicf("[Error: %s]: Error reading RSA Public Key from client\n", err.Error())
	}

	fmt.Printf("New Connection Request by Name: %s, IP: %s\n", clientData.Name, conn.RemoteAddr().String())

	var allowConnection bool = false
	fmt.Print("Allow Connection [true/FALSE]:")
	fmt.Scanf("%t\n", &allowConnection)
	if !allowConnection {
		return
	}

	encryptedKey, err := c.AESKey.EncryptKeyByRSA(clientData.PublicKey)
	if err != nil {
		log.Panicf("[Error: %s]: Error encrypting AES Key\n", err.Error())
	}
	var hostData = &HandshakeMessage2{
		Name:            c.AppConfig.Name,
		PublicKey:       *c.KeyPair.PublicKey,
		EncryptedAESKey: encryptedKey,
	}
	err = connWriter.Encode(hostData)
	if err != nil {
		log.Panicf("[Error: %s]: Error sending encrypted AES Key to client\n", err.Error())
	}
	log.Println("[Info]: Starting Chat UI...")
	ui := CreateUI(c.AppConfig, func(message string) {
		encrypted := c.AESKey.Encrypt([]byte(message))
		shaHash := sha256.Sum256(encrypted)
		sig, err := c.KeyPair.RSASign(shaHash[:])
		if err != nil {
			log.Panicf("[Error: %s]: Error Signing Message\n", err.Error())
		}
		var msg = Message{
			EncryptedData: encrypted,
			Signature:     sig,
		}
		err = connWriter.Encode(msg)
		if err != nil {
			log.Panicf("[Error: %s]: Error sending encrypted Message to client\n", err.Error())
		}
		log.Printf("[Message]: (%s) %s\n", hostData.Name, message)
	})
	go ui.Run(c.Clean)

	for {
		var msg Message
		err = connReader.Decode(&msg)
		if err != nil {
			log.Panicf("[Error: %s]: Error parsing Message from client\n", err.Error())
		}
		shaHash := sha256.Sum256([]byte(msg.EncryptedData))
		err = encryption.RSAVerify(&clientData.PublicKey, shaHash[:], msg.Signature)
		if err != nil {
			log.Panicf("[Error: %s]: Message Compromised, Digital Signature don't match\n", err.Error())
		}
		var decryptedMessage = c.AESKey.Decrypt(msg.EncryptedData)
		ui.AddMessage(clientData.Name, string(decryptedMessage))
		log.Printf("[Message]: (%s) %s\n", clientData.Name, strings.Trim(string(decryptedMessage), string([]byte{0, ' '})))
	}
}

func (c *chatapp) runClient() {
	fmt.Print("Enter Code: ")
	var code string
	fmt.Scanln(&code)
	var codeParts = strings.Split(code, "-")
	var codedIP, _ = new(big.Int).SetString(codeParts[0], 62)
	var ip net.IP = net.IPv4(0, 0, 0, 0)
	binary.BigEndian.PutUint32(ip[12:16], uint32(codedIP.Uint64()))

	var codedPort, _ = new(big.Int).SetString(codeParts[1], 62)
	var port = codedPort.Uint64()

	var conn, err = net.Dial("tcp4", fmt.Sprintf("%s:%d", ip.String(), port))
	if err != nil {
		log.Panicf("[Error: %s]: Error connecting to host\n", err.Error())
	}
	defer conn.Close()

	var connWriter = json.NewEncoder(conn)
	var connReader = json.NewDecoder(conn)

	var clientData = HandshakeMessage1{
		Name:      c.AppConfig.Name,
		PublicKey: *c.KeyPair.PublicKey,
	}

	err = connWriter.Encode(clientData)
	if err != nil {
		log.Panicf("[Error: %s]: Error sending Public Key to host\n", err.Error())
	}

	var hostData HandshakeMessage2
	err = connReader.Decode(&hostData)
	if err != nil {
		log.Panicf("[Error: %s]: Error recieving AES Key from host\n", err.Error())
	}
	aesKey, err := c.KeyPair.RSADecrypt(hostData.EncryptedAESKey)
	if err != nil {
		log.Panicf("[Error: %s]: Error decrypting AES Key\n", err.Error())
	}
	c.AESKey, err = encryption.CreateAESKey(aesKey)
	if err != nil {
		log.Panicf("[Error: %s]: Couldn't create AES Block Cipher\n", err.Error())
	}

	log.Println("[Info]: Starting Chat UI...")
	ui := CreateUI(c.AppConfig, func(message string) {
		encrypted := c.AESKey.Encrypt([]byte(message))
		shaHash := sha256.Sum256(encrypted)
		sig, err := c.KeyPair.RSASign(shaHash[:])
		if err != nil {
			log.Panicf("[Error: %s]: Error Signing Message\n", err.Error())
		}
		var msg = Message{
			EncryptedData: encrypted,
			Signature:     sig,
		}
		err = connWriter.Encode(msg)
		if err != nil {
			log.Panicf("[Error: %s]: Error sending encrypted Message to client\n", err.Error())
		}
		log.Printf("[Message]: (%s) %s\n", clientData.Name, message)
	})
	go ui.Run(c.Clean)

	for {
		var msg Message
		err = connReader.Decode(&msg)
		if err != nil {
			log.Panicf("[Error: %s]: Error parsing Message from client\n", err.Error())
		}
		shaHash := sha256.Sum256([]byte(msg.EncryptedData))
		err = encryption.RSAVerify(&hostData.PublicKey, shaHash[:], msg.Signature)
		if err != nil {
			log.Panicf("[Error: %s]: Message Compromised, Digital Signature don't match\n", err.Error())
		}
		var decryptedMessage = c.AESKey.Decrypt(msg.EncryptedData)
		ui.AddMessage(hostData.Name, string(decryptedMessage))

		log.Printf("[Message]: (%s) %s\n", hostData.Name, strings.Trim(string(decryptedMessage), string([]byte{0, ' '})))
	}
}

func (c *chatapp) Run() {
	config, _ := json.Marshal(c.AppConfig)
	log.Printf("[Info]: Config %s\n", string(config))
	if c.AppConfig.IsHost {
		if !c.AppConfig.IsLocal {
			//Use this code for connection between host-client
			c.IP = c.Device.PublicIP.String()
			code := getCode(c.Device.PublicIP, c.Device.ForwardedPort)
			fmt.Printf("Connection Code: %v\n", code)
			log.Printf("[Info]: Connection Code: %v\n", code)
		} else {
			var localIP = net.ParseIP("127.0.0.1")
			var localConn, err = net.Dial("udp", "1.1.1.1:80")
			if err != nil {
				//Can be a panic wont allow running if not connected to a network
				c.IP = "127.0.0.1"
				log.Printf("[Warning: %s]: You are not connected to a network", err.Error())
			} else {
				localIP = net.ParseIP(strings.Split(localConn.LocalAddr().String(), ":")[0])
				localConn.Close()
			}
			c.IP = localIP.String()
			code := getCode(localIP, c.AppConfig.Port)
			fmt.Printf("Connection Code: %v\n", code)
			log.Printf("[Info]: Connection Code: %v\n", code)
		}
		c.runHost()
	} else {
		c.runClient()
	}
}

func (c *chatapp) Clean() {
	if c.Device != nil {
		var err = c.Device.Close()
		if err != nil {
			log.Printf("[Warning: %s]: Failed to close forwarded port\n", err.Error())
		}
	}
}
