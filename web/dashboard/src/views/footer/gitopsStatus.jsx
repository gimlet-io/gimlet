
import React, { memo, useState } from 'react';
import { ExpandedFooter } from "./ExpandedFooter"

const GitopsStatus = memo(function GitopsStatus({fluxStates, fluxEvents, handleNavigationSelect, selectedTab, gimletClient, store, targetReference}) {
  const [selectedEnv, setSelectedEnv] = useState(Object.keys(fluxStates)[0])
  const fluxState = fluxStates[selectedEnv];
  const fluxK8sEvents = fluxEvents[selectedEnv];

  if (!fluxState) {
    return null
  }

  let sources = []
  if (fluxState.ociRepositories) {
    sources.push(...fluxState.ociRepositories)
    sources.push(...fluxState.gitRepositories)
    sources.push(...fluxState.buckets)
  }
  sources = [...sources].sort((a, b) => a.metadata.name.localeCompare(b.metadata.name));

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
      <div className="w-full h-full overscroll-contain">
        <ExpandedFooter
          client={gimletClient}
          handleNavigationSelect={handleNavigationSelect}
          targetReference={targetReference}
          fluxState={fluxState}
          fluxEvents={fluxK8sEvents}
          sources={sources}
          selected={selectedTab}
          store={store}
        />
      </div>
    </>
  )
})

export default GitopsStatus;

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}
