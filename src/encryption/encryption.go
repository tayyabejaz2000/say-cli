package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
)

type KeyPair struct {
	PublicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func GenerateKeyPair() *KeyPair {
	var privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("[Error: %s]: Error Generating RSA Key value pair", err.Error())
		os.Exit(-1)
	}
	var publicKey = &privateKey.PublicKey
	return &KeyPair{
		PublicKey:  publicKey,
		privateKey: privateKey,
	}
}

//Will i need this function?
func (k *KeyPair) RSAEncrypt(blob []byte) []byte {
	var encrypted, err = rsa.EncryptPKCS1v15(rand.Reader, k.PublicKey, blob)
	if err != nil {
		fmt.Printf("[Error: %s]: Error Encrypting Message", err.Error())
		os.Exit(-1)
	}
	return encrypted
}

func (k *KeyPair) RSADecrypt(encrypted []byte) []byte {
	var decrypted, err = rsa.DecryptPKCS1v15(rand.Reader, k.privateKey, encrypted)
	if err != nil {
		fmt.Printf("[Error: %s]: Error Generating RSA Key value pair", err.Error())
		os.Exit(-1)
	}
	return decrypted
}
