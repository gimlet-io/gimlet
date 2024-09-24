import {InfraComponent} from '../environment/category';
import { useState, useEffect } from 'react';
import {produce} from 'immer';
import { Combobox, ComboboxButton, ComboboxInput, ComboboxOption, ComboboxOptions, Label } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'

export function DatabasesTab(props) {
  const { gimletClient, store } = props;
  const { environment } = props;
  const { databaseConfig, setDatabaseValues } = props
  const { plainModules } = props;

  const validationCallback = (variable, validationErrors) => {
    if(validationErrors) {
      console.log(variable, validationErrors)
    }
  }

  console.log(databaseConfig)

  return (
    <div className="w-full space-y-8">
      <Example/>
      {/* {databases.map(component =>
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
      )} */}
    </div>
  )
}


const people = [
  { id: 1, name: 'Leslie Alexander' },
]


export default function Example() {
  const [query, setQuery] = useState('')
  const [selectedPerson, setSelectedPerson] = useState(null)

  const filteredPeople =
    query === ''
      ? people
      : people.filter((person) => {
          return person.name.toLowerCase().includes(query.toLowerCase())
        })

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

        {filteredPeople.length > 0 && (
          <ComboboxOptions className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
            {filteredPeople.map((person) => (
              <ComboboxOption
                key={person.id}
                value={person}
                className="group relative cursor-default select-none py-2 pl-3 pr-9 text-gray-900 data-[focus]:bg-indigo-600 data-[focus]:text-white"
              >
                <span className="block truncate group-data-[selected]:font-semibold">{person.name}</span>

                <span className="absolute inset-y-0 right-0 hidden items-center pr-4 text-indigo-600 group-data-[selected]:flex group-data-[focus]:text-white">
                  <CheckIcon className="h-5 w-5" aria-hidden="true" />
                </span>
              </ComboboxOption>
            ))}
          </ComboboxOptions>
        )}
      </div>
    </Combobox>
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
    ]
  },
  {
    variable: "postgresql",
    schema: {
      "$schema": "http://json-schema.org/draft-07/schema",
      "$id": "#postgresql",
      "type": "object",
      "title": "PostgreSQL",
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
    ]
  }
]
