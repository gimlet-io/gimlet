package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

type EncryptionMeddler struct {
	// Has to be 32 bytes long
	EnryptionKey string
}

// PreRead is called before a Scan operation for fields that have the EncryptionMeddler
func (m EncryptionMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// give a pointer to a byte buffer to grab the raw data
	return new(string), nil
}

// PostRead is called after a Scan operation for fields that have the EncryptionMeddler
func (m EncryptionMeddler) PostRead(fieldAddr, scanTarget interface{}) error {
	ptr := scanTarget.(*string)
	if ptr == nil {
		return fmt.Errorf("EncryptionMeddler.PostRead: nil pointer")
	}
	raw := *ptr

	plaintextBytes, err := decrypt([]byte(raw), []byte(m.EnryptionKey))
	fieldAddrStringPtr := fieldAddr.(*string)
	*fieldAddrStringPtr = string(plaintextBytes)
	return err
}

// PreWrite is called before an Insert or Update operation for fields that have the EncryptionMeddler
func (m EncryptionMeddler) PreWrite(field interface{}) (saveValue interface{}, err error) {
	return encrypt([]byte(field.(string)), []byte(m.EnryptionKey))
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
