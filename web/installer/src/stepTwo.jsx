import { useEffect, useState } from 'react';
import { InformationCircleIcon } from '@heroicons/react/solid';
import { Switch } from '@headlessui/react'
import { SeparateEnvironments } from 'shared-components';

const StepTwo = ({ appId }) => {
  const [env, setEnv] = useState('production');
  const [repoPerEnv, setRepoPerEnv] = useState(false);
  const [useExistingPostgres, setUseExistingPostgres] = useState(false);
  const [hostAndPort, setHostAndPort] = useState('postgresql:5432');
  const [dashboardDb, setDashboardDb] = useState('gimlet_dashboard');
  const [dashboardUsername, setDashboardUsername] = useState('gimlet_dashboard');
  const [dashboardPassword, setDashboardPassword] = useState('');
  const [gimletdDb, setGimletdDb] = useState('gimletd');
  const [gimletdUsername, setGimletdUsername] = useState('gimletd');
  const [gimletdPassword, setGimletPassword] = useState('');
  const [infra, setInfra] = useState('gitops-infra');
  const [apps, setApps] = useState('gitops-apps');

  useEffect(() => {
    if (repoPerEnv) {
      setInfra(`gitops-${env}-infra`);
      setApps(`gitops-${env}-apps`);
    } else {
      setInfra(`gitops-infra`);
      setApps(`gitops-apps`);
    }
  }, [repoPerEnv, env]);

  return (
    <div class="mt-32 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="max-w-4xl mx-auto">
        <div class="md:flex md:items-center md:justify-between">
          <div class="flex-1 min-w-0">
            <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate mb-12">Gimlet Installer</h2>
          </div>
        </div>
        <nav aria-label="Progress">
          <ol class="border border-gray-300 rounded-md divide-y divide-gray-300 md:flex md:divide-y-0">
            <li class="relative md:flex-1 md:flex">
              {/* <!-- Completed Step --> */}
              <div class="group flex items-center w-full select-none cursor-default">
                <span class="px-6 py-4 flex items-center text-sm font-medium">
                  <span class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-indigo-600 rounded-full">
                    {/* <!-- Heroicon name: solid/check --> */}
                    <svg class="w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                      fill="currentColor" aria-hidden="true">
                      <path fill-rule="evenodd"
                        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                        clip-rule="evenodd" />
                    </svg>
                  </span>
                  <span class="ml-4 text-sm font-medium text-gray-900">Create Github Application</span>
                </span>
              </div>

              {/* <!-- Arrow separator for lg screens and up --> */}
              <div class="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                <svg class="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                  <path d="M0 -2L20 40L0 82" vector-effect="non-scaling-stroke" stroke="currentcolor"
                    stroke-linejoin="round" />
                </svg>
              </div>
            </li>

            <li class="relative md:flex-1 md:flex">
              {/* <!-- Current Step --> */}
              <div class="px-6 py-4 flex items-center text-sm font-medium select-none cursor-default" aria-current="step">
                <span
                  class="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-indigo-600 rounded-full">
                  <span class="text-indigo-600">02</span>
                </span>
                <span class="ml-4 text-sm font-medium text-indigo-600">Prepare gitops repository</span>
              </div>

              {/* <!-- Arrow separator for lg screens and up --> */}
              <div class="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                <svg class="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                  <path d="M0 -2L20 40L0 82" vector-effect="non-scaling-stroke" stroke="currentcolor"
                    stroke-linejoin="round" />
                </svg>
              </div>
            </li>

            <li class="relative md:flex-1 md:flex">
              {/* <!-- Upcoming Step --> */}
              <div class="group flex items-center select-none cursor-default">
                <span class="px-6 py-4 flex items-center text-sm font-medium">
                  <span
                    class="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-gray-300 rounded-full">
                    <span class="text-gray-500 ">03</span>
                  </span>
                  <span class="ml-4 text-sm font-medium text-gray-500 ">Bootstrap gitops automation</span>
                </span>
              </div>
            </li>
          </ol>
        </nav>

        {appId === "" &&
          <div class="rounded-md bg-red-50 p-4 my-8">
            <div class="flex">
              <div class="flex-shrink-0">
                {/* <!-- Heroicon name: solid/x-circle --> */}
                <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                  aria-hidden="true">
                  <path fill-rule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clip-rule="evenodd" />
                </svg>
              </div>
              <div class="ml-3">
                <h3 class="text-sm font-medium text-red-800">A Github Application was not yet created by the installer. Go
                  to <a href="/" class="font-bold">Step One</a> to create it.</h3>
              </div>
            </div>
          </div>
        }
        <form action="/bootstrap" method="post">
          <div class="mt-8 text-sm">
            <div class="mt-4 rounded-md bg-blue-50 p-4">
              <div class="flex">
                <div class="flex-shrink-0">
                  <InformationCircleIcon class="h-5 w-5 text-blue-400" aria-hidden="true" />
                </div>
                <div class="ml-3 md:justify-between">
                  <p class="text-sm text-blue-500">
                    By default, infrastructure manifests of this environment will be placed in the <span
                      class="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded">{env}</span>
                    folder of the shared <span
                      class="text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">gitops-infra</span>
                    git repository,
                    and application manifests will be placed in the <span
                      class="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded">{env}</span>
                    folder of the shared <span
                      class="text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">gitops-apps</span>
                    git repository
                  </p>
                </div>
              </div>
            </div>

            <div class="text-gray-700">
              <div class="flex mt-4">
                <div class="font-medium self-center">Environment name</div>
                <div class="max-w-lg flex rounded-md ml-4">
                  <div class="max-w-lg w-full lg:max-w-xs">
                    <input id="apps" name="env"
                      value={env}
                      onChange={e => setEnv(e.target.value)}
                      class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                      type="text" />
                  </div>
                </div>
              </div>
              <div class="text-sm text-gray-500 leading-loose"></div>
            </div>
            <SeparateEnvironments
              repoPerEnv={repoPerEnv}
              setRepoPerEnv={setRepoPerEnv}
              infraRepo={infra}
              appsRepo={apps}
            />
            <div class="flex mt-4">
              <div class="font-medium self-center">Use my existing Postgresql database</div>
              <div class="max-w-lg flex rounded-md ml-4">
                <Switch
                  checked={useExistingPostgres}
                  onChange={setUseExistingPostgres}
                  className={(
                    useExistingPostgres ? "bg-indigo-600" : "bg-gray-200") +
                    " relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200"
                  }
                >
                  <span className="sr-only">Use setting</span>
                  <span
                    aria-hidden="true"
                    className={(
                      useExistingPostgres ? "translate-x-5" : "translate-x-0") +
                      " pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                    }
                  />
                </Switch>
                <input id="useExistingPostgres" name="useExistingPostgres"
                  value={useExistingPostgres}
                  onChange={e => setUseExistingPostgres(e.target.value)}
                  type="hidden" />
              </div>
            </div>
            <div class="text-sm text-gray-500 leading-loose">By default, a Postgresql database will be installed to store the Gimlet data
            </div>

            {useExistingPostgres &&
              <div class="ml-8">
                <div class="flex mt-4">
                  <div class="font-medium self-center">Host and port</div>
                  <div class="max-w-lg flex rounded-md ml-4">
                    <div class="max-w-lg w-full lg:max-w-xs">
                      <input id="hostAndPort" name="hostAndPort"
                        value={hostAndPort}
                        onChange={e => setHostAndPort(e.target.value)}
                        class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                        type="text" />
                    </div>
                  </div>
                </div>
                <div class="flex mt-4">
                  <div class="font-medium self-center">Dashboard Database</div>
                  <div class="max-w-lg flex rounded-md ml-4">
                    <div class="max-w-lg w-full lg:max-w-xs">
                      <input id="dashboardDb" name="dashboardDb"
                        value={dashboardDb}
                        onChange={e => setDashboardDb(e.target.value)}
                        class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                        type="text" />
                    </div>
                  </div>
                </div>
                <div class="flex mt-4">
                  <div class="font-medium self-center">Dashboard Username</div>
                  <div class="max-w-lg flex rounded-md ml-4">
                    <div class="max-w-lg w-full lg:max-w-xs">
                      <input id="dashboardUsername" name="dashboardUsername"
                        value={dashboardUsername}
                        onChange={e => setDashboardUsername(e.target.value)}
                        class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                        type="text" />
                    </div>
                  </div>
                </div>
                <div class="flex mt-4">
                  <div class="font-medium self-center">Dashboard Password</div>
                  <div class="max-w-lg flex rounded-md ml-4">
                    <div class="max-w-lg w-full lg:max-w-xs">
                      <input id="dashboardPassword" name="dashboardPassword"
                        value={dashboardPassword}
                        onChange={e => setDashboardPassword(e.target.value)}
                        class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                        type="text" />
                    </div>
                  </div>
                </div>
                <div class="flex mt-4">
                  <div class="font-medium self-center">GimletD Database</div>
                  <div class="max-w-lg flex rounded-md ml-4">
                    <div class="max-w-lg w-full lg:max-w-xs">
                      <input id="gimletdDb" name="gimletdDb"
                        value={gimletdDb}
                        onChange={e => setGimletdDb(e.target.value)}
                        class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                        type="text" />
                    </div>
                  </div>
                </div>
                <div class="flex mt-4">
                  <div class="font-medium self-center">GimletD Username</div>
                  <div class="max-w-lg flex rounded-md ml-4">
                    <div class="max-w-lg w-full lg:max-w-xs">
                      <input id="gimletdUsername" name="gimletdUsername"
                        value={gimletdUsername}
                        onChange={e => setGimletdUsername(e.target.value)}
                        class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                        type="text" />
                    </div>
                  </div>
                </div>
                <div class="flex mt-4">
                  <div class="font-medium self-center">GimletD Password</div>
                  <div class="max-w-lg flex rounded-md ml-4">
                    <div class="max-w-lg w-full lg:max-w-xs">
                      <input id="gimletdPassword" name="gimletdPassword"
                        value={gimletdPassword}
                        onChange={e => setGimletPassword(e.target.value)}
                        class="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                        type="text" />
                    </div>
                  </div>
                </div>
              </div>
            }

            <div class="p-0 flow-root my-8">
              <span class="inline-flex rounded-md shadow-sm gap-x-3 float-right">
                <button
                  class="bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700 inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150">
                  Prepare gitops repository
                </button>
              </span>
            </div>
          </div>
        </form>
      </div>
    </div>

  );
};

export default StepTwo;
