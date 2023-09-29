import React, { Component, useEffect, useState, useRef } from 'react';
import { RolloutHistory } from "../rolloutHistory/rolloutHistory";
import Emoji from "react-emoji-render";
import { XIcon } from '@heroicons/react/solid'
import {
  ACTION_TYPE_ROLLOUT_HISTORY,
  ACTION_TYPE_CLEAR_PODLOGS,
  ACTION_TYPE_CLEAR_DEPLOYMENT_DETAILS,
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
} from "../../redux/redux";
import { copyToClipboard } from '../../views/settings/settings';
import { Menu } from '@headlessui/react'
import { usePostHog } from 'posthog-js/react'
import Timeline from './timeline';
import { AlertPanel } from '../../views/pulse/pulse';

function ServiceDetail(props) {
  const { stack, rolloutHistory, rollback, envName, owner, repoName, navigateToConfigEdit, linkToDeployment, configExists, config, fileName, releaseHistorySinceDays, gimletClient, store, deploymentFromParams, scmUrl, builtInEnv, serviceAlerts } = props;
  const ref = useRef(null);
  const posthog = usePostHog()

  useEffect(() => {
    if (deploymentFromParams === stack.service.name) {
      window.scrollTo({ behavior: 'smooth', top: ref.current.offsetTop })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deploymentFromParams, stack.service.name]);

  useEffect(() => {
    gimletClient.getRolloutHistoryPerApp(owner, repoName, envName, stack.service.name)
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_ROLLOUT_HISTORY, payload: {
            owner: owner,
            repo: repoName,
            env: envName,
            app: stack.service.name,
            releases: data,
          }
        });
      }, () => {/* Generic error handler deals with it */ });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [logsOverlayVisible, setLogsOverlayVisible] = useState(false)
  const [logsOverlayNamespace, setLogsOverlayNamespace] = useState("")
  const [logsOverlayService, setLogsOverlayService] = useState("")

  const closeLogsOverlayHandler = (namespace, serviceName) => {
    setLogsOverlayVisible(false)
    gimletClient.stopPodlogsRequest(namespace, serviceName);
    store.dispatch({
      type: ACTION_TYPE_CLEAR_PODLOGS, payload: {
        pod: namespace + "/" + serviceName
      }
    });
  }

  const closeDetailsHandler = (namespace, serviceName) => {
    setLogsOverlayVisible(false)
    store.dispatch({
      type: ACTION_TYPE_CLEAR_DEPLOYMENT_DETAILS, payload: {
        deployment: namespace + "/" + serviceName
      }
    });
  }

  useEffect(() => {
    if (typeof window != 'undefined' && window.document) {
      if (logsOverlayVisible) {
        document.body.style.overflow = 'hidden';
        document.body.style.paddingRight = '15px';
      }
      return () => {
        document.body.style.overflow = 'unset';
        document.body.style.paddingRight = '0px';
      }
    }
  }, [logsOverlayVisible]);

  const deleteAppInstance = () => {
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Deleting application instance..."
      }
    });

    gimletClient.deleteAppInstance(envName, stack.service.name)
      .then(() => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Application instance deleted",
          }
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
      });
  }

  const deployment = stack.deployment;
  const repo = stack.repo;

  return (
    <>
      <LogsOverlay
        closeLogsOverlayHandler={closeLogsOverlayHandler}
        closeDetailsHandler={closeDetailsHandler}
        namespace={logsOverlayNamespace}
        svc={logsOverlayService}
        visible={logsOverlayVisible}
        store={store}
      />
      <div className="w-full flex items-center justify-between space-x-6 bg-stone-100 pb-6 rounded-lg">
        <div className="flex-1">
          <h3 ref={ref} className="flex text-lg font-bold rounded cursor-pointer p-4">
            <span onClick={() => linkToDeployment(envName, stack.service.name)}>{stack.service.name}</span>
            {configExists ?
              <>
              <a href={`${scmUrl}/${owner}/${repoName}/blob/main/.gimlet/${fileName}`} target="_blank" rel="noopener noreferrer">
              <svg xmlns="http://www.w3.org/2000/svg"
                className="inline fill-current text-gray-500 hover:text-gray-700 ml-1 h-4 w-4"
                viewBox="0 0 24 24">
                <path d="M0 0h24v24H0z" fill="none" />
                <path
                  d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
              </svg>
              </a>
              <div className="flex items-center ml-auto space-x-2">
                <button 
                  className="bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded"
                  onClick={() => {
                    posthog?.capture('Env config edit pushed')
                    navigateToConfigEdit(envName, stack.service.name)
                    }}
                  >
                  Edit
                </button>
                { deployment &&
                <>
                <button 
                  className="bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded"
                  >
                  Logs
                </button>
                <button className="bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded">
                  Describe
                </button>
                </>
                }
                { rolloutHistory.length != 0 &&
                <button 
                  className="inline-flex items-center bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 mr-1">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 15L3 9m0 0l6-6M3 9h12a6 6 0 010 12h-3" />
                    </svg>
                    <span>Instant rollback</span>
                </button>
                }
              </div>
              </>
              :
              <div className="flex items-center ml-auto">
                <svg xmlns="http://www.w3.org/2000/svg"
                  onClick={() => {
                    // eslint-disable-next-line no-restricted-globals
                    confirm(`Are you sure you want to delete the ${stack.service.name} application instance?`) &&
                    deleteAppInstance()
                  }}
                  className="items-center cursor-pointer inline text-red-400 hover:text-red-600 opacity-70 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </div>
              
            }
          </h3>
          {/* <AlertPanel alerts={serviceAlerts?.filter(alert => alert.status === "Firing")} hideButton /> */}
          <div>
            <div className="grid grid-cols-12 mt-4 px-4">
              <div className="col-span-5 border-r space-y-4">
                <div>
                  <p className="text-base text-gray-600">Status</p>
                  {
                    deployment && deployment.pods && deployment.pods.map((pod) => (
                      <Pod key={pod.name} pod={pod} />
                    ))
                  }
                </div>
                { deployment &&
                <div>
                  <p className="text-base text-gray-600">Version</p>
                  <p className="text-gray-900">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512" className="h4 w-4 inline fill-current"><path d="M320 336a80 80 0 1 0 0-160 80 80 0 1 0 0 160zm156.8-48C462 361 397.4 416 320 416s-142-55-156.8-128H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H163.2C178 151 242.6 96 320 96s142 55 156.8 128H608c17.7 0 32 14.3 32 32s-14.3 32-32 32H476.8z"/></svg>
                    <span className="text-xs pl-2 font-mono"><a href={`${scmUrl}/${repo}/commit/${deployment.sha}`} target="_blank" rel="noopener noreferrer">{deployment.sha.slice(0, 8)}</a></span>
                    <span className="pl-2 text-sm font-normal">{deployment.commitMessage && <Emoji text={deployment.commitMessage} />}</span>
                  </p>
                </div>
                }
              </div>
              <div className="col-span-7 space-y-4 pl-2">
                { deployment &&
                <div>
                  <p className="text-base text-gray-600">Address</p>
                  <div className="text-gray-900 text-sm">
                    <div className="relative">
                    {stack.service.name}.{stack.service.namespace}.svc.cluster.local
                    <button className="absolute right-0 bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded">
                      Port-forward command
                    </button>
                    </div>
                    {stack.ingresses ? stack.ingresses.map((ingress) =>
                      <p key={`${ingress.namespace}/${ingress.name}`}>
                        <a href={'https://' + ingress.url} target="_blank" rel="noopener noreferrer">{ingress.url}</a>
                      </p>
                      ) : null
                    }
                  </div>
                </div>
                }
                { deployment &&
                <div>
                  <p className="text-base text-gray-600">Health</p>
                  <div className="text-gray-900 text-sm">
                    <Timeline alerts={serviceAlerts} />
                  </div>
                </div>
                }
                <div>
                  <p className="text-base text-gray-600">Deployed</p>
                  <div className="text-gray-900 text-sm pt-2">
                    <RolloutHistory
                      env={envName}
                      app={stack.service.name}
                      rollback={rollback}
                      appRolloutHistory={rolloutHistory}
                      releaseHistorySinceDays={releaseHistorySinceDays}
                      scmUrl={scmUrl}
                      builtInEnv={builtInEnv}
                    />
                  </div>
                </div>
              </div>
            </div>
            {/* <p className="text-xs truncate w-9/12">{deployment.namespace}/{deployment.name}</p> */}
          </div>
          {/* <div className="flex flex-wrap text-sm">
            <div className="flex-1 min-w-full md:min-w-0">
              {stack.ingresses ? stack.ingresses.map((ingress) => <Ingress ingress={ingress} key={`${ingress.namespace}/${ingress.name}`} />) : null}
            </div>
            <div className="flex-1 md:ml-2 min-w-full md:min-w-0">
              <Deployment
                envName={envName}
                repo={stack.repo}
                deployment={stack.deployment}
                service={stack.service}
                gimletClient={gimletClient}
                config={config}
                setLogsOverlayVisible={setLogsOverlayVisible}
                setLogsOverlayNamespace={setLogsOverlayNamespace}
                setLogsOverlayService={setLogsOverlayService}
                scmUrl={scmUrl}
              />
            </div>
            <div className="flex-1 min-w-full md:min-w-0" />
          </div> */}
        </div>
      </div>
    </>
  )
}

class Ingress extends Component {
  render() {
    const { ingress } = this.props;

    if (ingress === undefined) {
      return null;
    }

    return (
      <div className="bg-gray-100 p-2 mb-1 border rounded-sm border-gray-200 text-gray-500 relative">
        <span className="text-xs text-gray-400 absolute bottom-0 right-0 p-2">ingress</span>
        <div className="mb-1 truncate "><a href={'https://' + ingress.url} target="_blank" rel="noopener noreferrer">{ingress.url}</a>
        </div>
        <p className="text-xs truncate mb-6">{ingress.namespace}/{ingress.name}</p>
      </div>
    );
  }
}

class Deployment extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isCopied: false,
    };
  }

  handleCopyClick() {
    this.setState({ isCopied: true });

    setTimeout(() => {
      this.setState({ isCopied: false });
    }, 2000);
  };

  render() {
    const { deployment, service, repo, gimletClient, config, setLogsOverlayVisible, setLogsOverlayNamespace, setLogsOverlayService, scmUrl } = this.props;

    if (deployment === undefined) {
      return null;
    }

    let hostPort = "<host-port>"
    let appPort = "<app-port>"
    if (config) {
      appPort = config.values.containerPort ?? 80;

      if (appPort < 99) {
        hostPort = "100" + appPort
      } else if (appPort < 999) {
        hostPort = "10" + appPort
      } else {
        hostPort = appPort
      }

      if (hostPort === "10080") { // Connections to HTTP, HTTPS or FTP servers on port 10080 will fail. This is a mitigation for the NAT Slipstream 2.0 attack.
        hostPort = "10081"
      }
    }

    return (
      <div className="grid grid-cols-10">
        <div className="col-span-9 bg-gray-100 p-2 mb-1 border rounded-sm border-blue-200, text-gray-500 relative">
          <span className="text-xs text-gray-400 absolute bottom-0 right-0 p-2">deployment</span>
          <div className="mb-1">
            <p className="truncate">{deployment.commitMessage && <Emoji text={deployment.commitMessage} />}</p>
            <p className="text-xs italic"><a href={`${scmUrl}/${repo}/commit/${deployment.sha}`} target="_blank"
              rel="noopener noreferrer">{deployment.sha.slice(0, 6)}</a></p>
          </div>
          <p className="text-xs truncate w-9/12">{deployment.namespace}/{deployment.name}</p>
          {
            deployment.pods && deployment.pods.map((pod) => (
              <Pod key={pod.name} pod={pod} />
            ))
          }
        </div>
        <div className="bg-slate-400 rounded-r-lg p-2 text-white text-xs space-y-2 mb-1 text-left relative w-10">
          <Menu as="div" className="relative inline-block text-left">
            <Menu.Button className="flex items-center text-gray-200 hover:text-gray-500">
              <span className="sr-only">Open options</span>
              <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </Menu.Button>
            <Menu.Items className="origin-top-right absolute right-0 md:left-8 md:right-0 md:-top-4 z-10 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
              <div className="py-1">
                <Menu.Item key="logs">
                  {({ active }) => (
                    <button
                      onClick={() => {
                        setLogsOverlayVisible(true)
                        setLogsOverlayNamespace(deployment.namespace);
                        setLogsOverlayService(service.name);
                        gimletClient.podLogsRequest(deployment.namespace, service.name);
                      }}
                      className={(
                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700') +
                        ' block px-4 py-2 text-sm w-full text-left'
                      }
                    >
                      View app logs
                    </button>
                  )}
                </Menu.Item>
                <Menu.Item>
                  {({ active }) => (
                    <button
                      onClick={() => {
                        setLogsOverlayVisible(true);
                        setLogsOverlayNamespace(deployment.namespace);
                        setLogsOverlayService(service.name);
                        gimletClient.deploymentDetailsRequest(deployment.namespace, service.name);
                      }}
                      className={(
                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700') +
                        ' block px-4 py-2 text-sm w-full text-left'
                      }
                    >
                      View deployment details
                    </button>
                  )}
                </Menu.Item>
                <Menu.Item>
                  {({ active }) => (
                    <button
                      onClick={() => {
                        copyToClipboard(`kubectl port-forward deploy/${deployment.name} -n ${deployment.namespace} ${hostPort}:${appPort}`);
                        this.handleCopyClick();
                      }}
                      className={(
                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700') +
                        ' block px-4 py-2 text-sm w-full text-left'
                      }
                    >
                      Copy kubectl port-forward command
                    </button>
                  )}
                </Menu.Item>
              </div>
            </Menu.Items>
          </Menu>
          {this.state.isCopied && (
            <div className="absolute -top-8 right-1/2">
              <div className="p-2 bg-indigo-600 select-none text-white inline-block rounded">
                Copied!
              </div>
            </div>
          )}
        </div>
      </div>
    );
  }
}

export default ServiceDetail;

const LogsOverlay = ({ visible, namespace, svc, closeLogsOverlayHandler, closeDetailsHandler, store }) => {
  let reduxState = store.getState();
  const service = namespace + "/" + svc;

  const [logs, setLogs] = useState(reduxState.podLogs[service])
  const [details, setDetails] = useState(reduxState.deploymentDetails[service])

  store.subscribe(() => {
    setLogs(reduxState.podLogs[service])
    setDetails(reduxState.deploymentDetails[service])
  });

  const logsEndRef = useRef(null);

  useEffect(() => {
    logsEndRef.current.scrollIntoView();
  }, [logs, details]);

  const handleClose = () => {
    if (details) {
      closeDetailsHandler(namespace, svc);
    } else {
      closeLogsOverlayHandler(namespace, svc);
    }
  };

  return (
    <div
      className={(visible ? "visible" : "invisible") + " fixed flex inset-0 z-10 bg-gray-500 bg-opacity-75"}
      onClick={handleClose}
    >
      <div className="flex self-center items-center justify-center w-full p-8 h-4/5">
        <div className="transform flex flex-col overflow-hidden bg-white rounded-xl h-4/5 max-h-full w-4/5 p-6"
          onClick={e => e.stopPropagation()}
        >
          <div className="absolute top-0 right-0 p-1.5">
            <button
              className="rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none"
              onClick={handleClose}
            >
              <span className="sr-only">Close</span>
              <XIcon className="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div className="h-full relative overflow-y-auto p-4 bg-slate-800 rounded-lg">
            {logs?.map((line, idx) => <p key={idx} className={`font-mono text-xs ${line.color}`}>{line.content}</p>)}
            {details?.map((line, idx) => <p key={idx} className="font-mono text-xs text-yellow-200 whitespace-pre">{line}</p>)}
            <p ref={logsEndRef} />
          </div>
        </div>
      </div>
    </div>
  )
}

function Pod(props) {
  const {pod} = props;

  let color;
  let pulsar;
  switch (pod.status) {
    case 'Running':
      color = 'text-blue-200';
      pulsar = '';
      break;
    case 'PodInitializing':
    case 'ContainerCreating':
    case 'Pending':
      color = 'text-blue-100';
      pulsar = 'pulsar-green';
      break;
    case 'Terminating':
      color = 'text-blue-800';
      pulsar = 'pulsar-gray';
      break;
    default:
      color = 'text-red-600';
      pulsar = '';
      break;
  }

  return (
    <span className="inline-block w-4 mr-1 mt-2">
      <svg viewBox="0 0 1 1"
           className={`fill-current ${color} ${pulsar}`}>
        <g>
          <title>{pod.name} - {pod.status}</title>
          <rect width="1" height="1"/>
        </g>
      </svg>
    </span>
  );
}
