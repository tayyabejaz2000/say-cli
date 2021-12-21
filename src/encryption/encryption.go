package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"errors"
)

type KeyPair struct {
	PublicKey  *rsa.PublicKey `json:"public_key"`
	privateKey *rsa.PrivateKey
}

type AESKey struct {
	block *cipher.Block
	key   []byte
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
		return nil, err
	}
	return encrypted, nil
}

func (k *KeyPair) RSADecrypt(encrypted []byte) ([]byte, error) {
	var decrypted, err = rsa.DecryptPKCS1v15(rand.Reader, k.privateKey, encrypted)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func (k *KeyPair) RSASign(data []byte) ([]byte, error) {
	var signed, err = rsa.SignPKCS1v15(rand.Reader, k.privateKey, 0, data)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

func RSAVerify(publicKey *rsa.PublicKey, data []byte, sig []byte) error {
	return rsa.VerifyPKCS1v15(publicKey, 0, data, sig)
}

func GenerateAESKey() (*AESKey, error) {
	var key = make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESKey{
		block: &block,
		key:   key,
	}, nil
}

func CreateAESKey(key []byte) (*AESKey, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESKey{
		block: &block,
		key:   nil,
	}, nil
}

func (k *AESKey) Encrypt(blob []byte) []byte {
	if len(blob) < 16 {
		temp := make([]byte, 16)
		copy(temp, blob)
		blob = temp
	}

	encrypted := make([]byte, len(blob))
	size := 16

	for bs, be := 0, size; bs < len(blob); bs, be = bs+size, be+size {
		(*k.block).Encrypt(encrypted[bs:be], blob[bs:be])
	}

	return encrypted
}
func (k *AESKey) Decrypt(blob []byte) []byte {
	if len(blob) < 16 {
		temp := make([]byte, 16)
		copy(temp, blob)
		blob = temp
	}

	decrypted := make([]byte, len(blob))
	size := 16

	for bs, be := 0, size; bs < len(blob); bs, be = bs+size, be+size {
		(*k.block).Decrypt(decrypted[bs:be], blob[bs:be])
	}

	return decrypted
}

func (k *AESKey) EncryptKeyByRSA(rsaKey rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, &rsaKey, k.key)
}
