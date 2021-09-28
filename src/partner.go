package say

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"math/big"
)

type partner struct {
	Name      string
	PublicKey *rsa.PublicKey
}

func CreatePartner(name string, publicKey_E int, publicKey_N *big.Int) *partner {
	var publicKey = rsa.PublicKey{
		E: publicKey_E,
		N: publicKey_N,
	}
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
