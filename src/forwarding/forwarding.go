package forwarding

import (
	"fmt"

	upnp "gitlab.com/NebulousLabs/go-upnp"
)

type Device struct {
	PublicIP      string
	ForwardedPort int16

	upnpDevice *upnp.IGD
}

func CreateDevice(port int, description string) (*Device, error) {
	var igd, err = upnp.Discover()
	if err != nil {
		fmt.Printf("[Error: %s]: Error Initializing UPnP Device\n", err.Error())
		return nil, err
	}
	err = igd.Forward(uint16(port), description)
	if err != nil {
		fmt.Printf("[Error: %s]: Error Forwarding Port %d\n", err.Error(), port)
		return nil, err
	}
	ip, err := igd.ExternalIP()
	if err != nil {
		fmt.Printf("[Error: %s]: Error retrieving Public IP for Device\n", err.Error())
		return &Device{
			PublicIP:      "",
			ForwardedPort: int16(port),
			upnpDevice:    igd,
		}, err
	}

	return &Device{
		PublicIP:      ip,
		ForwardedPort: int16(port),
		upnpDevice:    igd,
	}, nil
}

func (d *Device) Close() {
	d.upnpDevice.Clear(uint16(d.ForwardedPort))
}
