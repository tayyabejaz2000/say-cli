package say

import (
	"fmt"
	"net"
	"say/src/encryption"
	"say/src/forwarding"
)

///TODO: (bool)Read Key Value Pair from cert/file
type Config struct {
	Name            string `json:"name"`
	BroadcastName   bool   `json:"broadcast_name"`
	IsLocal         bool   `json:"is_local"`
	Port            uint16 `json:"port"`
	PortDescription string `json:"port_description"`
}

type chatapp struct {
	ClientKeyPair *encryption.KeyPair
	Device        *forwarding.Device
	AppConfig     *Config
	Other         *partner
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
		ClientKeyPair: keyPair,
		Device:        device,
		AppConfig:     config,
		Other:         nil,
	}, nil
}

func (c *chatapp) Run() {
	var listener, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", "127.0.0.1", c.AppConfig.Port))
	if err != nil {
		fmt.Printf("[Error: %s]: Error Opening TCP Socket\n", err.Error())
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("[Error: %s]: Error Accepting Connection\n", err.Error())
	}

	/*
		TODO: Actual App Code
	*/

	err = conn.Close()
	if err != nil {
		fmt.Printf("[Error: %s]: Error Closing Client Connection\n", err.Error())
	}

	err = listener.Close()
	if err != nil {
		fmt.Printf("[Error: %s]: Error Closing Host Connection\n", err.Error())
	}
}

func (c *chatapp) Clean() {
	if c.Device != nil {
		var err = c.Device.Close()
		if err != nil {
			fmt.Printf("[Error: %s]: Error Closing\n", err.Error())
		}
	}
}
