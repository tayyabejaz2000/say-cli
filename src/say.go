package say

import (
	"crypto/rsa"
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
	RSAKeyPair *encryption.KeyPair `json:"rsa_key_pair"`
	Device     *forwarding.Device  `json:"device"`
	AppConfig  *Config             `json:"app_config"`
	Other      *partner            `json:"other"`
}

type DataHeader struct {
	Name      string        `json:"name"`
	PublicKey rsa.PublicKey `json:"public_key"`
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

	//Can do it once both parties join
	var keyPair, err = encryption.GenerateKeyPair()
	if err != nil {
		log.Panicf("[Error: %s]: Error generating RSA Key Pair\n", err.Error())
	}

	var app = &chatapp{
		RSAKeyPair: keyPair,
		Device:     device,
		AppConfig:  config,
		Other:      nil,
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
	var listener, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", "127.0.0.1", c.AppConfig.Port))
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

	//Send host data to client
	var name = "Host"
	if c.AppConfig.BroadcastName {
		name = c.AppConfig.Name
	}
	err = connWriter.Encode(DataHeader{name, *c.RSAKeyPair.PublicKey})
	if err != nil {
		log.Panicf("[Error: %s]: Error sending data to client\n", err.Error())
	}

	//Read client data from client
	var clientData DataHeader
	err = connReader.Decode(&clientData)
	if err != nil {
		log.Panicf("[Error: %s]: Error recieving data from client\n", err.Error())
	}
	c.Other = CreatePartner(clientData.Name, clientData.PublicKey)

	//Chatting can begin
	var message = "Hello World"
	encrypted, _ := c.Other.EncryptMessage([]byte(message))
	err = connWriter.Encode(Message{encrypted})
	if err != nil {
		log.Printf("[Error: %s]: Error sending message to client\n", err.Error())
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

	//Recieve host data from host
	var hostData DataHeader
	err = connReader.Decode(&hostData)
	if err != nil {
		log.Panicf("[Error: %s]: Error recieving data from host\n", err.Error())
	}
	c.Other = CreatePartner(hostData.Name, hostData.PublicKey)

	//Send client data to host
	var name = "Client"
	if c.AppConfig.BroadcastName {
		name = c.AppConfig.Name
	}
	err = connWriter.Encode(DataHeader{name, *c.RSAKeyPair.PublicKey})
	if err != nil {
		log.Panicf("[Error: %s]: Error sending data to host\n", err.Error())
	}

	//Chatting can begin
	var message Message
	err = connReader.Decode(&message)
	if err != nil {
		log.Printf("[Error: %s]: Error receiving message from host\n", err.Error())
	}
	decrypted, _ := c.RSAKeyPair.RSADecrypt(message.EncryptedData)
	fmt.Printf("Message: %v\n", string(decrypted))
}

func (c *chatapp) Run() {
	if c.AppConfig.IsHost {
		if !c.AppConfig.IsLocal {
			//Use this code for connection between host-client
			log.Printf("Your Code: %v\n", getCode(c.Device.PublicIP, c.Device.ForwardedPort))
		} else {
			var localIP = net.ParseIP("127.0.0.1")
			var localConn, err = net.Dial("udp", "1.1.1.1:80")
			if err != nil {
				//Can be a panic wont allow running if not connected to a network
				log.Printf("[Warning: %s]: You are not connected to a network", err.Error())
			} else {
				localIP = net.ParseIP(strings.Split(localConn.LocalAddr().String(), ":")[0])
				localConn.Close()
			}
			log.Printf("Your Code: %v\n", getCode(localIP, c.AppConfig.Port))
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
