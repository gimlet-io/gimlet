import {Menu, Transition} from '@headlessui/react'
import {ChevronDownIcon} from '@heroicons/react/24/solid'
import { usePostHog } from 'posthog-js/react'
import { useState, Fragment } from 'react';
import { FunnelIcon } from '@heroicons/react/24/solid'

export default function DeployWidget(props) {
  const {deployTargets, deployHandler, sha, repo } = props;
  const posthog = usePostHog()
  const [filter, setFilter] = useState("")

  if (!deployTargets) {
    return null;
  }

  let deployTargetsByEnv = {};
  for (let target of deployTargets) {
    if (!deployTargetsByEnv[target.env]) {
      deployTargetsByEnv[target.env] = [];
    }

    deployTargetsByEnv[target.env].push(target);
  }

  const filteredTargets = filterTargets(deployTargets, filter)

  return (
    <span className="relative inline-flex flex-row rounded-md">
      <Menu as="span" className="relative inline-flex shadow-sm rounded-md align-middle">
        <div>
          <Menu.Button className="relative primaryButton pl-4 pr-8">
            Deploy..
            <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
              <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
            </span>
          </Menu.Button>
        </div>
        <span className="-ml-px relative block">
         
          <Transition as={Fragment} afterLeave={() => setFilter('')}>
          <Menu.Items
              className="absolute right-0 top-10 z-10 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-white dark:bg-neutral-800 ring-1 ring-black dark:ring-neutral-700 ring-opacity-5 focus:outline-none">
              <div className="py-1">
              {deployTargets.length > 10 &&
                <div className="relative -mt-1">
                  <input
                    className="block border-0 rounded-t-md border-t border-b pl-8 border-black border-opacity-5 w-full pt-1.5 pb-1 text-neutral-700 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    placeholder="Enter Filter"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    type="search"
                  />
                  <div className="absolute left-0 inset-y-0 flex items-center">
                    <FunnelIcon className="ml-2 h-5 w-5 text-neutral-400" aria-hidden="true" />
                  </div>
                </div>
              }
              {Object.keys(deployTargetsByEnv).map((env) => {
                if (deployTargetsByEnv[env].length > 1 && filter === "") {
                  return (
                    <Menu.Item key={`${env}`}>
                      {({ active }) => (
                        <button
                          onClick={() => {
                              posthog?.capture('Deploy button pushed')
                              deployHandler({
                                env: env,
                                app: "",
                                artifactId: deployTargetsByEnv[env][0].artifactId
                              }, sha, repo)
                            }
                          }
                          className={(
                            active ? 'text-white bg-indigo-600' : 'text-neutral-900 dark:text-neutral-200') +
                            ' block px-4 py-2 text-sm w-full text-left'
                          }
                        >
                          All to {env}
                        </button>
                      )}
                    </Menu.Item>
                  )
                } else {
                  return null;
                }
              })}
              {filteredTargets.length === 0 ?
                <div className="select-none text-neutral-900 dark:text-neutral-200 opacity-50 block px-4 py-2 text-sm w-full text-left">
                  No matches found.
                </div>
                :
                filteredTargets.map((target) => (
                  <Menu.Item key={`${target.app}-${target.env}`}>
                    {({ active }) => (
                      <button
                        onClick={() => deployHandler(target, sha, repo)}
                        className={(
                          active ? 'text-white bg-indigo-600' : 'text-neutral-900 dark:text-neutral-200') +
                          ' block px-4 py-2 text-sm w-full text-left'
                        }
                      >
                         {target.app} to {target.env}
                      </button>
                    )}
                  </Menu.Item>
                ))}
            </div>
          </Menu.Items>
          </Transition>
        </span>
      </Menu>
    </span>
  )
}

function filterTargets(targets, filter) {
  return [...targets].filter(target => target.app.includes(filter) || target.env.includes(filter))
}
