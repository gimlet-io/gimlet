package manifest

import (
	"testing"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/stretchr/testify/assert"
)

func Test_tfVariablesParsing_ErrorParsingConfig(t *testing.T) {
	rawInvalid := `invalid config`

	variables := []*tfconfig.Variable{}
	err := ParseVariables([]byte(rawInvalid), &variables)
	assert.Error(t, err)
}

func Test_tfVariablesParsing_ValidConfig(t *testing.T) {
	rawVariables := `
variable "user" {
  type = string
}

variable "admin_username" {
  type = string
  sensitive = true
}

variable "image_id" {
  type = string
  description = "The id of the machine image (AMI) to use for the server."
}
`

	v := []*tfconfig.Variable{}
	ParseVariables([]byte(rawVariables), &v)

	assert.Equal(t, 3, len(v))

	assert.Equal(t, "user", v[0].Name)
	assert.Equal(t, false, v[0].Sensitive)
	assert.Equal(t, "string", v[0].Type)

	assert.Equal(t, "admin_username", v[1].Name)
	assert.Equal(t, true, v[1].Sensitive)
	assert.Equal(t, "string", v[1].Type)

	assert.Equal(t, "The id of the machine image (AMI) to use for the server.", v[2].Description)
}

func Test_tfVariablesParsing_DefaultField(t *testing.T) {
	raw := `
variable "availability_zone_names" {
  type    = list(string)
  default = ["us-west-1a"]
}
`
	v := []*tfconfig.Variable{}
	ParseVariables([]byte(raw), &v)

	assert.Equal(t, []interface{}{"us-west-1a"}, v[0].Default)
}
