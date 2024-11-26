package asymcrypto

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

var path string

func setup(t *testing.T) {
	var err error
	path, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetPrivateKey(t *testing.T) {
	privateKeyPath := filepath.Join(path, "private.pem")
	got, err := GetPrivateKey(privateKeyPath)
	if err != nil {
		t.Errorf("GetPrivateKey() error = %v", err)
		return
	}
	if got == nil {
		t.Errorf("GetPrivateKey() got = %v", got)
		return
	}
}

func TestGetPublicKey(t *testing.T) {
	publicKeyPath := filepath.Join(path, "public.pem")
	got, err := GetPublicKey(publicKeyPath)
	if err != nil {
		t.Errorf("GetPublicKey() error = %v", err)
		return
	}
	if got == nil {
		t.Errorf("GetPublicKey() got = %v", got)
		return
	}
}

func TestEncrypt(t *testing.T) {
	publicKeyPath := filepath.Join(path, "public.pem")
	privateKeyPath := filepath.Join(path, "private.pem")
	publicKey, err := GetPublicKey(publicKeyPath)
	if err != nil {
		t.Errorf("GetPublicKey() error = %v", err)
	}
	privateKey, err := GetPrivateKey(privateKeyPath)
	if err != nil {
		t.Errorf("GetPublicKey() error = %v", err)
	}

	msg := []byte("Hello world!")

	sessionKey, encrypted, err := Encrypt(publicKey, msg)
	if err != nil {
		t.Errorf("Encrypt() error = %v", err)
	}
	got, err := Decrypt(privateKey, sessionKey, encrypted)
	if err != nil {
		t.Errorf("Decrypt() error = %v", err)
	}
	if !bytes.Equal(got, msg) {
		t.Errorf("want %v, get %v", string(msg), string(got))
	}
}
