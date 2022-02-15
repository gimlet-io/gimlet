import {Menu} from '@headlessui/react'
import {ChevronDownIcon} from '@heroicons/react/solid'

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}

export default function DeployWidget(props) {
  const {deployTargets, deployHandler, sha, repo} = props;

  if (!deployTargets) {
    return (
      // eslint-disable-next-line
      <a href="https://gimlet.io/gimletd/on-demand-releases/" target="_blank"
         class="text-xs text-gray-400 cursor-pointer">
        Want to deploy this version?
      </a>
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
    <span className="relative inline-flex shadow-sm rounded-md">
      <button
        type="button"
        className="relative cursor-default inline-flex items-center px-4 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-700 hover:bg-gray-50"
      >
        Deploy..
      </button>
      <Menu as="span" className="-ml-px relative block">
        <Menu.Button
          className="relative z-0 inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500">
          <span className="sr-only">Open options</span>
          <ChevronDownIcon className="h-5 w-5" aria-hidden="true"/>
        </Menu.Button>
          <Menu.Items
            className="origin-top-right absolute z-50 right-0 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
            <div className="py-1">
              {Object.keys(deployTargetsByEnv).map((env) => {
                if (deployTargetsByEnv[env].length > 1) {
                  return (
                    <Menu.Item key={`${env}`}>
                      {({active}) => (
                        <button
                          onClick={() => deployHandler({
                            env: env,
                            app: "",
                            artifactId: deployTargetsByEnv[env][0].artifactId
                          }, sha, repo)}
                          className={classNames(
                            active ? 'bg-yellow-100 text-gray-900' : 'bg-yellow-50 text-gray-700',
                            'block px-4 py-2 text-sm w-full text-left'
                          )}
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
                  {({active}) => (
                    <button
                      onClick={() => deployHandler(target, sha, repo)}
                      className={classNames(
                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                        'block px-4 py-2 text-sm w-full text-left'
                      )}
                    >
                      {target.app} to {target.env}
                    </button>
                  )}
                </Menu.Item>
              ))}
            </div>
          </Menu.Items>
      </Menu>
    </span>
  )
}