import { useRef, useState, useEffect } from 'react'
import { format, formatDistance } from "date-fns";
import { InformationCircleIcon, XCircleIcon } from '@heroicons/react/solid'
import { StackUI, BootstrapGuide, SeparateEnvironments } from 'shared-components';
import {
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWERRORLIST,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWOPENED,
  ACTION_TYPE_GITOPS_COMMITS
} from "../../redux/redux";

const EnvironmentCard = ({ store, isOnline, env, deleteEnv, gimletClient, refreshEnvs, tab, envFromParams }) => {
  let reduxState = store.getState();
  const [repoPerEnv, setRepoPerEnv] = useState(false)
  const [infraRepo, setInfraRepo] = useState("gitops-infra")
  const [appsRepo, setAppsRepo] = useState("gitops-apps")
  /*eslint no-unused-vars: ["error", { "varsIgnorePattern": "popupWindow" }]*/
  const [popupWindow, setPopupWindow] = useState(reduxState.popupWindow)
  const [gitopsCommits, setGitopsCommits] = useState(reduxState.gitopsCommits);
  const [bootstrapMessage, setBootstrapMessage] = useState(undefined);
  const ref = useRef();

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

  store.subscribe(() => {
    let reduxState = store.getState();
    setPopupWindow(reduxState.popupWindow);
    setGitopsCommits(reduxState.gitopsCommits);
  });

  function scrollTo(ref) {
    if (!ref.current) return;
    ref.current.scrollIntoView({ behavior: "smooth" });
  }

  if (!tab || envFromParams !== env.name) {
    tab = "";
  }

  const [tabs, setTabs] = useState([
    { name: "Gitops repositories", current: tab === "" },
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

  let initStack = {};
  if (env.stackConfig) {
    initStack = env.stackConfig.config;
  }

  const [stack, setStack] = useState(initStack);
  const [stackNonDefaultValues, setStackNonDefaultValues] = useState(initStack);
  const [errors, setErrors] = useState({});

  const gitopsRepositories = [
    { name: env.infraRepo, href: `https://github.com/${env.infraRepo}` },
    { name: env.appsRepo, href: `https://github.com/${env.appsRepo}` }
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
      type: ACTION_TYPE_POPUPWINDOWOPENED, payload: {
        header: "Saving..."
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
        resetPopupWindowAfterThreeSeconds()
        return false
      }
    }

    gimletClient.saveInfrastructureComponents(env.name, stackNonDefaultValues)
      .then(() => {
        console.log("Components saved")
        refreshEnvs();
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Component saved"
          }
        });
        resetPopupWindowAfterThreeSeconds()
      }, (err) => {
        console.log("Couldn't save components");
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
        resetPopupWindowAfterThreeSeconds()
      })
  }

  const bootstrapGitops = (envName, repoPerEnv) => {
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWOPENED, payload: {
        header: "Bootstrapping..."
      }
    });

    gimletClient.bootstrapGitops(envName, repoPerEnv, infraRepo, appsRepo)
      .then((data) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Gitops environment bootstrapped"
          }
        });
        refreshEnvs();
        setBootstrapMessage(data)
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
    return (
      <div className="mt-4">
        {gitopsRepositories.map((gitopsRepo) =>
        (
          <div className="flex">
            <a className="mb-1 font-mono text-sm text-gray-500 hover:text-gray-600" href={gitopsRepo.href} target="_blank" rel="noreferrer">{gitopsRepo.name}
              <svg xmlns="http://www.w3.org/2000/svg"
                className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
                viewBox="0 0 24 24">
                <path d="M0 0h24v24H0z" fill="none" />
                <path
                  d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
              </svg>
            </a>
          </div>
        ))}
      </div>
    )
  }

  const refreshGitopsCommits = () => {
    gimletClient.getGitopsCommits()
      .then(data => store.dispatch({
        type: ACTION_TYPE_GITOPS_COMMITS, payload:
          data
      }), () => {
        /* Generic error handler deals with it */
      });
  }

  const configureAgent = (envName) => {
    console.log(envName)
    console.log("will call a dedicated API endpoint here")
  }

  const gitopsCommitColorByStatus = (status) => {
    return status.includes("Succeeded") ?
      "green"
      :
      status.includes("Failed") ?
        "red"
        :
        "yellow"
  }

  const renderGitopsCommit = (gitopsCommit, idx, arr) => {
    const gitopsCommitSha = gitopsCommit.sha.slice(0, 6);
    const exactDate = format(gitopsCommit.created * 1000, 'h:mm:ss a, MMMM do yyyy');
    const dateLabel = formatDistance(gitopsCommit.created * 1000, new Date());
    const gitopsCommitColor = gitopsCommitColorByStatus(gitopsCommit.status);

    return (<li key={idx}
      className={`bg-${gitopsCommitColor}-100 hover:bg-${gitopsCommitColor}-200 p-4 rounded`}
    >
      <div className="relative">
        {idx !== arr.length - 1 &&
          <span className="absolute top-8 left-4 -ml-px h-full w-0.5 bg-gray-400" aria-hidden="true"></span>
        }
        <div className="relative flex items-start space-x-3">
          <img
            className={`h-8 w-8 rounded-full ring-gray-400 flex items-center justify-center ring-4`}
            src={`https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png`}
            alt="triggerer"
          />
          <div className="min-w-0 flex-1">
            <p className="font-medium text-gray-900">{gitopsCommit.status}</p>
            <p className=" -mt-1">
              <a
                className="text-xs text-gray-500 hover:text-gray-600"
                title={exactDate}
                href={`https://github.com/${env.appsRepo}/commit/${gitopsCommit.sha}`}
                target="_blank"
                rel="noopener noreferrer">
                <span className="font-mono">{gitopsCommitSha}</span> state recorded {dateLabel} ago
              </a>
            </p>
            <p className="text-gray-700 mt-2">
              <span>{gitopsCommit.statusDesc}</span>
            </p>
          </div>
        </div>
      </div>
    </li>)
  }

  const gitopsCommitsTab = () => {
    return (
      <div className="flow-root">
        <ul className="mt-4">
          <div className="flow-root">
            <svg onClick={() => refreshGitopsCommits()} xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 mb-4 text-gray-500 hover:text-gray-600 cursor-pointer float-right" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          {gitopsCommits.filter(gitopsCommit => gitopsCommit.env === env.name).map((gitopsCommit, idx, arr) => renderGitopsCommit(gitopsCommit, idx, arr))}
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
        <div className="mt-4 rounded-md bg-blue-50 p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
            </div>
            <div className="ml-3 md:justify-between">
              <p className="text-sm text-blue-500">
                By default, infrastructure manifests of this environment will be placed in the <span className="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded">{env.name}</span> folder of the shared <span className="text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">gitops-infra</span> git repository,
                <br />
                and application manifests will be placed in the <span className="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded">{env.name}</span> folder of the shared <span className="text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded">gitops-apps</span> git repository
              </p>
            </div>
          </div>
        </div>
        <SeparateEnvironments
          repoPerEnv={repoPerEnv}
          setRepoPerEnv={setRepoPerEnv}
          infraRepo={infraRepo}
          appsRepo={appsRepo}
          setInfraRepo={setInfraRepo}
          setAppsRepo={setAppsRepo}
        />
        <div className="p-0 flow-root mt-8">
          <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
            <button
              onClick={() => bootstrapGitops(env.name, repoPerEnv)}
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

  const gimletAgentConfigured = stack.gimletAgent && stack.gimletAgent.enabled;

  return (
    <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
      <div ref={ref} className="px-4 py-5 sm:px-6">
        <div className="flex justify-between">
          <div className="inline-flex">
            <h3 className="text-lg leading-6 font-medium text-gray-900 pr-1">
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
            {!isOnline && !gimletAgentConfigured &&
            <>
              <div className="rounded-md bg-red-50 p-4">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <XCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-red-800">This environment is disconnected</h3>
                    <div className="mt-2 text-sm text-red-700">
                      Configure the Gimlet Agent for realtime Kubernetes data under <span className="italic">Infrastructure components &gt; Gimlet Agent</span><br />
                      Or use the <span
                        className="font-medium cursor-pointer"
                        onClick={(e) => {
                          // eslint-disable-next-line no-restricted-globals
                          confirm('The 1-click-config will place a commit in your gitops repo.\nAre you sure you want proceed?') &&
                            configureAgent(env.name, e);
                        }}
                      >1-click-config</span>.
                    </div>
                  </div>
                </div>
              </div>

              <div className="rounded-md bg-red-50 p-4 mt-2">
              <div className="flex">
                <div className="flex-shrink-0">
                  <XCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800">Deployment automation is not configured for this environment</h3>
                  <div className="mt-2 text-sm text-red-700">
                    Configure Gimletd to be able to deploy to this environment under <span className="italic">Infrastructure components &gt; Gimletd</span><br />
                    Or use the <span
                      className="font-medium cursor-pointer"
                      onClick={(e) => {
                        // eslint-disable-next-line no-restricted-globals
                        confirm('The 1-click-config will place a commit in your gitops repo.\nAre you sure you want proceed?') &&
                          configureAgent(env.name, e);
                      }}
                    >1-click-config</span>.
                  </div>
                </div>
              </div>
            </div>
            </>
            }
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
            {bootstrapMessage &&
              <>
                <h3 className="text-2xl font-bold p-2 mt-4 text-gray-900">Finalize Gitops bootstrapping with these two steps below</h3>
                <BootstrapGuide
                  envName={bootstrapMessage.envName}
                  repoPath={bootstrapMessage.infraRepo}
                  repoPerEnv={bootstrapMessage.repoPerEnv}
                  publicKey={bootstrapMessage.infraPublicKey}
                  secretFileName={bootstrapMessage.infraSecretFileName}
                  gitopsRepoFileName={bootstrapMessage.infraGitopsRepoFileName}
                  isNewRepo={bootstrapMessage.isNewInfraRepo}
                />
                <BootstrapGuide
                  envName={bootstrapMessage.envName}
                  repoPath={bootstrapMessage.appsRepo}
                  repoPerEnv={bootstrapMessage.repoPerEnv}
                  publicKey={bootstrapMessage.appsPublicKey}
                  secretFileName={bootstrapMessage.appsSecretFileName}
                  gitopsRepoFileName={bootstrapMessage.appsGitopsRepoFileName}
                  isNewRepo={bootstrapMessage.isNewAppsRepo}
                  notificationsFileName={bootstrapMessage.notificationsFileName}
                />
                <h2 className='text-gray-900'>Happy Gitopsing🎊</h2>
              </>
            }
            <div className="hidden sm:block">
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
          :
          gitopsBootstrapWizard()
        }
      </div>
    </div >
  )
};

export default EnvironmentCard;
