package server

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

func saveInfrastructureComponents(w http.ResponseWriter, r *http.Request) {
	var infrastructureComponents map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&infrastructureComponents)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	infrastructureComponentsString, err := json.Marshal(infrastructureComponents)
	if err != nil {
		logrus.Errorf("cannot serialize infrastructure components: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(infrastructureComponentsString)
}

func stackDefinition(w http.ResponseWriter, r *http.Request) {
	stackDefinitionString := `
	{
		"name": "Gimlet Stack Reference",
		"description": "Providing a reference stack to demonstrate Gimlet Stack",
		"categories": [
		  {
			"name": "⬅️Ingress",
			"id": "ingress"
		  },
		  {
			"name": "\uD83D\uDCD1 Logging",
			"id": "logging"
		  }
		],
		"components": [
		  {
			"name": "Nginx",
			"variable": "nginx",
			"category": "ingress",
			"logo": "https://raw.githubusercontent.com/gimlet-io/gimlet-stack-reference/main/assets/nginx.png",
			"description": "",
			"onePager": "An Nginx proxy server that routes traffic to your applications based on the host name or path.",
			"schema": {
			  "$schema": "http://json-schema.org/draft-07/schema",
			  "$id": "#nginx",
			  "type": "object",
			  "properties": {
				"enabled": {
				  "$id": "#/properties/enabled",
				  "type": "boolean",
				  "title": "Enabled"
				}
			  },
			  "dependencies": {
				"enabled": {
				  "oneOf": [
					{
					  "properties": {
						"enabled": {
						  "const": false
						}
					  }
					},
					{
					  "properties": {
						"enabled": {
						  "const": true
						},
						"host": {
						  "$id": "#/properties/host",
						  "type": "string",
						  "title": "Host",
						  "description": "Your company domain you will expose your services on"
						}
					  },
					  "required": [
						"host"
					  ]
					}
				  ]
				}
			  }
			},
			"uiSchema": [
			  {
				"schemaIDs": [
				  "#nginx"
				],
				"uiSchema": {},
				"metaData": {}
			  }
			]
		  },
		  {
			"name": "CertManager",
			"variable": "certManager",
			"category": "ingress",
			"logo": "https://raw.githubusercontent.com/gimlet-io/gimlet-stack-reference/main/assets/certManager.png",
			"description": "",
			"onePager": "",
			"schema": {
			  "$schema": "http://json-schema.org/draft-07/schema",
			  "$id": "http://example.com/example.json",
			  "type": "object",
			  "title": "The root schema",
			  "description": "The root schema comprises the entire JSON document.",
			  "required": [
				"email"
			  ],
			  "properties": {
				"enabled": {
				  "$id": "#/properties/enabled",
				  "type": "boolean",
				  "title": "Enabled"
				},
				"email": {
				  "$id": "#/properties/email",
				  "type": "string",
				  "title": "Administrator email",
				  "description": "Let's Encrypt will email you on this email upon expiring certificates",
				  "default": "",
				  "examples": [
					"admin@mycompany.com"
				  ]
				}
			  }
			},
			"uiSchema": [
			  {
				"schemaIDs": [
				  "#/properties/enabled",
				  "#/properties/email"
				],
				"uiSchema": {},
				"metaData": {}
			  }
			]
		  },
		  {
			"name": "Loki",
			"variable": "loki",
			"category": "logging",
			"logo": "https://raw.githubusercontent.com/gimlet-io/gimlet-stack-reference/main/assets/loki.png",
			"description": "",
			"onePager": "",
			"schema": {
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
				}
			  }
			},
			"uiSchema": [
			  {
				"schemaIDs": [
				  "#/properties/enabled"
				],
				"uiSchema": {},
				"metaData": {}
			  }
			]
		  }
		]
	  }
	  `

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(stackDefinitionString))
}

func stack(w http.ResponseWriter, r *http.Request) {
	stackString := `
	{
		"certManager": {
    "email": "laszlo@gimlet.io",
    "enabled": "true"
		},
  "loki": {
    "enabled": "true",
    "persistence": "true"
  },
  "nginx": {
    "enabled": "true",
    "host": "staging.gimlet.io"
  },
  "prometheus": {
    "enabled": "true",
    "persistence": "true"
  },
  "sealedSecrets": {
    "enabled": "true"
  }
	}
	`

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(stackString))
}
