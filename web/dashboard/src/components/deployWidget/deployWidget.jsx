import {Menu} from '@headlessui/react'
import {ChevronDownIcon} from '@heroicons/react/solid'
// eslint-disable-next-line import/no-webpack-loader-syntax
import logo from "!file-loader!./logo.svg";
import { usePostHog } from 'posthog-js/react'
import { commits } from '../../redux/eventHandlers/eventHandlers';

export default function DeployWidget(props) {
  const {deployTargets, deployHandler, magicDeployHandler, sha, repo, envs, envConfigs } = props;
  const posthog = usePostHog()

  if (!deployTargets) { // magic deploy cases
    if(!envConfigs) {
      return null
    }

    if(envs.length === 0) {
      return null
    }

    let targets = [];
    for (const env of envs) {
      if (!env.isOnline) {
        continue
      }
      const configs = envConfigs[env.name];
      if (!configs) {
        continue
      }
      for (const config of configs) { // Adding env configs
        const exists = targets.find(t => t.app === config.app && t.env === env.name);
        if (exists) {
          continue
        }

        const repository = config.values.image?.repository
        const tag = config.values.image?.tag
        const hasVariable = repository?.includes("{{") || tag?.includes("{{")
        const pointsToBuiltInRegistry = repository?.includes("127.0.0.1:32447")

        if (hasVariable && !pointsToBuiltInRegistry) {
          // this is the dynamic image tag case when CI sends an artifact

          this only regards the configs at the latest commit
          what we should do instead is that for static + static-site + buildpacks commits we should
          create an artifact right when we learn about the commit (analyize configs + create artifacts),
          then magic deploy and regular deploy paths move closer

          continue
        }

        targets.push({
          env: env.name,
          app: config.app,
        })
      }
    }

    if (targets.length === 0) {
      return null
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
                {
                  targets.map((target) => (
                      <Menu.Item key={`${target.app}-${target.env}`}>
                      {({ active }) => (
                        <button
                          onClick={() => {
                            posthog?.capture('Magic deploy button pushed')
                            magicDeployHandler(target.env, target.app, repo, sha)
                          }}
                          className={(
                            active ? 'bg-slate-600 text-slate-100' : 'text-slate-100') +
                            ' block px-4 py-2 text-sm w-full text-left'
                          }
                        >
                          {target.app} to {target.env}
                        </button>
                      )}
                      </Menu.Item>
                  ))
                }
              </div>
            </Menu.Items>
          </span>
        </Menu>
      </span>
    )
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
          className="relative cursor-pointer inline-flex items-center px-4 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          Deploy..
        </Menu.Button>
        <span className="-ml-px relative block">
          <Menu.Button
            className="relative z-0 inline-flex items-center px-2 py-3 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50">
            <span className="sr-only">Open options</span>
            <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
          </Menu.Button>
          <Menu.Items
            className="origin-top-right absolute z-50 right-0 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
            <div className="py-1">
              {Object.keys(deployTargetsByEnv).map((env) => {
                if (deployTargetsByEnv[env].length > 1) {
                  return (
                    <Menu.Item key={`${env}`}>
                      {({ active }) => (
                        <button
                          onClick={() => deployHandler({
                            env: env,
                            app: "",
                            artifactId: deployTargetsByEnv[env][0].artifactId
                          }, sha, repo)}
                          className={(
                            active ? 'bg-yellow-100 text-gray-900' : 'bg-yellow-50 text-gray-700') +
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
                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700') +
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