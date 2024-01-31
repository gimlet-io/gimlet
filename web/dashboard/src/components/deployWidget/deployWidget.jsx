import {Menu, Transition} from '@headlessui/react'
import {ChevronDownIcon} from '@heroicons/react/solid'
// eslint-disable-next-line import/no-webpack-loader-syntax
import logo from "!file-loader!./logo.svg";
import { usePostHog } from 'posthog-js/react'
import { useState, Fragment } from 'react';
import { FilterIcon } from '@heroicons/react/solid'

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
          <Menu.Button
            className="inline-flex items-center gap-x-1.5 bg-white px-3.5 py-2.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 bg-slate-800 
                       relative cursor-pointer px-4 py-2 rounded-l-md"
          >
            <img
              className="h-5 w-auto" src={logo} alt="Deploy"/>
            <span
              className="bg-gradient-to-r from-orange-400 from-0% via-pink-400 via-40% to-pink-500 to-90% text-transparent bg-clip-text">
              Deploy
            </span>
          </Menu.Button>
        <span className="-ml-px relative block">
          <Menu.Button
            className="relative z-0 inline-flex items-center px-2 py-3 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-pink-500 bg-slate-800">
            <span className="sr-only">Open options</span>
            <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
          </Menu.Button>
          <Transition as={Fragment} afterLeave={() => setFilter('')}>
          <Menu.Items
              className="origin-top-right absolute z-50 right-0 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-slate-800 text-white ring-1 ring-black ring-opacity-5 focus:outline-none">
              <div className="py-1">
              {deployTargets.length > 10 &&
                <div className="relative -mt-1">
                  <input
                    className="block border-0 rounded-t-md border-t border-b pl-8 border-gray-300 w-full pt-1.5 pb-1 text-gray-900 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    placeholder="Enter Filter"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    type="search"
                  />
                  <div className="absolute left-0 inset-y-0 flex items-center">
                    <FilterIcon className="ml-2 h-5 w-5 text-gray-400" aria-hidden="true" />
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
                            active ? 'bg-slate-600 text-slate-100' : 'text-slate-100') +
                            ' block px-4 py-2 text-sm w-full text-left'
                          }
                        >
                          Deploy all to {env}
                        </button>
                      )}
                    </Menu.Item>
                  )
                } else {
                  return null;
                }
              })}
              {filteredTargets.length === 0 ?
                <div className="select-none text-slate-100 opacity-50 block px-4 py-2 text-sm w-full text-left">
                  No matches found.
                </div>
                :
                filteredTargets.map((target) => (
                  <Menu.Item key={`${target.app}-${target.env}`}>
                    {({ active }) => (
                      <button
                        onClick={() => deployHandler(target, sha, repo)}
                        className={(
                          active ? 'bg-slate-600 text-slate-100' : 'text-slate-100') +
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
