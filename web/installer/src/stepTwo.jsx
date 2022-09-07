import { useEffect, useState } from 'react';
import { StackUI, SeparateEnvironments } from 'shared-components';

const StepTwo = ({ getContext }) => {
  const [context, setContext] = useState(null);
  const [env, setEnv] = useState('test');
  const [repoPerEnv, setRepoPerEnv] = useState(true);
  const [infra, setInfra] = useState('gitops-test-infra');
  const [apps, setApps] = useState('gitops-test-apps');

  useEffect(() => {
    getContext().then(data => {
      let environment = data.stackConfig.config.gimletd.environments[0]
      environment.name = env
      environment.repoPerEnv = repoPerEnv
      environment.gitopsRepo = apps

      setContext({
        ...data,
        stackConfig: {
          ...data.stackConfig,
          config: {
            ...data.stackConfig.config,
            gimletAgent: {
              ...data.stackConfig.config.gimletAgent,
              environment: env
            },
            gimletd: {
              ...data.stackConfig.config.gimletd,
              environments: [environment]
            }
          }
        }
      })
    }).catch(err => {
        console.error(`Error: ${err}`);
      });
  }, [getContext]);

  useEffect(() => {
    if (repoPerEnv) {
      setInfra(`gitops-${env}-infra`);
      setApps(`gitops-${env}-apps`);
    } else {
      setInfra(`gitops-infra`);
      setApps(`gitops-apps`);
    }
    if(context) {
      setUserValuesInStackConfig(context)
    }
  }, [repoPerEnv, env]);

  if (!context) {
    return null;
  }

  const setUserValuesInStackConfig = (data) => {
    let environment = data.stackConfig.config.gimletd.environments[0]
    environment.name = env
    environment.repoPerEnv = repoPerEnv
    environment.gitopsRepo = repoPerEnv ? `gitops-${env}-apps` : `gitops-apps`

    setContext({
      ...data,
      stackConfig: {
        ...data.stackConfig,
        config: {
          ...data.stackConfig.config,
          gimletAgent: {
            ...data.stackConfig.config.gimletAgent,
            environment: env
          },
          gimletd: {
            ...data.stackConfig.config.gimletd,
            environments: [environment]
          }
        }
      }
    })
  }

  const setValues = (variable, values, nonDefaultValues) => {
    setContext({
      ...context,
      stackConfig: {
        ...context.stackConfig,
        config: {
          ...context.stackConfig.config,
          [variable]: nonDefaultValues
        }
      }
    })
  }

  const validationCallback = (variable, validationErrors) => {
    if (validationErrors !== null) {
      console.log(validationErrors)
    }
  }

  console.log(context.stackConfig.config)

  return (
    <div className="mt-32 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        <div className="md:flex md:items-center md:justify-between">
          <div className="flex-1 min-w-0">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate mb-12">Gimlet Installer</h2>
          </div>
        </div>
        <nav aria-label="Progress">
          <ol className="border border-gray-300 rounded-md divide-y divide-gray-300 md:flex md:divide-y-0">
            <li className="relative md:flex-1 md:flex">
              {/* <!-- Completed Step --> */}
              <div className="group flex items-center w-full select-none cursor-default">
                <span className="px-6 py-4 flex items-center text-sm font-medium">
                  <span className="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-indigo-600 rounded-full">
                    {/* <!-- Heroicon name: solid/check --> */}
                    <svg className="w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                      fill="currentColor" aria-hidden="true">
                      <path fillRule="evenodd"
                        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                        clipRule="evenodd" />
                    </svg>
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-900">Create Github Application</span>
                </span>
              </div>

              {/* <!-- Arrow separator for lg screens and up --> */}
              <div className="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                <svg className="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                  <path d="M0 -2L20 40L0 82" vectorEffect="non-scaling-stroke" stroke="currentcolor"
                    strokeLinejoin="round" />
                </svg>
              </div>
            </li>

            <li className="relative md:flex-1 md:flex">
              {/* <!-- Current Step --> */}
              <div className="px-6 py-4 flex items-center text-sm font-medium select-none cursor-default" aria-current="step">
                <span
                  className="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-indigo-600 rounded-full">
                  <span className="text-indigo-600">02</span>
                </span>
                <span className="ml-4 text-sm font-medium text-indigo-600">Prepare gitops repository</span>
              </div>

              {/* <!-- Arrow separator for lg screens and up --> */}
              <div className="hidden md:block absolute top-0 right-0 h-full w-5" aria-hidden="true">
                <svg className="h-full w-full text-gray-300" viewBox="0 0 22 80" fill="none" preserveAspectRatio="none">
                  <path d="M0 -2L20 40L0 82" vectorEffect="non-scaling-stroke" stroke="currentcolor"
                    strokeLinejoin="round" />
                </svg>
              </div>
            </li>

            <li className="relative md:flex-1 md:flex">
              {/* <!-- Upcoming Step --> */}
              <div className="group flex items-center select-none cursor-default">
                <span className="px-6 py-4 flex items-center text-sm font-medium">
                  <span
                    className="flex-shrink-0 w-10 h-10 flex items-center justify-center border-2 border-gray-300 rounded-full">
                    <span className="text-gray-500 ">03</span>
                  </span>
                  <span className="ml-4 text-sm font-medium text-gray-500 ">Bootstrap gitops automation</span>
                </span>
              </div>
            </li>
          </ol>
        </nav>

        {context.appId === "" &&
          <div className="rounded-md bg-red-50 p-4 my-8">
            <div className="flex">
              <div className="flex-shrink-0">
                {/* <!-- Heroicon name: solid/x-circle --> */}
                <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                  aria-hidden="true">
                  <path fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">A Github Application was not yet created by the installer. Go
                  to <a href="/" className="font-bold">Step One</a> to create it.</h3>
              </div>
            </div>
          </div>
        }
        <form action="/bootstrap" method="post">
          <div className="mt-8 text-sm">

            <div className="text-gray-700">
              <div className="flex mt-4">
                <div className="font-medium self-center">Environment name</div>
                <div className="max-w-lg flex rounded-md ml-4">
                  <div className="max-w-lg w-full lg:max-w-xs">
                    <input id="apps" name="env"
                      value={env}
                      onChange={e => setEnv(e.target.value)}
                      className="block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                      type="text" />
                  </div>
                </div>
              </div>
            </div>

            <SeparateEnvironments
              repoPerEnv={repoPerEnv}
              setRepoPerEnv={setRepoPerEnv}
              infraRepo={infra}
              appsRepo={apps}
            />
            <input type="hidden" name="repoPerEnv" value={repoPerEnv} />

            <div className='mt-8 mb-16'>
              {context.stackDefinition && context.stackConfig &&
              <StackUI
                stack={context.stackConfig.config}
                stackDefinition={context.stackDefinition}
                setValues={setValues}
                validationCallback={validationCallback}
                categoriesToRender={['cloud', 'ingress', 'gimlet']}
                componentsToRender={['civo', 'nginx', 'gimletd', 'gimletAgent', 'gimletDashboard']}
                hideTitle={true}
              />
              }
            </div>
            <input type="hidden" name="stackConfig" value={JSON.stringify(context.stackConfig.config)} />

            <div className="p-0 flow-root my-8">
              <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
                <button
                  className="bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700 inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150">
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
