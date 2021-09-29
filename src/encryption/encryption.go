package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
)

type KeyPair struct {
	PublicKey  *rsa.PublicKey `json:"public_key"`
	privateKey *rsa.PrivateKey
}

func GenerateKeyPair() (*KeyPair, error) {
	var privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.New("error generating rsa key pair")
	}
	var publicKey = &privateKey.PublicKey
	return &KeyPair{
		PublicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

//Will i need this function?
func (k *KeyPair) RSAEncrypt(blob []byte) ([]byte, error) {
	var encrypted, err = rsa.EncryptPKCS1v15(rand.Reader, k.PublicKey, blob)
	if err != nil {
		return nil, errors.New("error encrypting message")
	}
	return encrypted, nil
}

func (k *KeyPair) RSADecrypt(encrypted []byte) ([]byte, error) {
	var decrypted, err = rsa.DecryptPKCS1v15(rand.Reader, k.privateKey, encrypted)
	if err != nil {
		return nil, errors.New("error encrypting message")
	}
	return decrypted, nil
}
