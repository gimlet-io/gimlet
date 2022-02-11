package seal

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bitnami-labs/sealed-secrets/pkg/crypto"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/mdaverde/jsonpath"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/cert"
)

var Command = cli.Command{
	Name:      "seal",
	Usage:     "Seals secrets in the manifest",
	UsageText: `gimlet seal -f values.yaml -o values.yaml -p sealedSecrets -k sealingKey.crt`,
	Action:    seal,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "manifest file,folder or \"-\" for stdin (mandatory)",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:     "path",
			Aliases:  []string{"p"},
			Usage:    "path(s) of the field(s) to seal (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "key",
			Aliases:  []string{"k"},
			Usage:    "path to the sealing key (mandatory)",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "output sealed file",
		},
	},
}

func seal(c *cli.Context) error {
	file := c.String("file")
	jsonPaths := c.StringSlice("path")
	outputPath := c.String("output")
	keyPath := c.String("key")

	key, err := loadKey(keyPath)
	if err != nil {
		return fmt.Errorf("could not load sealing key: %s", err)
	}

	files, err := commands.InputFiles(file)
	if err != nil {
		return err
	}
	for path, contents := range files {
		err := sealFile(path, contents, jsonPaths, outputPath, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func sealFile(path string, contents string, jsonPaths []string, outputPath string, key *rsa.PublicKey) error {
	var parsed map[string]interface{}
	err := yaml.Unmarshal([]byte(contents), &parsed)
	if err != nil {
		return fmt.Errorf("could not parse %s: %s", path, err)
	}

	for _, jsonPath := range jsonPaths {
		secrets, err := jsonpath.Get(parsed, jsonPath)
		if err != nil {
			return fmt.Errorf("could not lookup %s: %s", jsonPath, err)
		}
		sealedSecrets := map[string]string{}
		switch secretMap := secrets.(type) {
		case map[string]interface{}:
			for k, v := range secretMap {
				vString := v.(string)
				sealed, err := sealed(vString)
				if err != nil {
					return fmt.Errorf("could not check %s state: %s", jsonPath, err)
				}

				if sealed {
					sealedSecrets[k] = vString
				} else {
					sealedValue, err := sealValue(key, vString)
					if err != nil {
						return fmt.Errorf("could not seal %s: %s", v, err)
					}
					sealedSecrets[k] = sealedValue
				}
			}
		default:
			return fmt.Errorf("%s is not a map of secrets", jsonPath)
		}

		err = jsonpath.Set(parsed, jsonPath, sealedSecrets)
		if err != nil {
			return fmt.Errorf("could not set sealed secrets: %s", err)
		}
	}

	yamlString := bytes.NewBufferString("")
	e := yaml.NewEncoder(yamlString)
	e.SetIndent(2)
	err = e.Encode(parsed)
	if err != nil {
		return fmt.Errorf("could not marshal yaml: %s", err)
	}
	if outputPath != "" {
		fi, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("cannot stat %s: %s", outputPath, err)
		}
		if fi.IsDir() {
			outputPath = outputPath + filepath.Base(path)
		}

		err = ioutil.WriteFile(outputPath, yamlString.Bytes(), commands.File_RW_RW_R)
		if err != nil {
			return fmt.Errorf("cannot write output file %s", err)
		}
	} else {
		fmt.Println(yamlString)
	}

	return nil
}

/*
From https://github.com/bitnami-labs/sealed-secrets/blob/f903596e6561bd3151e9b2d12591472e886f24da/pkg/crypto/crypto.go#L86
And https://github.com/bitnami-labs/sealed-secrets/blob/master/docs/crypto.md

If it doesn't follow the sealed secret format constraints, it is not a sealed secret.
It may still follow it, and not be a sealed secret. We don't cover that marginal case - famous last words
*/
func sealed(value string) (bool, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		if _, ok := err.(base64.CorruptInputError); ok {
			return false, nil
		}
		return false, err
	}
	if len(ciphertext) < 2 {
		return false, nil
	}
	rsaLen := int(binary.BigEndian.Uint16(ciphertext))

	if rsaLen != 512 { // edge case detection. SealedSecrets works with 4096 bit long (512 bytes) long keys
		return false, nil
	}

	if len(ciphertext) < rsaLen+2 {
		return false, nil
	}

	//rsaCiphertext := ciphertext[2 : rsaLen+2]
	//aesCiphertext := ciphertext[rsaLen+2:]
	//fmt.Println(rsaCiphertext)
	//fmt.Println(aesCiphertext)

	return true, nil
}

// From https:/	/github.com/bitnami-labs/sealed-secrets/blob/f903596e6561bd3151e9b2d12591472e886f24da/cmd/kubeseal/main.go#L678
func sealValue(pubKey *rsa.PublicKey, data string) (string, error) {
	clusterWide := []byte("")
	result, err := crypto.HybridEncrypt(rand.Reader, pubKey, []byte(data), clusterWide)
	return base64.StdEncoding.EncodeToString(result), err
}

func loadKey(keyPath string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	certs, err := cert.ParseCertsPEM(data)
	if err != nil {
		return nil, err
	}

	cert, ok := certs[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected RSA public key but found %v", certs[0].PublicKey)
	}

	return cert, nil
}
