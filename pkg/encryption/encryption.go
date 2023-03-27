package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func DecryptData(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	label := []byte("")
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, label)
}

func EncryptData(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	label := []byte("")
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, label)
}

func parseFile(filePath string) (*pem.Block, error) {
	keyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block from private key file %s", filePath)
	}
	return block, nil
}

func NewPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	block, err := parseFile(filePath)
	if err != nil {
		return nil, err
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func NewPublicKey(filePath string) (*rsa.PublicKey, error) {
	block, err := parseFile(filePath)
	if err != nil {
		return nil, err
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key file %s contains a non-RSA key", filePath)
	}
	return rsaPubKey, nil
}
