import {Menu} from '@headlessui/react'
import {ChevronDownIcon} from '@heroicons/react/solid'
// eslint-disable-next-line import/no-webpack-loader-syntax
import logo from "!file-loader!./logo.svg";
import { usePostHog } from 'posthog-js/react'
import { commits } from '../../redux/eventHandlers/eventHandlers';

export default function DeployWidget(props) {
  const {deployTargets, deployHandler, magicDeployHandler, sha, repo, envs, envConfigs } = props;
  const posthog = usePostHog()

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
          <Menu.Items
              className="origin-top-right absolute z-50 right-0 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-slate-800 text-white ring-1 ring-black ring-opacity-5 focus:outline-none">
              <div className="py-1">
              {Object.keys(deployTargetsByEnv).map((env) => {
                if (deployTargetsByEnv[env].length > 1) {
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
              {deployTargets.map((target) => (
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
        </span>
      </Menu>
    </span>
  )
}