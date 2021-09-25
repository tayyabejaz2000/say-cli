package forwarding

import (
	"context"
	"fmt"
	"time"

	upnp "gitlab.com/NebulousLabs/go-upnp"
)

type Device struct {
	PublicIP      string
	ForwardedPort uint16

	upnpDevice *upnp.IGD
}

func CreateDevice(port uint16, description string) (*Device, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var igd, err = upnp.DiscoverCtx(ctx)
	if err != nil {
		fmt.Printf("[Error: %s]: Error Initializing UPnP Device\n", err.Error())
		return nil, err
	}
	err = igd.Forward(port, description)
	if err != nil {
		fmt.Printf("[Error: %s]: Error Forwarding Port %d\n", err.Error(), port)
		return nil, err
	}
	ip, err := igd.ExternalIP()
	if err != nil {
		fmt.Printf("[Error: %s]: Error retrieving Public IP for Device\n", err.Error())
		return &Device{
			PublicIP:      "",
			ForwardedPort: port,
			upnpDevice:    igd,
		}, err
	}

	return &Device{
		PublicIP:      ip,
		ForwardedPort: port,
		upnpDevice:    igd,
	}, nil
}

func (d *Device) Close() {
	d.upnpDevice.Clear(uint16(d.ForwardedPort))
}
