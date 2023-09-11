import { Menu } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/solid'

const TenantSelector = ({ tenants, selectedTenant, setSelectedTenant }) => {
  if (tenants.length === 0) {
    return null;
  }

  return (
    <Menu as="span" className="relative inline-flex shadow-sm rounded-md align-middle mt-8 sm:mt-4">
      <Menu.Button
        className="relative cursor-pointer inline-flex items-center px-4 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-700"
      >
        <span className="max-w-xs truncate">
          {selectedTenant === "" ? "All tenants" : selectedTenant}
        </span>
      </Menu.Button>
      <span className="-ml-px relative block">
        <Menu.Button
          className="relative z-0 inline-flex items-center px-2 py-3 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500">
          <span className="sr-only">Open options</span>
          <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
        </Menu.Button>
        <Menu.Items
          className="origin-top-right absolute z-50 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
          <div className="py-1">
            <Menu.Item key="all">
              {({ active }) => (
                <button
                  onClick={() => setSelectedTenant("")}
                  className={(
                    active ? 'bg-yellow-100 text-gray-900' : 'bg-yellow-50 text-gray-700') +
                    ' block px-4 py-2 text-sm w-full text-left'
                  }
                >
                  All tenants
                </button>
              )}
            </Menu.Item>
            {tenants.map((tenant) => (
              <Menu.Item key={tenant}>
                {({ active }) => (
                  <button
                    onClick={() => setSelectedTenant(tenant)}
                    className={(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700') +
                      ' block px-4 py-2 text-sm w-full text-left'
                    }
                  >
                    {tenant}
                  </button>
                )}
              </Menu.Item>
            ))}
          </div>
        </Menu.Items>
      </span>
    </Menu>)
};

export default TenantSelector;
