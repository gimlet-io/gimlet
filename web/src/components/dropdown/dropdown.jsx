import { Fragment, useEffect, useState } from 'react'
import { Listbox, Transition } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/24/solid'

export default function Dropdown(props) {
  const { items, value, changeHandler, icon, buttonClass, onCard } = props;
  const [selected, setSelected] = useState(value)

  useEffect(() => {
    changeHandler(selected)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected]);

  if (!items) {
    return null;
  }

  let colors = "bg-white dark:bg-neutral-800 text-neutral-900 dark:text-neutral-200"
  if (onCard) {
    colors = "bg-white dark:bg-neutral-900 text-neutral-900 dark:text-neutral-300"
  }

  return (
    <Listbox value={selected} onChange={setSelected}>
      {({ open }) => (
        <>
          <div className="relative">
            <Listbox.Button
              className={`relative w-full cursor-default rounded-md py-2 pl-3 pr-10 text-left ${colors} ring-1 ring-inset ring-neutral-200 dark:ring-neutral-700 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6`}>
              <div className="flex items-center">
                <span>{icon}</span>
                <span className={`block truncate ${buttonClass}`}>{value}</span>
              </div>
              <span className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <ChevronUpDownIcon className="h-5 w-5 text-neutral-400" aria-hidden="true" />
              </span>
            </Listbox.Button>

            <Transition
              show={open}
              as={Fragment}
              leave="transition ease-in duration-100"
              leaveFrom="opacity-100"
              leaveTo="opacity-0"
            >
              <Listbox.Options
                static
                className={`absolute z-10 mt-1 w-full ${colors} shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black dark:ring-neutral-700 ring-opacity-5 overflow-auto focus:outline-none sm:text-sm`}
              >
                {items.map((item) => (
                  <Listbox.Option
                    key={item}
                    className={({ active }) =>
                      (active ? 'text-white bg-indigo-600' : 'text-neutral-900 dark:text-neutral-200') +
                      ' cursor-default select-none relative py-2 pl-3 pr-9'
                    }
                    value={item}
                  >
                    {({ selected, active }) => (
                      <>
                        <span className={(selected ? 'font-semibold' : 'font-normal') + ' block truncate'}>
                          {item}
                        </span>

                        {selected ? (
                          <span
                            className={(active ? 'text-white' : 'text-indigo-600') +
                              ' absolute inset-y-0 right-0 flex items-center pr-4'
                            }
                          >
                            <CheckIcon className="h-5 w-5" aria-hidden="true" />
                          </span>
                        ) : null}
                      </>
                    )}
                  </Listbox.Option>
                ))}
              </Listbox.Options>
            </Transition>
          </div>
        </>
      )}
    </Listbox>
  )
}