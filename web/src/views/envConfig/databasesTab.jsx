import { useState } from 'react';
import HelmUI from "helm-react-ui";
import { EncryptedWidget } from '../environment/encryptedWidget';

export function DatabasesTab(props) {
  const { gimletClient, store } = props;
  const { environment } = props;
  const { databaseConfig, setDatabaseValues } = props
  const { plainModules } = props;
  const [ values, setValues ] = useState({})

  const validationCallback = (variable, validationErrors) => {
    if(validationErrors) {
      console.log(variable, validationErrors)
    }
  }

  const customFields = {
    "encryptedSingleLineWidget": (props) => <EncryptedWidget
      {...props}
      gimletClient={gimletClient}
      store={store}
      env={environment}
      singleLine={true}
    />,
  }
  
  console.log(databaseConfig)

  return (
    <div className="w-full card p-6 pb-8">
      {/* <ModuleSelector modules={plainModules} /> */}
      <HelmUI
        schema={schema2}
        config={uiSchema2}
        fields={customFields}
        values={values}
        setValues={setValues}
        validate={true}
        validationCallback={(errors) => validationCallback("", errors)}
      />
    </div>
  )
}

const schema2 = {
  "$schema": "http://json-schema.org/draft-07/schema",
  "$id": "#/properties/dependencies",
  "type": "array",
  "title": "Dependencies",
  "default": [],
  "additionalItems": true,
  "items": {
    "$id": "#/properties/dependencies/items",
    "type": "object",
    "anyOf": [
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "$id": "#redis",
        "type": "object",
        "title": "Redis",
        "description": "Install a Redis instance dedicated for your application",
        "properties": {
          "encryptedPassword": {
            "$id": "#/properties/encryptedPassword",
            "type": "string",
            "title": "Password"
          }
        }
      },
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "$id": "#postgresql",
        "title": "PostgreSQL",
        "description": "A containerized PostgreSQL dedicated for your application, without backups",
        "type": "object",
        "properties": {
          "name": {
            "$id": "#/properties/name",
            "type": "string",
            "title": "Database Name"
          },
          "user": {
            "$id": "#/properties/user",
            "type": "string",
            "title": "Username"
          },
          "encryptedPassword": {
            "$id": "#/properties/encryptedPassword",
            "type": "string",
            "title": "Password"
          }
        }
      },
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "$id": "#rabbitmq",
        "title": "RabbitMQ",
        "description": "A containerized RabbitMQ dedicated for your application",
        "type": "object",
        "properties": {
          "user": {
            "$id": "#/properties/user",
            "type": "string",
            "title": "Username"
          },
          "encryptedPassword": {
            "$id": "#/properties/encryptedPassword",
            "type": "string",
            "title": "Password"
          }
        }
      }
    ]
  }
}

const uiSchema2 = [
  {
    "schemaIDs": [
      "#/properties/dependencies"
    ],
    "uiSchema": {
      "#/properties/dependencies": {
        "items": {
          "encryptedPassword": {
            "ui:field": "encryptedSingleLineWidget"
          },
          "encryptedPassword": {
            "ui:field": "encryptedSingleLineWidget"
          },
          "encryptedPassword": {
            "ui:field": "encryptedSingleLineWidget"
          }
        }
      }
    },
    "metaData": {
    }
  }
]