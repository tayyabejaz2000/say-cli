package say

import (
	"fmt"
	"math/big"
	"net"
	"say/src/encryption"
	"say/src/forwarding"
	"strconv"
	"strings"
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
	RSAKeyPair *encryption.KeyPair
	Device     *forwarding.Device
	AppConfig  *Config
	Other      *partner
}

func CreateChatApp(config *Config) (*chatapp, error) {
	var port = config.Port
	var description = config.PortDescription
	var device *forwarding.Device = nil

	//Forward Port by UPnP if not running in Local
	if !config.IsLocal {
		var createdDevice, err = forwarding.CreateDevice(port, description)
		//Run in local is Port Forwarding failed
		if err != nil {
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
		return nil, err
	}

	return &chatapp{
		RSAKeyPair: keyPair,
		Device:     device,
		AppConfig:  config,
		Other:      nil,
	}, nil
}

func (c *chatapp) runHost() {
	//Open TCP Socket
	var listener, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", "127.0.0.1", c.AppConfig.Port))
	if err != nil {
		fmt.Printf("[Error: %s]: Error opening TCP socket\n", err.Error())
	}
	//Accept Client Connection
	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("[Error: %s]: Error accepting client connection\n", err.Error())
	}

	//Exchange Names
	var name = "Host"
	if c.AppConfig.BroadcastName {
		name = c.AppConfig.Name
	}
	_, err = conn.Write([]byte(name)) //Send host name to client
	if err != nil {
		fmt.Printf("[Error: %s]: Error sending host name\n", err.Error())
	}
	var clientName = make([]byte, 50)
	_, err = conn.Read(clientName) //Read client name from client
	if err != nil {
		fmt.Printf("[Error: %s]: Error receiving client name\n", err.Error())
	}

	//Exchange Public Keys
	var hostPublicKey = c.RSAKeyPair.PublicKey
	var keyBlob = []byte(fmt.Sprintf("%s,%d", hostPublicKey.N.String(), hostPublicKey.E))
	_, err = conn.Write(keyBlob) //Send host public key to client
	if err != nil {
		fmt.Printf("[Error: %s]: Error sending host public key\n", err.Error())
	}
	var clientKey = make([]byte, 1000)
	_, err = conn.Read(clientKey) //Read client public key from client
	if err != nil {
		fmt.Printf("[Error: %s]: Error receiving client public key\n", err.Error())
	}
	//Split public key components
	var clientPublicKey = strings.Split(string(clientKey), ",")

	//Delete Buffers
	keyBlob = nil
	clientKey = nil
	clientName = nil

	publicKey_N, _ := new(big.Int).SetString(clientPublicKey[0], 10)
	publicKey_E, _ := strconv.Atoi(clientPublicKey[1])
	//Fill partner data
	c.Other = CreatePartner(string(clientName), publicKey_E, publicKey_N)

	/*
		TODO: Add Chat
	*/

	//Close Client Connection
	err = conn.Close()
	if err != nil {
		fmt.Printf("[Error: %s]: Error closing connection to client\n", err.Error())
	}
	//Close TCP Socket
	err = listener.Close()
	if err != nil {
		fmt.Printf("[Error: %s]: Error closing host connection\n", err.Error())
	}
}

func (c *chatapp) runClient() {
	var conn, err = net.Dial("tcp4", fmt.Sprintf("%s:%d", "127.0.0.1", c.AppConfig.Port))
	if err != nil {
		fmt.Printf("[Error: %s]: Error connecting to host\n", err.Error())
	}

	//Exchange Names
	var hostName = make([]byte, 50)
	_, err = conn.Read(hostName) //Read host name from host
	if err != nil {
		fmt.Printf("[Error: %s]: Error receiving host name\n", err.Error())
	}

	var name = "Client"
	if c.AppConfig.BroadcastName {
		name = c.AppConfig.Name
	}
	_, err = conn.Write([]byte(name)) //Send client name to host
	if err != nil {
		fmt.Printf("[Error: %s]: Error sending client name\n", err.Error())
	}

	//Exchange public keys
	var hostKey = make([]byte, 1000)
	_, err = conn.Read(hostKey) //Read host public key from host
	if err != nil {
		fmt.Printf("[Error: %s]: Error receiving host public key\n", err.Error())
	}

	var clientPubicKey = c.RSAKeyPair.PublicKey
	var keyBlob = []byte(fmt.Sprintf("%s,%d", clientPubicKey.N.String(), clientPubicKey.E))

	_, err = conn.Write(keyBlob) //Send client public key to host
	if err != nil {
		fmt.Printf("[Error: %s]: Error sending client public key\n", err.Error())
	}
	//Split public key components
	var hostPublicKey = strings.Split(string(hostKey), ",")

	//Delete Buffers
	keyBlob = nil
	hostKey = nil
	hostName = nil

	publicKey_N, _ := new(big.Int).SetString(hostPublicKey[0], 10)
	publicKey_E, _ := strconv.Atoi(hostPublicKey[1])
	//Fill partner data
	c.Other = CreatePartner(string(hostName), publicKey_E, publicKey_N)

	/*
		TODO: Add Chat
	*/

	//Close host connection
	err = conn.Close()
	if err != nil {
		fmt.Printf("[Error: %s]: Error closing connection to host\n", err.Error())
	}
}

func (c *chatapp) Run() {
	if c.AppConfig.IsHost {
		c.runHost()
	} else {
		c.runClient()
	}
	fmt.Println(*c)
}

func (c *chatapp) Clean() {
	if c.Device != nil {
		var err = c.Device.Close()
		if err != nil {
			fmt.Printf("[Error: %s]: Error closing\n", err.Error())
		}
	}
}
