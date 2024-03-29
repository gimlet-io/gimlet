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
	"testing"
)

var (
	encryptionKey    = "the-key-has-to-be-32-bytes-long!"
	encryptionKeyNew = ""
)

func TestStoreInit(t *testing.T) {
	s := NewTest(encryptionKey, encryptionKeyNew)
	defer func() {
		s.Close()
	}()
}
