import React, { Fragment } from 'react';
import { Menu, Transition } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/20/solid'

export default function MenuButton(props) {
  const { children, items, handleClick } = props;

  if (!items) {
    return null
  }

  return (
    <Menu as="div" className="relative inline-block text-left">
      <div>
        <Menu.Button className="relative primaryButton pl-4 pr-8">
          {children}
          <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
            <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
          </span>
        </Menu.Button>
      </div>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="absolute right-0 z-10 mt-2 w-full origin-top-right rounded-md bg-white dark:bg-neutral-800 shadow-lg ring-1 ring-black dark:ring-neutral-700 ring-opacity-5 focus:outline-none">
          <div className="py-1">
            {items.map(item => {
              return (
                <Menu.Item key={item.name}>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(item.name)}
                    className={(
                      active ? 'text-white bg-indigo-600' : 'text-neutral-900 dark:text-neutral-200') +
                      ' block px-4 py-2 text-sm w-full text-left'
                    }
                  >
                    to <span className='capitalize'>{item.name}</span>
                  </button>
                )}
              </Menu.Item>
              )
            })}
          </div>
        </Menu.Items>
      </Transition>
    </Menu>
  )
}
