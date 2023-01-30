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
	"fmt"
	"os"
	"testing"

	"github.com/russross/meddler"
	"github.com/stretchr/testify/assert"
)

func TestStoreInit(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()
}

type Dummy struct {
	ID     int64  `json:"-" meddler:"id,pk"`
	Secret string `json:"-" meddler:"secret,encrypted"`
}

func TestEncryption(t *testing.T) {
	s := NewTest()
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
