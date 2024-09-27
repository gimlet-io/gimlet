import { useEffect, useState } from 'react';
import { Combobox, ComboboxButton, ComboboxInput, ComboboxOption, ComboboxOptions, Label } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import {InfraComponent} from '../environment/category';
import {produce} from 'immer';
import { v4 as uuidv4 } from 'uuid';

export function DatabasesTab(props) {
  const { gimletClient, store } = props;
  const { environment, plainModules } = props;
  const parsedModules = plainModules.map((m) => {
    return {
      ...m,
      schema: JSON.parse(m.schema),
      uiSchema: JSON.parse(m.uiSchema),
    }
  })
  const { configFileDependencies, setConfigFileDependencies } = props;
  const [ selectedModule, setSelectedModule ] = useState()
  const [ dependencies, setDependencies ] = useState(fromConfigFileDependencies(configFileDependencies, parsedModules))

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

  const addDependency = () => {
    setDependencies(produce(dependencies, draft => {
      draft[uuidv4()] = {
        url: selectedModule.url,
        title: selectedModule.schema.title,
        values: {}
      }
    }))
  }

  const deleteDependency = (id) => {
    setDependencies(produce(dependencies, draft => {
      delete draft[id]
    }))
  }

  useEffect(() => {
    console.log(dependencies)
    const rebuiltDependencies = []
    for(const dependency of Object.values(dependencies)) {
      rebuiltDependencies.push({
        name: dependency.title.toLowerCase(),
        kind: "plain",
        spec: {
          module: {
            url: dependency.url
          },
          values: dependency.values
        }
      })
    }
    setConfigFileDependencies(rebuiltDependencies)
  }, [dependencies]);

  return (
    <div className='space-y-12'>
      <div className='flex space-x-2'>
        <div className='flex-grow'>
          <ModuleSelector parsedModules={parsedModules} setSelectedModule={setSelectedModule} />
        </div>
        <button onClick={addDependency} className="primaryButton px-8">Add</button>
      </div>
      {Object.keys(dependencies).map((id) => {
        const dependency = dependencies[id]
        const module = plainModules.find(m => m.url == dependency.url)

        return (
          <div key={id} className='relative'>
            <button onClick={() => deleteDependency(id)} className="destructiveButtonSecondary absolute top-6 right-6">Delete</button>
            <InfraComponent
              componentDefinition={module}
              config={dependency.values}
              setValues={(variable, values, nonDefaultValues) => setDependencyValues(id, values, nonDefaultValues)}
              validationCallback={(variable, validationErrors) => validationCallback(id, validationErrors)}
              gimletClient={gimletClient}
              store={store}
              environment={{name: environment}}
            />
          </div>
        )})
      }
    </div>
  )
}

export default function ModuleSelector(props) {
  const { setSelectedModule, parsedModules } = props
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
      <div className="relative">
        <ComboboxInput
          className="input"
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

function fromConfigFileDependencies(configfileDependencies, parsedModules){
  if (!configfileDependencies) {
    return {}
  }

  const dependencies = {}
  for (const dependency of configfileDependencies) {
    dependencies[uuidv4()] = {
      url: dependency.spec.module.url,
      title: parsedModules.find((m) => m.url === dependency.spec.module.url).schema.title,
      values: dependency.spec.values,
    }
  }

  return dependencies
}
