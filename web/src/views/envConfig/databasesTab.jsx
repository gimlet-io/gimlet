import { useEffect, useState } from 'react';
import HelmUI from "helm-react-ui";
import { EncryptedWidget } from '../environment/encryptedWidget';
import { Combobox, ComboboxButton, ComboboxInput, ComboboxOption, ComboboxOptions, Label } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import {InfraComponent} from '../environment/category';

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

  useEffect(() => {
    console.log(values)
  }, [values]);
  
  return (
    <div className="">
      <ModuleSelector modules={plainModules} />
      {/* <HelmUI
        schema={schema2}
        config={uiSchema2}
        fields={customFields}
        values={values}
        setValues={setValues}
        validate={true}
        validationCallback={(errors) => validationCallback("", errors)}
      /> */}
      <InfraComponent
        key={plainModules[0].schema.$id}
        componentDefinition={plainModules[0]}
        config={{}}
        setValues={setDatabaseValues}
        validationCallback={validationCallback}
        gimletClient={gimletClient}
        store={store}
        environment={{name: environment}}
      />
    </div>
  )
}

export default function ModuleSelector(props) {
  const parsedModules = props.modules.map((m) => {
    return {
      ...m,
      schema: JSON.parse(m.schema),
      uiSchema: JSON.parse(m.uiSchema),
    }
  })
  const [query, setQuery] = useState('')
  const [selectedPerson, setSelectedPerson] = useState(null)

  const filteredModules =
    query === ''
      ? parsedModules
      : parsedModules.filter((module) => {
          return module.schema.title.toLowerCase().includes(query.toLowerCase())
        })

  console.log(parsedModules)

  return (
    <Combobox
      as="div"
      value={selectedPerson}
      onChange={(person) => {
        setQuery('')
        setSelectedPerson(person)
      }}
    >
      <Label className="block text-sm font-medium leading-6 text-gray-900">Assigned to</Label>
      <div className="relative mt-2">
        <ComboboxInput
          className="w-full rounded-md border-0 bg-white py-1.5 pl-3 pr-10 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
          onChange={(event) => setQuery(event.target.value)}
          onBlur={() => setQuery('')}
          displayValue={(person) => person?.name}
        />
        <ComboboxButton className="absolute inset-y-0 right-0 flex items-center rounded-r-md px-2 focus:outline-none">
          <ChevronUpDownIcon className="h-5 w-5 text-gray-400" aria-hidden="true" />
        </ComboboxButton>

        {filteredModules.length > 0 && (
          <ComboboxOptions className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
            {filteredModules.map((module) => (
              <ComboboxOption
                key={module.schema.id}
                value={module.schema.title}
                className="group relative cursor-default select-none py-2 pl-3 pr-9 text-gray-900 data-[focus]:bg-indigo-600 data-[focus]:text-white"
              >
                <div>
                  <span className="block truncate group-data-[selected]:font-semibold">{module.schema.title}</span>
                  <span className="absolute inset-y-0 right-0 hidden items-center pr-4 text-indigo-600 group-data-[selected]:flex group-data-[focus]:text-white">
                    <CheckIcon className="h-5 w-5" aria-hidden="true" />
                  </span>
                </div>
                <div>
                  <span className="block truncate group-data-[selected]:font-semibold text-xs">{module.schema.description}</span>
                </div>
              </ComboboxOption>
            ))}
          </ComboboxOptions>
        )}
      </div>
    </Combobox>
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