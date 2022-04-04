package gitops

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"

	"github.com/caarlos0/sshmarshal"
	"golang.org/x/crypto/ssh"
)

func GenerateEd25519() ([]byte, []byte, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	block, err := sshmarshal.MarshalPrivateKey(priv, "keygen@gimlet.io")
	if err != nil {
		return nil, nil, err
	}

	b, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(b)

	return pem.EncodeToMemory(block), publicKeyBytes, nil
}
