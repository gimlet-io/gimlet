import {InfraComponent} from '../environment/category';
import { useState, useEffect } from 'react';
import {produce} from 'immer';

export function DatabasesTab(props) {
  const { gimletClient, store } = props;
  const { environment } = props;
  const { databaseConfig, setDatabaseValues } = props

  const validationCallback = (variable, validationErrors) => {
    if(validationErrors) {
      console.log(variable, validationErrors)
    }
  }

  console.log(environment)

  return (
    <div className="w-full space-y-8">
      {databases.map(component =>
        <InfraComponent
          key={component.schema.$id}
          componentDefinition={component}
          config={databaseConfig[component.variable] ?? {}}
          setValues={setDatabaseValues}
          validationCallback={validationCallback}
          gimletClient={gimletClient}
          store={store}
          environment={{name: environment}}
        />
      )}
    </div>
  )
}

const databases = [
  {
    variable: "redis",
    schema: {
      "$schema": "http://json-schema.org/draft-07/schema",
      "$id": "#redis",
      "type": "object",
      "title": "Redis",
      "description": "Install a Redis instance dedicated for your application",
      "properties": {
        "enabled": {
          "$id": "#/properties/enabled",
          "type": "boolean",
          "title": "Enabled"
        },
        "encryptedPassword": {
          "$id": "#/properties/encryptedPassword",
          "type": "string",
          "title": "Password"
        }
      }
    },
    uiSchema: [
      {
        "schemaIDs": [
          "#redis"
        ],
        "uiSchema": {
          "#redis": {
            "encryptedPassword": {
              "ui:field": "encryptedSingleLineWidget"
            }
          }
        },
        "metaData": {
          "link": {
            "label": "Redis",
            "href": "https://gimlet.io/docs/databases/redis"
          }
        }
      }
    ],
  },
  {
    variable: "postgresql",
    schema: {
      "$schema": "http://json-schema.org/draft-07/schema",
      "$id": "#postgresql",
      "type": "object",
      "title": "PostgreSQL database",
      "description": "Provision a logical database with a user and password in the centralized PostgreSQL instance",
      "properties": {
        "enabled": {
          "$id": "#/properties/enabled",
          "type": "boolean",
          "title": "Enabled"
        },
        "database": {
          "$id": "#/properties/database",
          "type": "string",
          "title": "Database"
        },
        "user": {
          "$id": "#/properties/user",
          "type": "string",
          "title": "User"
        },
        "encryptedPassword": {
          "$id": "#/properties/encryptedPassword",
          "type": "string",
          "title": "Password"
        }
      }
    },
    uiSchema: [
      {
        "schemaIDs": [
          "#postgresql"
        ],
        "uiSchema": {
          "#postgresql": {
            "encryptedPassword": {
              "ui:field": "encryptedSingleLineWidget"
            }
          }
        },
        "metaData": {
          "link": {
            "label": "Shared PostgreSQL database",
            "href": "https://gimlet.io/docs/databases/postgresql"
          }
        }
      }
    ],
  }
]
