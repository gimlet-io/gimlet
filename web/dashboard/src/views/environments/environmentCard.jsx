import { useRef, useState, useEffect } from 'react'
import { format, formatDistance } from "date-fns";
import StackUI from './stack-ui';
import BootstrapGuide from './bootstrapGuide';
import SeparateEnvironments from './separateEnvironments';
import KustomizationPerApp from './kustomizationPerApp';
import { InformationCircleIcon } from '@heroicons/react/solid'
import {
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWERRORLIST,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_ENVUPDATED,
  ACTION_TYPE_SAVE_ENV_PULLREQUEST,
  ACTION_TYPE_RELEASE_STATUSES
} from "../../redux/redux";
import { renderPullRequests } from '../../components/env/env';
import { rolloutWidget } from '../../components/rolloutHistory/rolloutHistory';

const EnvironmentCard = ({ store, isOnline, env, deleteEnv, gimletClient, refreshEnvs, tab, envFromParams, releaseStatuses, popupWindow, pullRequests, scmUrl, host, userToken }) => {
  const [repoPerEnv, setRepoPerEnv] = useState(true)
  const [kustomizationPerApp, setKustomizationPerApp] = useState(false)
  const [infraRepo, setInfraRepo] = useState("gitops-infra")
  const [appsRepo, setAppsRepo] = useState("gitops-apps")
  const ref = useRef();

  useEffect(() => {
    gimletClient.getReleases(env.name, 10)
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_RELEASE_STATUSES,
          payload: {
            envName: env.name,
            data: data,
          }
        });
      }, () => {/* Generic error handler deals with it */
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (repoPerEnv && infraRepo === "gitops-infra") {
    setInfraRepo(`gitops-${env.name}-infra`);
  }
  if (repoPerEnv && appsRepo === "gitops-apps") {
    setAppsRepo(`gitops-${env.name}-apps`);
  }
  if (!repoPerEnv && infraRepo === `gitops-${env.name}-infra`) {
    setInfraRepo("gitops-infra");
  }
  if (!repoPerEnv && appsRepo === `gitops-${env.name}-apps`) {
    setAppsRepo("gitops-apps");
  }

  function scrollTo(ref) {
    if (!ref.current) return;
    ref.current.scrollIntoView({ behavior: "smooth" });
  }

  if (!tab || envFromParams !== env.name) {
    tab = "";
  }

  const [tabs, setTabs] = useState([
    { name: "Gitops configs", current: tab === "" },
    { name: "Infrastructure components", current: tab === "components" },
    { name: "Gitops commits", current: tab === "gitops-commits" }
  ]);

  useEffect(() => {
    if (envFromParams === env.name) {
      switchTabHandler("Gitops commits");

      scrollTo(ref);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [envFromParams, env.name]);

  const hasGitopsRepo = env.infraRepo !== "";

  const [stack, setStack] = useState({});
  const [stackNonDefaultValues, setStackNonDefaultValues] = useState({});
  const [errors, setErrors] = useState({});

  useEffect(() => {
    if (env.stackConfig) {
      setStack(env.stackConfig.config);
      setStackNonDefaultValues(env.stackConfig.config);
    }
  }, [env.stackConfig]);

  const gitopsRepositories = [
    { name: env.infraRepo, href: `${scmUrl}/${env.infraRepo}` },
    { name: env.appsRepo, href: `${scmUrl}/${env.appsRepo}` }
  ];

  const switchTabHandler = (tabName) => {
    setTabs(tabs.map(tab => {
      if (tab.name === tabName) {
        return { ...tab, current: true }
      } else {
        return { ...tab, current: false }
      }
    }))
  }

  const setValues = (variable, values, nonDefaultValues) => {
    setStack({ ...stack, [variable]: values })
    setStackNonDefaultValues({ ...stackNonDefaultValues, [variable]: nonDefaultValues })
  }

  const validationCallback = (variable, validationErrors) => {
    if (validationErrors !== null) {
      validationErrors = validationErrors.filter(error => error.keyword !== 'oneOf');
      validationErrors = validationErrors.filter(error => error.dataPath !== '.enabled');
    }

    setErrors({ ...errors, [variable]: validationErrors })
  }

  const resetPopupWindowAfterThreeSeconds = () => {
    setTimeout(() => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  };

  const saveComponents = () => {
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Saving components..."
      }
    });

    for (const variable of Object.keys(errors)) {
      if (errors[variable] !== null) {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERRORLIST, payload: {
            header: "Error",
            errorList: errors
          }
        });
        return false
      }
    }

    gimletClient.saveInfrastructureComponents(env.name, stackNonDefaultValues)
      .then((data) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Pull request was created",
            link: data.createdPr.link
          }
        });
        store.dispatch({
          type: ACTION_TYPE_SAVE_ENV_PULLREQUEST, payload: {
            envName: data.envName,
            createdPr: data.createdPr
          }
        });
        store.dispatch({
          type: ACTION_TYPE_ENVUPDATED, name: env.name, payload: data.stackConfig
        });
      }, (err) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
        resetPopupWindowAfterThreeSeconds()
      })
  }

  const bootstrapGitops = (envName, repoPerEnv, kustomizationPerApp) => {
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Bootstrapping..."
      }
    });

    gimletClient.bootstrapGitops(envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo)
      .then(() => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Gitops environment bootstrapped"
          }
        });
        refreshEnvs();
        resetPopupWindowAfterThreeSeconds()
      }, (err) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
        resetPopupWindowAfterThreeSeconds()
      })
  }

  const gitopsRepositoriesTab = () => {
    if (!env.infraRepo || !env.appsRepo) {
      return null;
    }

    const isRepoPerEnvEnabled = env.repoPerEnv ? "enabled" : "disabled";
    const isKustomizationPerAppEnabled = env.kustomizationPerApp ? "enabled" : "disabled";

    return (
      <div className="mt-4 text-sm text-gray-500 px-2">
        <div className="space-y-1">
          <span className="flex"><p className="mr-1 font-semibold text-gray-600">Kustomization per app setting</p> is {isKustomizationPerAppEnabled} for this environment.</span>
          <span className="flex"><p className="mr-1 font-semibold text-gray-600">Separate environments by git repositories setting</p> is {isRepoPerEnvEnabled} for this environment.</span>
        </div>
        <div className="mt-4 mb-1">
          <span className="font-bold text-gray-600">Gitops repositories</span>
          <div className="ml-4 mt-1 font-mono">
            {gitopsRepositories.map((gitopsRepo) =>
            (
              <div className="flex" key={gitopsRepo.href}>
                { !env.builtIn &&
                <a className="mb-1 hover:text-gray-600" href={gitopsRepo.href} target="_blank" rel="noreferrer">{gitopsRepo.name}
                  <svg xmlns="http://www.w3.org/2000/svg"
                    className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
                    viewBox="0 0 24 24">
                    <path d="M0 0h24v24H0z" fill="none" />
                    <path
                      d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                  </svg>
                </a>
                }
                { env.builtIn &&
                  <div className="mb-1" href={gitopsRepo.href} target="_blank" rel="noreferrer">{gitopsRepo.name}</div>
                }
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  const refreshReleaseStatuses = () => {
    gimletClient.getReleases(env.name, 10)
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_RELEASE_STATUSES,
          payload: {
            envName: env.name,
            data: data,
          }
        });
      }, () => {/* Generic error handler deals with it */
      })
  }

  const gitopsCommitsTab = () => {
    if (!releaseStatuses) {
      return null
    }

    let renderReleaseStatuses = [];

    releaseStatuses.forEach((rollout, idx, arr) => {
      const exactDate = format(rollout.created * 1000, 'h:mm:ss a, MMMM do yyyy');
      const dateLabel = formatDistance(rollout.created * 1000, new Date());

      renderReleaseStatuses.unshift(rolloutWidget(idx, arr, exactDate, dateLabel, undefined, undefined, undefined, undefined, rollout, scmUrl, env.builtIn))
    })

    return (
      <div className="flow-root">
        <ul className="mt-4">
          <div className="flow-root">
            <svg onClick={() => refreshReleaseStatuses()} xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 mb-4 text-gray-500 hover:text-gray-600 cursor-pointer float-right" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          <div className="flow-root">
            <ul>
              {renderReleaseStatuses}
            </ul>
          </div>
        </ul>
      </div>
    )
  }

  const infrastructureComponentsTab = () => {
    return (
      <div className="mt-4 text-gray-700">
        <div>
          <StackUI
            stack={stack}
            stackDefinition={env.stackDefinition}
            setValues={setValues}
            validationCallback={validationCallback}
          />
          { !env.builtIn &&
          <div className="p-0 flow-root my-8">
            <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
              <button
                onClick={() => saveComponents()}
                disabled={popupWindow.visible}
                className={(popupWindow.visible ? 'bg-gray-600 cursor-default' : 'bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700') + ` inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150`}              >
                Save components
              </button>
            </span>
          </div>
          }
        </div>
      </div>
    )
  }

  const gitopsBootstrapWizard = () => {
    return (
      <>
        <div className="mt-2 pb-4 border-b border-gray-200">
          <h3 className="text-lg leading-6 font-medium text-gray-900">Bootstrap gitops repository</h3>
          <p className="mt-2 max-w-4xl text-sm text-gray-500">
            To initialize this environment, bootstrap the gitops repository first
          </p>
        </div>
        <KustomizationPerApp
          kustomizationPerApp={kustomizationPerApp}
          setKustomizationPerApp={setKustomizationPerApp}
        />
        <SeparateEnvironments
          repoPerEnv={repoPerEnv}
          setRepoPerEnv={setRepoPerEnv}
          infraRepo={infraRepo}
          appsRepo={appsRepo}
          setInfraRepo={setInfraRepo}
          setAppsRepo={setAppsRepo}
          envName={env.name}
        />
        <div className="p-0 flow-root mt-8">
          <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
            <button
              onClick={() => bootstrapGitops(env.name, repoPerEnv, kustomizationPerApp)}
              disabled={popupWindow.visible}
              className={(popupWindow.visible ? 'bg-gray-600 cursor-default' : 'bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700') + ` inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150`}
            >
              Bootstrap gitops repository
            </button>
          </span>
        </div>
      </>
    )
  }

  const builtInEnvInfo = () => {
    return (
      <div className="rounded-md bg-blue-50 p-4 mb-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
        </div>
        <div className="ml-3">
          <h3 className="text-sm font-medium text-blue-800">This is a built-in environment</h3>
          <div className="mt-2 text-sm text-blue-700">
            Gimlet made this environment for you so you can quickly get started, but you can't make changes to it.<br />
            Create another environment to tailor it to your needs.
          </div>
        </div>
      </div>
      </div>
    );
  }

  return (
    <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
      {renderPullRequests(pullRequests)}
      <div ref={ref} className="px-4 py-5 sm:px-6">
        <div className="flex justify-between">
          <div className="inline-flex">
            <h3 className="text-lg leading-6 capitalize font-medium text-gray-900 pr-1">
              {env.name}
            </h3>
            <span title={isOnline ? "Connected" : "Disconnected"}>
              <svg className={(isOnline ? "text-green-400" : "text-red-400") + " inline fill-current ml-1"} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
                <path
                  d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
                />
              </svg>
            </span>
            {!hasGitopsRepo &&
              <span title="Gitops automation is not bootstrapped">
                <svg xmlns="http://www.w3.org/2000/svg" className="ml-2 h-6 w-6 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </span>}
          </div>
          {!isOnline &&
            <div className="inline-flex">
              <svg xmlns="http://www.w3.org/2000/svg" onClick={deleteEnv} className="cursor-pointer inline text-red-400 hover:text-red-600 h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </div>
          }
        </div>
      </div>
      <div className="px-4 py-5 sm:px-6">
        {hasGitopsRepo ?
          <>
            <div className="sm:hidden">
              <label htmlFor="tabs" className="sr-only">
                Select a tab
              </label>
              <select
                id="tabs"
                name="tabs"
                className="block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                defaultValue={tabs.find((tab) => tab.current).name}
              >
                {tabs.map((tab) => (
                  <option key={tab.name}>{tab.name}</option>
                ))}
              </select>
            </div>
            {!isOnline &&
              <div className="mb-4">
                <h3 className="text-lg font-medium p-2 text-gray-900">Connect your cluster</h3>
                <BootstrapGuide
                  envName={env.name}
                  host={host}
                  token={userToken}
                />
              </div>
            }
            {isOnline &&
            <>
              <div className="hidden sm:block">
                {env.builtIn && builtInEnvInfo()}
                <div className="border-b border-gray-200">
                  <nav className="-mb-px flex space-x-8" aria-label="Tabs">
                    {tabs.map((tab) => (
                      <div
                        key={tab.name}
                        className={(
                          tab.current
                            ? "border-indigo-500 text-indigo-600"
                            : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300") +
                          " cursor-pointer select-none whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm"
                        }
                        aria-current={tab.current ? "page" : undefined}
                        onClick={() => switchTabHandler(tab.name)}
                      >
                        {tab.name}
                      </div>
                    ))}
                  </nav>
                </div>
              </div>
              {tabs[0].current ?
                gitopsRepositoriesTab()
                :
                tabs[1].current ?
                  infrastructureComponentsTab()
                  :
                  gitopsCommitsTab()
              }
            </>
            }
          </>
          :
          gitopsBootstrapWizard()
        }
      </div>
    </div >
  )
};

export default EnvironmentCard;
