package asymcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
)

// PrivateKey alias
type PrivateKey = rsa.PrivateKey

// PublicKey alias
type PublicKey = rsa.PublicKey

func getCommonKey(path string) (*pem.Block, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyBytes)
	return block, nil
}

// GetPublicKey returns the public key from the given path
func GetPublicKey(path string) (*PublicKey, error) {
	block, err := getCommonKey(path)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// GetPrivateKey returns the private key from the given path
func GetPrivateKey(path string) (*PrivateKey, error) {
	block, err := getCommonKey(path)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// Encrypt encrypts the data using the RSA public key and returns the encrypted session key and encrypted data
func Encrypt(publicKey *rsa.PublicKey, data []byte) (encryptedSessionKey []byte, encryptedData []byte, err error) {
	// Generate a random session key (AES key)
	sessionKey := make([]byte, 32) // 256-bit AES key
	_, err = io.ReadFull(rand.Reader, sessionKey)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt the session key with the RSA public key
	encryptedSessionKey, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, sessionKey)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt the large message with the session key
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, nil, err
	}
	iv := sessionKey[:aes.BlockSize]
	cfb := cipher.NewCFBEncrypter(block, iv)
	encryptedData = make([]byte, len(data))
	cfb.XORKeyStream(encryptedData, data)
	return encryptedSessionKey, encryptedData, nil
}

// Decrypt decrypts the data using the RSA private key and returns the decrypted data
func Decrypt(privateKey *rsa.PrivateKey, encryptedSessionKey []byte, encryptedData []byte) (decryptedData []byte, err error) {
	// Decrypt the session key using the private key
	sessionKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedSessionKey)
	if err != nil {
		return nil, err
	}

	// Decrypt the message blocks using the decrypted session key
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}
	iv := sessionKey[:aes.BlockSize]
	cfb := cipher.NewCFBDecrypter(block, iv)
	decryptedData = make([]byte, len(encryptedData))
	cfb.XORKeyStream(decryptedData, encryptedData)
	return decryptedData, nil
}
