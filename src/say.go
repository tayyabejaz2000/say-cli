package say

import (
	"say/src/encryption"
	"say/src/forwarding"
)

///TODO: (bool)Read Key Value Pair from cert/file
type Config struct {
	Host            bool   `json:"host"`
	Name            string `json:"name"`
	BroadcastName   bool   `json:"broadcast_name"`
	IsLocal         bool   `json:"is_local"`
	Port            int    `json:"port"`
	PortDescription string `json:"port_description"`
}

type chatapp struct {
	ClientKeyPair *encryption.KeyPair
	Device        *forwarding.Device
	AppConfig     *Config
	Other         *partner
}

func CreateChatApp(config *Config) *chatapp {
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
	var keyPair = encryption.GenerateKeyPair()

	return &chatapp{
		ClientKeyPair: keyPair,
		Device:        device,
		Other:         nil, //No connection uptill now
		AppConfig:     config,
	}
}
