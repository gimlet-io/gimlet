
import React, { memo, useState } from 'react';
import { GitRepositories, Kustomizations, HelmReleases, CompactServices } from './fluxState';

const GitopsStatus = memo(function GitopsStatus({fluxStates, handleNavigationSelect, selectedTab, gimletClient, store, targetReference}) {
  const [selectedEnv, setSelectedEnv] = useState(Object.keys(fluxStates)[0])
  const fluxState = fluxStates[selectedEnv];

  if (!fluxState) {
    return null
  }

  return (
    <>
      <nav className="flex space-x-8 px-6 pt-4" aria-label="Tabs">
        {Object.keys(fluxStates).map((env) => (
          <span
            key={env}
            onClick={() => { setSelectedEnv(env); return false }}
            className={classNames(
              env === selectedEnv
                ? 'border-indigo-500 text-gray-900'
                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
              'whitespace-nowrap border-b-2 pb-2 px-1 text-sm font-medium cursor-pointer'
            )}
            aria-current={env === selectedEnv ? 'page' : undefined}
          >
            {env.toUpperCase()}
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
              selectedMenu={handleNavigationSelect}
              selected={selectedTab}
            />
          </div>
        </div>

        <div className="w-full px-4 overflow-x-hidden overflow-y-scroll">
          <div className="w-full max-w-7xl mx-auto flex-col h-full">
            <div className="pb-24 pt-2">
              {selectedTab === "Kustomizations" &&
                <Kustomizations gimletClient={gimletClient} fluxState={fluxState} handleNavigationSelect={handleNavigationSelect} />
              }
              {selectedTab === "Helm Releases" &&
                <HelmReleases gimletClient={gimletClient} helmReleases={fluxState.helmReleases} />
              }
              {selectedTab === "Sources" &&
                <GitRepositories gimletClient={gimletClient} gitRepositories={fluxState.gitRepositories} targetReference={targetReference} />
              }
              {selectedTab === "Flux Runtime" &&
                <CompactServices gimletClient={gimletClient} store={store} services={fluxState.fluxServices} />
              }
            </div>
          </div>
        </div>
      </div>
    </>
  )
})

export default GitopsStatus;

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