package say

type Message struct {
	EncryptedData []byte `json:"encrypted_data"`
	Signature     []byte `json:"signature"`
}
