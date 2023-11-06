//go:build encryption

// Copyright 2019 Laszlo Fogas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/russross/meddler"
	"github.com/stretchr/testify/assert"
)

type Dummy struct {
	ID     int64  `json:"-" meddler:"id,pk"`
	Secret string `json:"-" meddler:"secret,encrypted"`
}

func TestEncryption(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	if os.Getenv("DATABASE_DRIVER") != "" {
		_, err := s.Exec("CREATE TABLE IF NOT EXISTS dummy (id SERIAL, secret text);")
		assert.Nil(t, err)
	} else {
		_, err := s.Exec("CREATE TABLE IF NOT EXISTS dummy (id INTEGER PRIMARY KEY AUTOINCREMENT,secret text);")
		assert.Nil(t, err)
	}

	err := meddler.Insert(s, "dummy", &Dummy{
		Secret: "superSecretValue",
	})
	assert.Nil(t, err)

	rawData := s.QueryRow("select secret from dummy where id = 1")
	encrypteSecretValue := new([]byte)
	err = rawData.Scan(encrypteSecretValue)
	assert.Nil(t, err)
	fmt.Println(string(*encrypteSecretValue))
	assert.NotEqual(t, "superSecretValue", string(*encrypteSecretValue))

	fromDB := new(Dummy)
	err = meddler.Load(s, "dummy", fromDB, 1)
	assert.Nil(t, err)
	assert.Equal(t, "superSecretValue", fromDB.Secret)
}

func TestReEncryption(t *testing.T) {
	encryptionKey := "the-key-has-to-be-32-bytes-long!"
	encryptionKeyNew := "new-key-has-to-be-32-bytes-long!"
	superSecretValue := "superSecretValue"
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	if os.Getenv("DATABASE_DRIVER") != "" {
		_, err := s.Exec("CREATE TABLE IF NOT EXISTS dummy (id SERIAL, secret text);")
		assert.Nil(t, err)
	} else {
		_, err := s.Exec("CREATE TABLE IF NOT EXISTS dummy (id INTEGER PRIMARY KEY AUTOINCREMENT,secret text);")
		assert.Nil(t, err)
	}

	c, err := aes.NewCipher([]byte(encryptionKey))
	assert.Nil(t, err)

	gcm, err := cipher.NewGCM(c)
	assert.Nil(t, err)

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		assert.Nil(t, err)
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(superSecretValue), nil)
	quoted := strconv.Quote(string(encrypted))

	// we're assuming that there is a data encrypted with the original key
	_, err = s.ExecContext(context.Background(), "INSERT INTO `dummy` (`id`, `secret`) VALUES (1, ?)", quoted)
	assert.Nil(t, err)

	// read encrypted data with the original key
	fromDB := new(Dummy)
	err = meddler.Load(s, "dummy", fromDB, 1)
	assert.Nil(t, err)
	assert.Equal(t, superSecretValue, fromDB.Secret)

	// update data with the new key
	err = meddler.Update(s, "dummy", fromDB)
	assert.Nil(t, err)

	//try to read data with the original key, after re-encryption, expected an error
	fromDB = new(Dummy)
	err = meddler.Load(s, "dummy", fromDB, 1)
	assert.NotEqual(t, superSecretValue, fromDB.Secret)
	assert.NotNil(t, err)
}

func TestInitEncryption(t *testing.T) {
	encryptionKey := ""
	encryptionKeyNew := "new-key-has-to-be-32-bytes-long!"
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()

	if os.Getenv("DATABASE_DRIVER") != "" {
		_, err := s.Exec("CREATE TABLE IF NOT EXISTS dummy (id SERIAL, secret text);")
		assert.Nil(t, err)
	} else {
		_, err := s.Exec("CREATE TABLE IF NOT EXISTS dummy (id INTEGER PRIMARY KEY AUTOINCREMENT,secret text);")
		assert.Nil(t, err)
	}

	_, err := s.Exec("INSERT INTO dummy(id, secret) VALUES(1, 'superSecretValue');")
	assert.Nil(t, err)

	// read data without encryption
	fromDB := new(Dummy)
	err = meddler.Load(s, "dummy", fromDB, 1)
	assert.Nil(t, err)

	// update data with the new encryption key
	err = meddler.Update(s, "dummy", fromDB)
	assert.Nil(t, err)

	//try to read data, after re-encryption, expected an error
	fromDB = new(Dummy)
	err = meddler.Load(s, "dummy", fromDB, 1)
	assert.NotEqual(t, "superSecretValue", fromDB.Secret)
	assert.NotNil(t, err)
}

func TestUnquote(t *testing.T) {
	secretValue := "superSecretValue"
	encryptionKey := "the-key-has-to-be-32-bytes-long!"

	c, err := aes.NewCipher([]byte(encryptionKey))
	assert.Nil(t, err)

	gcm, err := cipher.NewGCM(c)
	assert.Nil(t, err)

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		assert.Nil(t, err)
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(secretValue), nil)
	quoted := strconv.Quote(string(encrypted))

	unquotedEncripted, _ := strconv.Unquote(quoted)
	assert.NotEqual(t, "", unquotedEncripted)

	//When you try unquote an un-encrypted string it returns an empyt string.
	unquotedRaw, _ := strconv.Unquote(secretValue)
	assert.Equal(t, "", unquotedRaw)
}
