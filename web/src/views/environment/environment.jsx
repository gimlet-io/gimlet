import { useState, useEffect } from 'react';
import {
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWERRORLIST,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_ENVSPINNEDOUT,
  ACTION_TYPE_ENVS,
} from "../../redux/redux";
import General from './general';
import Category from './category';
import { SideBar } from '../envConfig/envConfig';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import {produce} from 'immer';
import yaml from "js-yaml";
import ReactDiffViewer from "react-diff-viewer";
import { Modal } from '../../components/modal'
import * as Diff from "diff";
import { InformationCircleIcon } from '@heroicons/react/20/solid';
import { format, formatDistance } from "date-fns";

export default function EnvironmentView(props) {
  const { store, gimletClient } = props
  const reduxState = props.store.getState();
  const { env } = props.match.params;

  const [connectedAgents, setConnectedAgents] = useState(reduxState.connectedAgents)
  const [environment, setEnvironment] = useState(findEnv(reduxState.envs, env))
  const [user, setUser] = useState(reduxState.user)
  const [pullRequests, setPullRequests] = useState([])
  const [settings, setSettings] = useState(reduxState.settings)
  const [errors, setErrors] = useState({})
  // eslint-disable-next-line no-unused-vars
  const [popupWindow, setPopupWindow] = useState()
  const [stackConfig, setStackConfig] = useState()
  const [savedStackConfig, setSavedStackConfig] = useState()
  // const [stackConfigLoaded, setStackConfigLoaded] = useState()
  const [stackDefinition, setStackDefinition] = useState()
  const [isOnline, setIsOnline] = useState(false)
  const [navigation, setNavigation] = useState([])
  const [showModal, setShowModal] = useState(false)

  store.subscribe(() => {
    const reduxState = store.getState()
    setConnectedAgents(reduxState.connectedAgents)
    setEnvironment(findEnv(reduxState.envs, env))
    setUser(reduxState.user)
    setPopupWindow(reduxState.popupWindow)
    setSettings(reduxState.settings)
  })

  useEffect(() => {
    setIsOnline(
      Object.keys(connectedAgents)
        .map(e => connectedAgents[e])
        .some(e => {
          if (!e || !env) { // newly created envs are not part of the data model
            return false
          }
          return e.name === env
        })
    )
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connectedAgents, environment]);

  useEffect(() => {
    //stack config changes
    gimletClient.getPullRequestsFromInfraRepo(env)
      .then(data => {
        setPullRequests(prevState => [...prevState, ...data]);
      }, () => {/* Generic error handler deals with it */
      });

    //stack version updates
    gimletClient.getGitopsUpdatePullRequests(env)
      .then(data => {
        setPullRequests(prevState => [...prevState, ...data]);
      }, () => {/* Generic error handler deals with it */
      });
    
    if (environment) {
      if (environment.infraRepo !== "") {
        gimletClient.getStackConfig(env)
          .then(data => {
            setStackConfig(data.stackConfig.config)
            let deepCopied = JSON.parse(JSON.stringify(data.stackConfig.config))
            setSavedStackConfig(deepCopied)
            setStackDefinition(data.stackDefinition)
            setNavigation(translateToNavigation(data.stackDefinition))
            // setStackConfigLoaded(true)
          }, () => {/* Generic error handler deals with it */
          })
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [environment]);

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

    gimletClient.saveInfrastructureComponents(environment.name, stackConfig)
      .then((data) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Pull request was created",
            link: data.createdPr.link
          }
        });
        setPullRequests(prevState => [...prevState, data.createdPr]);
        let deepCopied = JSON.parse(JSON.stringify(stackConfig))
        setSavedStackConfig(deepCopied)
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

  const resetPopupWindowAfterThreeSeconds = () =>  {
    setTimeout(() => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  };

  const setValues = (variable, values, nonDefaultValues) => {
    if (!stackConfig[variable] && Object.keys(nonDefaultValues).length === 0) {
      return
    }
    if (stackConfig[variable] && Object.keys(nonDefaultValues).length === 0) {
      setStackConfig(produce(stackConfig, draft => {
        delete draft[variable]
      }))
      return
    }

    setStackConfig(produce(stackConfig, draft => {
      draft[variable]=nonDefaultValues
    }))
  }

  const validationCallback = (variable, validationErrors) => {
    if (validationErrors !== null) {
      validationErrors = validationErrors.filter(error => error.keyword !== 'oneOf');
      validationErrors = validationErrors.filter(error => error.dataPath !== '.enabled');
    }

    setErrors(prevState => ({ ...prevState, [variable]: validationErrors }))
  }

  const stopEnv = () => {
    gimletClient.stopEnv()
    gimletClient.getEnvs()
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_ENVS,
          payload: data
        });
      }, () => {/* Generic error handler deals with it */}
      );
  }

  if (!environment) {
    return <SkeletonLoader />;
  }


  let hasChange = false
  var addedLines, removedLines
  if (stackConfig) {
    const stackConfigString = JSON.stringify(stackConfig)
    if (savedStackConfig) {
      const savedStackConfigString = JSON.stringify(savedStackConfig)
      hasChange = stackConfigString !== savedStackConfigString
      const diffStat = Diff.diffChars(savedStackConfigString, stackConfigString);
      const addedStat = diffStat.find(stat=>stat.added)?.count
      const removedStat = diffStat.find(stat=>stat.removed)?.count
      addedLines = addedStat ? addedStat : 0
      removedLines = removedStat ? removedStat : 0
    }  
  }
  
  let selectedNavigation = navigation.find(i => props.location.pathname.endsWith(i.href))
  if (!selectedNavigation) {
    selectedNavigation = navigation[0]
  }

  const expiringAt = new Date(environment.expiry * 1000);
  const expired = expiringAt < new Date()
  const exactDate = format(environment.expiry * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(environment.expiry * 1000, new Date());

  return (
    <>
      {showModal &&
        <Modal closeHandler={() => setShowModal(false)}>
          <ReactDiffViewer
              oldValue={yaml.dump(savedStackConfig)}
              newValue={yaml.dump(stackConfig)}
              splitView={false}
              showDiffOnly={false}
              useDarkTheme={document.documentElement.classList.contains('dark')}
              styles={{
                diffContainer: {
                  overflowX: "auto",
                  display: "block",
                  height: "100%",
                  "& pre": { whiteSpace: "pre" }
                },
              }} />
        </Modal>
      }
      <div className="fixed w-full bg-neutral-100 dark:bg-neutral-900 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 pb-8 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow py-0.5">Environment config
          <span className={`px-2 py-0.5 mx-1 ${isOnline ? 'text-teal-800 dark:text-teal-400' : 'text-red-700 dark:text-red-300'} text-xs font-medium ${isOnline ? 'bg-teal-100 dark:bg-teal-700' : 'bg-red-200 dark:bg-red-700'} rounded-full`}>
            {isOnline ? 'connected' : 'disconnected'}
          </span>
          </h1>
          {hasChange &&
            <span className="mr-8 text-sm bg-neutral-300 dark:bg-neutral-600 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-600 dark:text-neutral-300 ml-2 px-1 rounded-md cursor-pointer"
              onClick={()=> setShowModal(true)}
            >
              <span>Review changes (</span>
              <span className="font-mono text-teal-500">+{addedLines}</span>
              <span className="font-mono ml-1 text-red-500">-{removedLines}</span>
              <span>)</span>
            </span>
          }
          { stackConfig && !environment.builtIn &&
          <button
            type="button"
            disabled={!hasChange}
            className={(hasChange ? 'primaryButton' : 'primaryButtonDisabled') + ` px-4`}
            onClick={saveComponents}
          >
            Save
          </button>
          }
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-48">
      </div>
      {settings.trial &&
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-6">
      <div className="rounded-md bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300 p-4 w-full">
        <div className="flex">
          <div className="flex-shrink-0">
            <InformationCircleIcon className="h-5 w-5 " aria-hidden="true" />
          </div>
          <div className="ml-3 flex-1 md:flex md:justify-between">
            <div className="text-sm flex flex-col">
              <p className="font-semibold text-sm pb-4">This is an ephemeral environment</p>
              <p>Gimlet made this environment for you so you can start deploying.</p>
              <p className='pb-4'>This environment will run on our Kubernetes cluster for 7 days with plenty of resources for you to get started.</p>
              <p>Once you upgrade Gimlet, you will be able to connect your own Kubernetes cluster running on your preferred provider.</p>
              <p>We have a <a href="https://gimlet.io" className='underline' target="_blank" rel="noopener noreferrer">few recommendations</a> about how you can keep your Kubernetes experience simple and cheap.</p>
                
              {expired &&
              <p className='pt-4'>This environment was disabled <span className='font-medium text-red-500' title={`at ${exactDate}`}>{dateLabel} ago</span>.</p>
              }
              {!expired &&
              <p className='pt-4'>This environment will be disabled <span className='font-medium text-red-500' title={`at ${exactDate}`}>in {dateLabel}</span>.</p>
              }
              <p><a href={`https://gimletio.lemonsqueezy.com/buy/6305a31b-ad52-490a-8d8e-5ba3ab68147f?checkout[custom][instance]=${settings.instance}`} className='underline' target="_blank" rel="noopener noreferrer">Upgrade Gimlet</a> to keep your deployments running.</p>
            </div>
          </div>
        </div>
      </div>
      </div>
      }
      {!settings.trial && environment.ephemeral &&
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-6">
      <div className="rounded-md bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300 p-4 w-full">
        <div className="flex">
          <div className="flex-shrink-0">
            <InformationCircleIcon className="h-5 w-5 " aria-hidden="true" />
          </div>
          <div className="ml-3 flex-1 md:flex md:justify-between">
            <div className="text-sm flex flex-col">
              <p className="font-semibold text-sm pb-4">Thank you for your purchase 🎉</p>
              <p>Now it is time to connect your Kubernetes cluster running on your preferred provider.</p>
              <p className='pt-4'>Follow one of the following tutorials:</p>
              <ul className='list-disc ml-8'>
                <li><a href="" className='underline'>CIVO Cloud</a></li>
                <li><a href="" className='underline'>Digital Ocean</a></li>
                <li><a href="" className='underline'>Linode</a></li>
                <li><a href="" className='underline'>Scaleway</a></li>
                <li><a href="" className='underline'>Any other Kubernetes cluster</a></li>
              </ul>

              {expired &&
              <p className='pt-4'>
                This environment was disabled <span className='font-medium text-red-500' title={`at ${exactDate}`}>{dateLabel} ago</span>, you can
                <a
                  href="#"
                  className='underline ml-1'
                  onClick={() => {
                    // eslint-disable-next-line no-restricted-globals
                    confirm(`Are you sure you want to stop the ephemeral environment?`) &&
                    stopEnv();
                  }}
                  >
                  start the migration here
                </a>.
              </p>
              }
              {!expired &&
              <>
                <p className='pt-4'>To ease the transition we will host this environment for another 7 days.</p>
                <p className='pt-4'>
                  This environment will be disabled <span className='font-medium text-red-500' title={`at ${exactDate}`}>in {dateLabel}</span>, or you can
                  <a
                  href="#"
                  className='underline ml-1'
                  onClick={() => {
                    // eslint-disable-next-line no-restricted-globals
                    confirm(`Are you sure you want to stop the ephemeral environment?`) &&
                    stopEnv();
                  }}
                  >
                    start the migration here
                  </a>.
                </p>
              </>
              }
            </div>
          </div>
        </div>
      </div>
      </div>
      }
      { stackConfig &&
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex">
        <PullRequests title="Open Pull Requests" items={pullRequests} />
      </div>
      }
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-8">
        <div className="sticky top-0 h-96 top-56">
          <SideBar
            location={props.location}
            history={props.history}
            navigation={navigation}
            selected={selectedNavigation}
          />
        </div>
        <div className="w-full ml-14 space-y-8">
          { environment.builtIn && 
            <div className="w-full">
              <div className='w-full card'>
                <div className="p-6 pb-4 items-center">
                  <label htmlFor="label-title" className="block font-medium">
                    Create Gitops Repositories
                  </label>
                  <p className="text-sm text-neutral-800 dark:text-neutral-400 mt-4">
                    To make edits to this environment, create the gitops repositories first.
                    <br />
                    Gimlet will create two git repositories to host the infrastructure and application manifests.
                  </p>
                </div>
                <div className='learnMoreBox flex items-center'>
                  <div className='flex-grow'>
                    Learn more about <a href="https://gimlet.io" className='learnMoreLink'>Gitops <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
                  </div>
                  <button
                    type="button"
                    className="primaryButton px-4"
                    onClick={() => {
                      // eslint-disable-next-line no-restricted-globals
                      confirm(`Are you sure you want to convert to a gitops environment?`) &&
                        spinOutBuiltInEnv(store, gimletClient)
                    }}
                  >Create</button>
                </div>
              </div>
            </div>
          }
          { (!selectedNavigation || selectedNavigation?.name === "General") &&
            <General
              gimletClient={gimletClient}
              store={store}
              environment={environment}
              scmUrl={settings.scmUrl}
              provider={settings.provider}
              isOnline={isOnline}
              userToken={user.token}
              history={props.history}
            />
          }
          { selectedNavigation && selectedNavigation.name !== "General" &&  !environment.builtIn &&
            <Category
              category={selectedNavigation.category}
              stackDefinition={stackDefinition}
              stackConfig={stackConfig}
              environment={environment}
              gimletClient={gimletClient}
              store={store}
              setValues={setValues}
              validationCallback={validationCallback}
            />
          }
        </div>
      </div>
    </>
  )
}

function translateToNavigation(stackDefinition) {
  const navigation = stackDefinition.categories.map((category) => (
      {name: category.name, href: ref(category.name), category: category.id}
    ))
  navigation.unshift({name: "General", href: "/general"})
  return navigation
}

function ref(name) {
  return  "/" + name.replaceAll(" ", "-").toLowerCase()
}

const SkeletonLoader = () => {
  return (
    <>
      <div className="fixed w-full bg-neutral-100 dark:bg-neutral-900 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 pb-12 flex items-center">
          <h1 className="text-3xl leading-tight text-medium flex-grow py-0.5">Environment config</h1>
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700"></div>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex pt-56 animate-pulse">
        <div className="sticky h-96 top-56">
          <div className="w-56 p-4 pl-3 space-y-6">
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-3/5"></div>
            <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
          </div>
        </div>
        <div className="w-full ml-14">
          <div role="status" className="flex items-center justify-center h-72 bg-neutral-300 dark:bg-neutral-500 rounded-lg">
            <span className="sr-only">Loading...</span>
          </div>
        </div>
      </div>
    </>
  )
}

const findEnv = (envs, envName) => {
  if (envs.length === 0) {
    return undefined
  }

  return envs.find(env => env.name === envName)
};

const spinOutBuiltInEnv = (store, gimletClient) => {
  store.dispatch({
    type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
      header: "Converting environment..."
    }
  });
  gimletClient.spinOutBuiltInEnv()
    .then((data) => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
          header: "Success",
          message: "Environment was converted",
        }
      });
      store.dispatch({
        type: ACTION_TYPE_ENVSPINNEDOUT,
        payload: data
      });
    }, (err) => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
          header: "Error",
          message: err.statusText
        }
      });
      setTimeout(() => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWRESET
        });
      }, 3000);
    })
}

const PullRequests = ({title, items}) => {
  if (!items || items.length === 0) {
    return null
  }

  return (
    <div className="rounded-md bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300 p-4 w-full">
      <div className="flex">
        <div className="flex-shrink-0">
          <InformationCircleIcon className="h-5 w-5 " aria-hidden="true" />
        </div>
        <div className="ml-3 flex-1 md:flex md:justify-between">
          <div className="text-xs flex flex-col">
            <span className="font-semibold text-sm">{title}</span>
            <ul className="list-disc list-inside text-xs ml-2">
              {items.map (p => <li key={p.sha}>
                <a href={p.link} target="_blank" rel="noopener noreferrer">
                  {`#${p.number}`} {p.title}
                </a>
              </li>
              )}
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
