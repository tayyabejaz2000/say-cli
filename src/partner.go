package say

import (
	"crypto/rand"
	"crypto/rsa"
)

type partner struct {
	Name      string         `json:"name"`
	PublicKey *rsa.PublicKey `json:"public_key"`
}

func CreatePartner(name string, publicKey rsa.PublicKey) *partner {
	//TODO: Add IP and Port as arguments, decode the code from host to get IP and Port
	return &partner{
		Name:      name,
		PublicKey: &publicKey,
	}
}

func (p *partner) EncryptMessage(message []byte) ([]byte, error) {
	var encrypted, err = rsa.EncryptPKCS1v15(rand.Reader, p.PublicKey, message)
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}
