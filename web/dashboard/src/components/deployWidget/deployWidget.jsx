import {Menu} from '@headlessui/react'
import {ChevronDownIcon} from '@heroicons/react/solid'
// eslint-disable-next-line import/no-webpack-loader-syntax
import logo from "!file-loader!./logo.svg";

export default function DeployWidget(props) {
  const {deployTargets, deployHandler, sha, repo, hasBuiltInEnv } = props;

  if (!deployTargets) {
    if (!hasBuiltInEnv) {
      return null
    }
    return (
      // eslint-disable-next-line
      <button
        type="button"
        className="inline-flex items-center gap-x-1.5 rounded-md bg-white px-3.5 py-2.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 bg-slate-800"
      >
        <img
          className="h-5 w-auto" src={logo} alt="Deploy"/>
        <span className="bg-gradient-to-r from-orange-400 from-0% via-pink-400 via-40% to-pink-500 to-90% text-transparent bg-clip-text">
          Deploy
        </span>
      </button>
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