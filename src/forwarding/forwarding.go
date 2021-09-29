package forwarding

import (
	"context"
	"encoding/binary"
	"errors"
	"math/big"
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
		return nil, errors.New("error initializing upnp device")
	}
	err = igd.Forward(port, description)
	if err != nil {
		return nil, errors.New("error forwarding port")
	}
	ip, err := igd.ExternalIP()

	if err != nil {
		return &Device{
			PublicIP:      nil,
			ForwardedPort: port,
			upnpDevice:    igd,
		}, errors.New("error retrieving public ip for device")
	}

	return &Device{
		PublicIP:      net.ParseIP(ip),
		ForwardedPort: port,
		upnpDevice:    igd,
	}, nil
}

func (d *Device) GetCoded() string {
	var ip2int = func(ip net.IP) uint32 {
		if len(ip) == 16 {
			return binary.BigEndian.Uint32(ip[12:16])
		}
		return binary.BigEndian.Uint32(ip)
	}

	return big.NewInt(int64(ip2int(d.PublicIP))).Text(62) + "-" + big.NewInt(int64(d.ForwardedPort)).Text(62)
}

func (d *Device) Close() error {
	var err = d.upnpDevice.Clear(uint16(d.ForwardedPort))
	if err != nil {
		return errors.New("error removing forwarded port")
	}
	return nil
}
