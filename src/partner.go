package say

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"math/big"
	"net"
)

type partner struct {
	Name      string         `json:"name"`
	PublicKey *rsa.PublicKey `json:"public_key"`

	IP   net.IP `json:"ip"`
	Port uint16 `json:"port"`
}

func CreatePartner(name string, publicKey_E int, publicKey_N *big.Int) *partner {
	var publicKey = rsa.PublicKey{
		E: publicKey_E,
		N: publicKey_N,
	}
	//TODO: Add IP and Port as arguments, decode the code from host to get IP and Port
	return &partner{
		Name:      name,
		PublicKey: &publicKey,
	}
}

func (p *partner) EncryptMessage(message []byte) ([]byte, error) {
	var encrypted, err = rsa.EncryptPKCS1v15(rand.Reader, p.PublicKey, message)
	if err != nil {
		return nil, errors.New("error encrypting message")
	}
	return encrypted, nil
}
