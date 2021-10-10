package forwarding

import (
	"context"
	"net"
	"time"

	upnp "gitlab.com/NebulousLabs/go-upnp"
)

type Device struct {
	PublicIP      net.IP `json:"public_ip,omitempty"`
	ForwardedPort uint16 `json:"forwarded_port,omitempty"`

	upnpDevice *upnp.IGD
}

func CreateDevice(port uint16, description string) (*Device, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var igd, err = upnp.DiscoverCtx(ctx)
	if err != nil {
		return nil, err
	}
	err = igd.Forward(port, description)
	if err != nil {
		return nil, err
	}

	ip, err := igd.ExternalIP()
	if err != nil {
		return &Device{
			PublicIP:      nil,
			ForwardedPort: port,
			upnpDevice:    igd,
		}, err
	}

	return &Device{
		PublicIP:      net.ParseIP(ip),
		ForwardedPort: port,
		upnpDevice:    igd,
	}, nil
}

func (d *Device) Close() error {
	var err = d.upnpDevice.Clear(uint16(d.ForwardedPort))
	if err != nil {
		return err
	}
	return nil
}
