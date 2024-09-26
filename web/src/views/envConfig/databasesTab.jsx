import { useEffect, useState } from 'react';
import { Combobox, ComboboxButton, ComboboxInput, ComboboxOption, ComboboxOptions, Label } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import {InfraComponent} from '../environment/category';
import {produce} from 'immer';

export function DatabasesTab(props) {
  const { gimletClient, store } = props;
  const { environment } = props;
  const { databaseConfig, setDatabaseValues } = props
  const { plainModules } = props;
  const [ selectedModule, setSelectedModule ] = useState()
  const [ dependencies, setDependencies ] = useState({
    "xxx": {
      url: "https://github.com/gimlet-io/plain-modules.git?path=postgresql",
      values: {}
    }
  })

  const validationCallback = (id, validationErrors) => {
    if(validationErrors) {
      console.log(id, validationErrors)
    }
  }

  const setDependencyValues = (id, values, nonDefaultValues) => {
    console.log(id, nonDefaultValues)

    setDependencies(produce(dependencies, draft => {
      draft[id].values = nonDefaultValues
    }))
  }

  useEffect(() => {
    console.log(dependencies)
  }, [dependencies]);

  useEffect(() => {
    console.log(selectedModule)
  }, [selectedModule]);
  
  return (
    <div className="">
      <ModuleSelector modules={plainModules} setSelectedModule={setSelectedModule} />
      {Object.keys(dependencies).map((id) => {
        const dependency = dependencies[id]
        const module = plainModules.find(m => m.url == dependency.url)

        return <InfraComponent
            key={id}
            componentDefinition={module}
            config={dependency.values}
            setValues={(variable, values, nonDefaultValues) => setDependencyValues(id, values, nonDefaultValues)}
            validationCallback={(variable, validationErrors) => validationCallback(id, validationErrors)}
            gimletClient={gimletClient}
            store={store}
            environment={{name: environment}}
          />
        })
      }
    </div>
  )
}

export default function ModuleSelector(props) {
  const { setSelectedModule } = props
  const parsedModules = props.modules.map((m) => {
    return {
      ...m,
      schema: JSON.parse(m.schema),
      uiSchema: JSON.parse(m.uiSchema),
    }
  })
  const [query, setQuery] = useState('')
  const [selected, setSelected] = useState(parsedModules[0].schema.title)

  const filteredModules =
    query === ''
      ? parsedModules
      : parsedModules.filter((module) => {
          return module.schema.title.toLowerCase().includes(query.toLowerCase())
        })

  useEffect(() => {
    setSelectedModule(parsedModules.find(m => m.schema.title === selected))
  }, [selected]);

  return (
    <Combobox
      as="div"
      value={selected}
      onChange={(moduleTitle) => {
        setQuery('')
        setSelected(moduleTitle)
      }}
    >
      <div className="relative mt-2">
        <ComboboxInput
          className="w-full rounded-md border-0 bg-white py-1.5 pl-3 pr-10 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
          onChange={(event) => setQuery(event.target.value)}
          onBlur={() => setQuery('')}
          displayValue={selected}
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
