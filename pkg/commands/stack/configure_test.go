package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func manualTest_Configure(t *testing.T) {
	stackConfigYaml := `
stack:
  repository: "..."
config:
  nginx:
    enabled: true
`

	stackDefinitionYaml := `
name: "Gimlet Stack Reference"
description: "Providing a reference stack to demonstrate Gimlet Stack"
categories:
- name: "⬅️Ingress"
  id: "ingress"
- name: "📑 Logging"
  id: "logging"
components:
- name: "Nginx"
  variable: "nginx"
  category: "ingress"
  logo: "https://raw.githubusercontent.com/gimlet-io/gimlet-stack-reference/main/assets/nginx.png"
  description: ""
  onePager: "An Nginx proxy server that routes traffic to your applications based on the host name or path."
  schema: |-
    {
      "$schema": "http://json-schema.org/draft-07/schema",
      "$id": "http://example.com/example.json",
      "type": "object",
      "title": "The root schema",
      "description": "The root schema comprises the entire JSON document.",
      "properties": {
        "enabled": {
          "$id": "#/properties/enabled",
          "type": "boolean",
          "title": "Enabled"
        },
        "host": {
          "$id": "#/properties/host",
          "type": "string",
          "title": "Host",
          "description": "Your company domain you will expose your services on",
          "default": "",
          "examples": [
            "mycompany.com"
          ]
        }
      }
    }
  uiSchema: |-
    [
      {
        "schemaIDs": [
          "#/properties/enabled",
          "#/properties/host"
        ],
        "uiSchema": {},
        "metaData": {}
      }
    ]
`

	var stackConfig StackConfig
	err := yaml.Unmarshal([]byte(stackConfigYaml), &stackConfig)
	assert.Nil(t, err)

	var stackDefinition StackDefinition
	err = yaml.Unmarshal([]byte(stackDefinitionYaml), &stackDefinition)
	assert.Nil(t, err)

	_, _, err = Configure(stackDefinition, stackConfig)
}
