package seal

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/commands"
	"github.com/mdaverde/jsonpath"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

func Test_seal(t *testing.T) {
	sealingKeyPath, err := ioutil.TempFile("", "gimlet-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	_ = ioutil.WriteFile(sealingKeyPath.Name(), []byte(sealingKey), commands.File_RW_RW_R)
	defer os.Remove(sealingKeyPath.Name())

	args := strings.Split("gimlet seal", " ")
	args = append(args, "-f", "-")
	args = append(args, "-p", "sealedSecrets")
	args = append(args, "-p", "sealedSecrets2")
	args = append(args, "-k", sealingKeyPath.Name())

	g := goblin.Goblin(t)

	g.Describe("gimlet chart seal", func() {
		g.It("Should seal stdin", func() {
			const toSeal = `
key: value
another: one

sealedSecrets:
  secret1: value1
  secret2: value2
sealedSecrets2:
  secret3: value3
`

			oldStdin, err := feedStringToStdin(toSeal)
			g.Assert(err == nil).IsTrue(err)
			defer func() { os.Stdin = oldStdin }()

			old := os.Stdout // keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				outC <- buf.String()
			}()
			w.Close()
			os.Stdout = old
			sealedValue := <-outC
			g.Assert(strings.Contains(sealedValue, "secret1")).IsTrue(sealedValue)
			g.Assert(strings.Contains(sealedValue, "secret2")).IsTrue(sealedValue)
			g.Assert(strings.Contains(sealedValue, "secret3")).IsTrue(sealedValue)
			fmt.Println(sealedValue)
		})
		g.It("Should seal then unseal", func() {
			certs, err := cert.ParseCertsPEM([]byte(sealingKey))
			g.Assert(err == nil).IsTrue(err)

			cert, ok := certs[0].PublicKey.(*rsa.PublicKey)
			g.Assert(ok).IsTrue()
			val, err := sealValue(cert, "value1")
			g.Assert(err == nil).IsTrue(err)

			ss := &v1alpha1.SealedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sealedsecrets.bitnami.com/cluster-wide": "true",
					},
				},
				Spec: v1alpha1.SealedSecretSpec{
					EncryptedData: map[string]string{
						"secret1": val,
					},
				},
			}

			key, err := keyutil.ParsePrivateKeyPEM([]byte(privateKey))
			g.Assert(err == nil).IsTrue(err)

			secret, err := ss.Unseal(scheme.Codecs, map[string]*rsa.PrivateKey{"": key.(*rsa.PrivateKey)})
			g.Assert(err == nil).IsTrue(err)
			g.Assert(string(secret.Data["secret1"])).Equal("value1")
		})
	})

}

func feedStringToStdin(toSeal string) (*os.File, error) {
	tmpFile, err := ioutil.TempFile("", "dummyStdIn")
	if err != nil {
		return os.Stdin, err
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(toSeal))
	if err != nil {
		return os.Stdin, err
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return os.Stdin, err
	}

	oldStdin := os.Stdin
	os.Stdin = tmpFile

	return oldStdin, err
}

func Test_sealed(t *testing.T) {
	result, err := sealed("notBase64")
	if err != nil {
		t.Error(err)
	}
	if result {
		t.Errorf("Should be not sealed")
	}

	result, err = sealed("bm90U2VhbGVkCg==")
	if err != nil {
		t.Error(err)
	}
	if result {
		t.Errorf("Should be not sealed, just base64 encoded string")
	}

	result, err = sealed("AgBRuxQIic7NUct/U8Ih/pTV/qiUTv3OeHxSjqF+bHmtoZ08V4zcqDYYgypxE0GfUzP7mou7Hhto2pIpB6/ctwf6Fb328DiIRew9C/QojSXftryONZJTqsbLF21MASPS54u7vT+jG/KIaasnOvxMk2ZLeaKC8gGuy9wtk5IRtL6bP91H3vDTNKx+3GFVR4qrRmP7ZSBWAsgui8DH1S1h9v4P6CUAxy0mPCDLl2DZi50kr6kpj3d9N1wpQ2gDTMGRJ3K/d9kDdTWaS1dt8mQ6HLvdaRPYQr0JdI2m+KEMC1G4ApoAhoQj1NabECL5ZB4sKrB5Z1FxqswJCl48x8jP0SwMiB2tQuqEjfAtiYXXOFwQqQ+n2xqK23s3mnT3AYhukb8ztUrecKkIxA4Hah8x9mlsV7iEPHLJAETm5D8Gvs1/6IkT2cNcmY0yZnQwKSQfUmr+ClwQFGvDcK+J07s9yuW5Lt/+wLKErMG9c8n6uMQqpV0/H66nPVWEMcPlpITDHDXyuK4f8TWms+cgIonHBBtg7zppc30/QXzadhADuokdByNbgYGxkxsYkexpY+okquxuS5fEf0bE8sXnJP3zgFw78Px9Lc5HXZCkfa/si62K2L+JM0NdAgEdWUJAGm8leXC154ad6C6JMulDD+QlNPX8dCtk8YBBCbRJuoKxW6pnlfwV3FmYFRYAAzWOl01Fkw4x3D4g/xihLRiC3Y883s6PrGgFxyZ2vklbCKb18hcxsY/Xbhbp/g==")
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Errorf("Should be sealed")
	}

	result, err = sealed(edgeTestCase())
	if err != nil {
		t.Error(err)
	}
	if result {
		t.Error("Should be caught by edge case detection")
	}
}

func Test_jsonpath(t *testing.T) {
	var parsed map[string]interface{}
	err := yaml.Unmarshal([]byte(data), &parsed)
	if err != nil {
		t.Error(err)
	}

	picked, err := jsonpath.Get(parsed, "envs.staging[0].sealedSecrets")
	if err != nil {
		t.Error(err)
	}

	err = jsonpath.Set(parsed, "envs.staging[0].sealedSecrets", picked)
	if err != nil {
		t.Error(err)
	}
}

const data = `
envs:
  staging:
    - name: my-app
      sealedSecrets:
        secret1: value1
        secret2: value2
  production:
    - name: my-app
      sealedSecrets:
        secret1: prod1
        secret2: prod2
`

// If a raw secret (without base64 encoding) starts with a char sequence that is a valid BigEndian encoded 16bit int
// and the raw secret is long enough, then we could falsely identify it as a valid sealed secret.
func edgeTestCase() string {
	endian := make([]byte, 2)
	binary.BigEndian.PutUint16(endian, 16)

	longText := "This is a long text, longer than 16 bytes"
	longText = string(endian) + longText

	return base64.StdEncoding.EncodeToString([]byte(longText))
}

// openssl req -new -newkey rsa:4096 -nodes -x509 -keyout test.key -out test.cert
const sealingKey = `-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIUGX54jxohTCLp01b+zqYCNg+nkDgwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMDExMzAxMzU5NTlaFw0yMDEy
MzAxMzU5NTlaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDlBF9Y+HovrlaESdEQ86h340Z6IDMge4XLan2O0L8i
6IqHEnASQCVpkxinH0xyGxZrO3iKFgDdv0CjFj7a++mniCZjCIU0SlBRXjUvn3ht
G1Jmttwmu2adQWXSsMFEnqkKjzRs04yowuj/ry15r42yTumwheDTOl9hCqV/KksS
zC7oyxWZRdUsKc3y8tUhJdEQzWa0RZ7Ksghm76K6G5pqGEDApX/VS3J8xDnMdSnR
4NxgyQxc/c4UfzZsRix//4Q4Y2lICdcPpcvAkPfCP/0HlcARLlhux/eNHnz6h4OI
eCEp4OF1LfhS9uEv5SKFlVz3Z5B0to3fKVV4nHazi9WZzMglUTrAr8Xvc82K2UnZ
LXoVEDc+e/YXn166h+HlWk2+l+b/aZdBkZndodZCWJ+UtrLuOQJD3t3OgUNpE6hu
kromVG3uJiJQ2reB40p/RL+fP2FyBTY7mJFsIJ6BQtOBBb6TzoNPtuqjfYm149xu
e3uOrNCVttdHTdkYrev0dNPya+daOUXHv2+dMgh8Sb7C6fDYakurhRYTZEokPZjK
cN03n5hV8ioLcXGOJQW2zaYK9pIbtCQnxqvVfKH5lQpO0MgqXbAw8jUAy5w8E/Ct
vslPQt3Q9vCE6koqGdJCAw3wbUNRnLo8y++If8yGKl2rwA69f7S2kuOHBDvfNjzK
AQIDAQABo1MwUTAdBgNVHQ4EFgQUbF3qbWTnEAFVxRN9qS3KrJWrpwcwHwYDVR0j
BBgwFoAUbF3qbWTnEAFVxRN9qS3KrJWrpwcwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEAmk2BxClKbTzCrP7P+j6nF3v4gHho6ebUQXPJc2hn76tY
XaNmyQ0YgSWcytWpagvys+MX58L4iuK4GXfn428VbvUzMlrnqhezI/YTGQ4LUqSp
l8e6rw+tapXRN8yMEWb7gQBhfq93iRKYacbSO/yr9jfVR7RckgQmeQnnGKpZtQ5G
ZW5CzPuFGIfRe94eWporPrP6p6TTR0j3+6ts6ZAXXa1dP7+oVUAm702bl0Mc0lW1
90GfcT1NzSSzW9nZWcITgRpgv4+MBQLcmLPMaDwvF15vM70CVBzVHBC3XfQNnm7w
VtRcK7RU02pTAiYJB3p3HHphLfW2Hf8NLoNF1DjopCJPZ1LCB9QPFcLyqy3n02UK
o7/92TwZYL4Ll4a0CVlQXwqPW8TRZuC8cStw5hgpyHyyv+15juxybZVpMZe1NXpJ
0AEkU+6jfjVETiah5Kbkp48ZEcJ/YgpK8FswW3wVBV9pHhB9uaVP4ICJ2zGyVkNd
yhQNcE+JBXmiWIYDlVrSYCeZB/D0tBKEd8QYK8VP824RGvaw7bNPBS+jCqMTxO24
9dH7u+AwmLOcDG/yguDlNzugLef3TNJW5jes12wrdvY23LfuR79Oc0h9wJdKwJjp
oLHbDLJf+c+RkFlIRXGifA7gPM0y4EAl5ysgKVgbSImKjn3NClpX/rCpdGwuJwE=
-----END CERTIFICATE-----`

const privateKey = `-----BEGIN PRIVATE KEY-----
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
hGs6DoX5eUU/AHBRIVySTePugHD1vA==
-----END PRIVATE KEY-----`
