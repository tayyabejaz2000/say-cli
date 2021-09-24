package say

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"
	"os"
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

func (p *partner) EncryptMessage(message []byte) []byte {
	var encrypted, err = rsa.EncryptPKCS1v15(rand.Reader, p.PublicKey, message)
	if err != nil {
		fmt.Printf("[Error: %s]: Error Encrypting Message", err.Error())
		os.Exit(-1)
	}
	return encrypted
}
