import { ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/solid';
import React, { memo, useState, useCallback } from 'react';
import { GitRepositories, Kustomizations, HelmReleases, CompactServices, Summary } from './fluxState';

const Footer = memo(function Footer(props) {
  const { store, capacitorClient, targetReference, handleNavigationSelect } = props;
  const [fluxStates, setFluxStates] = useState(store.getState().fluxState);
  store.subscribe(() => setFluxStates(store.getState().fluxState))
  const [expanded, setExpanded] = useState(false)
  const [selected, setSelected] = useState("Kustomizations")
  const [tab, setTab] = useState("")

  const handleToggle = () => {
    setExpanded(!expanded)
  }

  const handlerSelect = useCallback((selectedNav) => {
    setSelected(selectedNav);
  },
    [setSelected]
  )

  if (!fluxStates) {
    return null
  }

  const fluxState = fluxStates[0].fluxState

  return (
    <div aria-labelledby="slide-over-title" role="dialog" aria-modal="true" className={`fixed inset-x-0 bottom-0 bg-neutral-200 z-40 border-t border-neutral-300 ${expanded ? 'h-4/5' : ''}`}>
      <div className={`flex justify-between w-full ${expanded ? '' : 'h-full'}`}>
        <div
          className='h-auto w-full cursor-pointer px-16 py-4 flex gap-x-12'
          onClick={handleToggle} >
          {!expanded &&
            <>
              <div className="grid grid-cols-3">
                {fluxStates.slice(0, 3).map(fluxState => {
                  if (!fluxState) {
                    return (
                      <div className="w-full truncate" key={fluxState.environment}>
                        <p className="font-semibold">
                          {fluxState.environment.toUpperCase()}
                          <span title="Disconnected">
                            <svg className="text-red-400 inline fill-current ml-1" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
                              <path
                                d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
                              />
                            </svg>
                          </span>
                        </p>
                      </div>
                    )
                  }

                  return (
                    <div className="w-full truncate" key={fluxState.environment}>
                      <p className="font-semibold text-neutral-700">{`${fluxState.environment.toUpperCase()}`}</p>
                      <div className="ml-2">
                        <Summary resources={fluxState.fluxState.gitRepositories} label="SOURCES" />
                        <Summary resources={fluxState.fluxState.kustomizations} label="KUSTOMIZATIONS" />
                        <Summary resources={fluxState.fluxState.helmReleases} label="HELM-RELEASES" />
                      </div>
                    </div>
                  )
                })}
              </div>
            </>
          }
        </div>
        <div className='px-4 py-2'>
          <button
            onClick={handleToggle}
            type="button" className="ml-1 rounded-md hover:bg-white hover:text-black text-neutral-700 p-1">
            <span className="sr-only">{expanded ? 'Close panel' : 'Open panel'}</span>
            {expanded ? <ArrowDownIcon className="h-5 w-5" aria-hidden="true" /> : <ArrowUpIcon className="h-5 w-5" aria-hidden="true" />}
          </button>
        </div>
      </div>
      {expanded &&
        <div>
          <nav className="flex space-x-8 px-6 pt-4" aria-label="Tabs">
            {fluxStates.map((fluxState) => (
              <span
                key={fluxState.environment}
                onClick={() => { setTab(fluxState.environment); return false }}
                className={classNames(
                  fluxState.environment === tab
                    ? 'border-indigo-500 text-gray-900'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                  'whitespace-nowrap border-b-2 pb-2 px-1 text-sm font-medium cursor-pointer'
                )}
                aria-current={fluxState.environment === tab ? 'page' : undefined}
              >
                {fluxState.environment.toUpperCase()}
              </span>
            ))}
          </nav>
          <div className="flex w-full h-full">
            <div>
              <div className="w-56 px-4 border-r border-neutral-300">
                <SideBar
                  navigation={[
                    { name: 'Sources', href: '#', count: fluxState.gitRepositories.length },
                    { name: 'Kustomizations', href: '#', count: fluxState.kustomizations.length },
                    { name: 'Helm Releases', href: '#', count: fluxState.helmReleases.length },
                    { name: 'Flux Runtime', href: '#', count: fluxState.fluxServices.length },
                  ]}
                  selectedMenu={handlerSelect}
                  selected={selected}
                />
              </div>
            </div>

            <div className="w-full px-4 overflow-x-hidden overflow-y-scroll">
              <div className="w-full max-w-7xl mx-auto flex-col h-full">
                <div className="pb-24 pt-2">
                  {selected === "Kustomizations" &&
                    <Kustomizations capacitorClient={capacitorClient} fluxState={fluxState} targetReference={targetReference} handleNavigationSelect={handleNavigationSelect} />
                  }
                  {selected === "Helm Releases" &&
                    <HelmReleases capacitorClient={capacitorClient} helmReleases={fluxState.helmReleases} targetReference={targetReference} handleNavigationSelect={handleNavigationSelect} />
                  }
                  {selected === "Sources" &&
                    <GitRepositories capacitorClient={capacitorClient} gitRepositories={fluxState.gitRepositories} targetReference={targetReference} />
                  }
                  {selected === "Flux Runtime" &&
                    <CompactServices capacitorClient={capacitorClient} store={store} services={fluxState.fluxServices} />
                  }
                </div>
              </div>
            </div>
          </div>
        </div>
      }
    </div>
  )
})

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}

function SideBar(props) {
  const { navigation, selectedMenu, selected } = props;

  return (
    <nav className="flex flex-1 flex-col" aria-label="Sidebar">
      <ul className="space-y-1">
        {navigation.map((item) => (
          <li key={item.name}>
            <a
              href={item.href}
              className={classNames(item.name === selected ? 'bg-white text-black' : 'text-neutral-700 hover:bg-white hover:text-black',
                'group flex gap-x-3 p-2 pl-3 text-sm leading-6 rounded-md')}
              onClick={() => selectedMenu(item.name)}
            >
              {item.name}
              {item.count ? (
                <span
                  className="ml-auto w-6 min-w-max whitespace-nowrap rounded-full bg-white px-2.5 py-0.5 text-center text-xs font-medium leading-5 text-neutral-700 ring-1 ring-inset ring-neutral-200"
                  aria-hidden="true"
                >
                  {item.count}
                </span>
              ) : null}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
};

export default Footer;
