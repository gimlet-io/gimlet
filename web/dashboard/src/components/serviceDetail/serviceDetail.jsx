import React, { useEffect, useState, useRef } from 'react';
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

  const [isCopied, setCopied] = useState(false)

  const handleCopyClick = () => {
    setCopied(true);

    setTimeout(() => {
      setCopied(false);
    }, 2000);
  };

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

  const silenceAlert = (object, hours) => {
    var date = new Date();
    date.setHours(date.getHours() + hours);

    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Silence deployment alerts..."
      }
    });

    gimletClient.silenceAlert(object, date.toISOString())
      .then(() => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
          }
        });
        setTimeout(() => {
          store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWRESET
          });
        }, 3000);
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
    <>
      <LogsOverlay
        closeLogsOverlayHandler={closeLogsOverlayHandler}
        closeDetailsHandler={closeDetailsHandler}
        namespace={logsOverlayNamespace}
        svc={logsOverlayService}
        visible={logsOverlayVisible}
        store={store}
      />
      <div className="w-full flex items-center justify-between space-x-6 bg-stone-100 pb-8 rounded-lg">
        <div className="flex-1">
          <h3 ref={ref} className="flex text-lg font-bold rounded p-4">
            <span className="cursor-pointer" onClick={() => linkToDeployment(envName, stack.service.name)}>{stack.service.name}</span>
            {configExists &&
              <a href={`${scmUrl}/${owner}/${repoName}/blob/main/.gimlet/${encodeURIComponent(fileName)}`} target="_blank" rel="noopener noreferrer">
                <svg xmlns="http://www.w3.org/2000/svg"
                  className="inline fill-current text-gray-500 hover:text-gray-700 ml-1 h-4 w-4"
                  viewBox="0 0 24 24">
                  <path d="M0 0h24v24H0z" fill="none" />
                  <path
                    d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                </svg>
              </a>
            }
            <div className="flex items-center ml-auto space-x-2">
              {configExists &&
                <button
                  className="bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded"
                  onClick={() => {
                    posthog?.capture('Env config edit pushed')
                    navigateToConfigEdit(envName, stack.service.name)
                  }}
                >
                  Edit
                </button>
              }
              {deployment &&
                <>
                  <button
                    onClick={() => {
                      // eslint-disable-next-line no-restricted-globals
                      confirm(`Are you sure you want to restart deployment ${deployment.name}?`) &&
                        gimletClient.restartDeploymentRequest(deployment.namespace, deployment.name)
                    }}
                    className="bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded">
                    Restart
                  </button>
                  <button
                    onClick={() => {
                      setLogsOverlayVisible(true)
                      setLogsOverlayNamespace(deployment.namespace);
                      setLogsOverlayService(stack.service.name);
                      gimletClient.podLogsRequest(deployment.namespace, stack.service.name);
                    }}
                    className="bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded"
                  >
                    Logs
                  </button>
                  <button
                    onClick={() => {
                      setLogsOverlayVisible(true);
                      setLogsOverlayNamespace(deployment.namespace);
                      setLogsOverlayService(stack.service.name);
                      gimletClient.deploymentDetailsRequest(deployment.namespace, stack.service.name);
                    }}
                    className="bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded">
                    Describe
                  </button>
                  {!configExists &&
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
                </>
              }
            </div>
          </h3>
          <AlertPanel
            alerts={serviceAlerts?.filter(alert => alert.status === "Firing")}
            silenceAlert={silenceAlert}
            hideButton
          />
          <div>
            <div className="grid grid-cols-12 mt-4 px-4">
              <div className="col-span-5 border-r space-y-4">
                { deployment &&
                <>
                <div>
                  <p className="text-base text-gray-600">Status</p>
                  {
                    deployment.pods && deployment.pods.map((pod) => (
                      <Pod key={pod.name} pod={pod} />
                    ))
                  }
                </div>
                <div>
                  <p className="text-base text-gray-600">Version</p>
                  <p className="text-gray-900">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512" className="h4 w-4 inline fill-current"><path d="M320 336a80 80 0 1 0 0-160 80 80 0 1 0 0 160zm156.8-48C462 361 397.4 416 320 416s-142-55-156.8-128H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H163.2C178 151 242.6 96 320 96s142 55 156.8 128H608c17.7 0 32 14.3 32 32s-14.3 32-32 32H476.8z"/></svg>
                    <span className="text-xs pl-2 font-mono"><a href={`${scmUrl}/${repo}/commit/${deployment.sha}`} target="_blank" rel="noopener noreferrer">{deployment.sha.slice(0, 8)}</a></span>
                    <span className="pl-2 text-sm font-normal">{deployment.commitMessage && <Emoji text={deployment.commitMessage} />}</span>
                  </p>
                </div>
                </>
                }
                {stack.osca && Object.keys(stack.osca.links).length !== 0 &&
                  <div>
                    <p className="text-base text-gray-600">Links</p>
                    <div className="text-gray-700 text-sm mt-2">
                      {
                        Object.keys(stack.osca.links).map((k, idx, ar) => {
                          return (
                            <div key={k}>
                              <a href={stack.osca.links[k]} rel="noreferrer" target="_blank" className="capitalize">
                                {k}
                                <svg xmlns="http://www.w3.org/2000/svg"
                                  className="inline fill-current text-gray-500 hover:text-gray-700 h-4 w-4"
                                  viewBox="0 0 24 24">
                                  <path d="M0 0h24v24H0z" fill="none" />
                                  <path
                                    d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                                </svg>
                              </a>
                              {idx !== ar.length - 1 && <span className="px-2">|</span>}
                            </div>
                          )
                        })
                      }
                    </div>
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
                    <button
                      onClick={() => {
                        copyToClipboard(`kubectl port-forward deploy/${deployment.name} -n ${deployment.namespace} ${hostPort}:${appPort}`);
                        handleCopyClick();
                      }}
                      className="absolute right-0 bg-transparent hover:bg-slate-100 font-medium text-sm text-gray-700 py-1 px-4 border border-gray-300 rounded">
                      Port-forward command
                    </button>
                    {isCopied && (
                      <div className="absolute -right-12 -top-10">
                        <div className="p-2 bg-indigo-600 select-none text-white inline-block rounded">
                          Copied!
                        </div>
                      </div>
                    )}
                    </div>
                    {stack.ingresses ? stack.ingresses.map((ingress) =>
                      <p key={`${ingress.namespace}/${ingress.name}`}>
                        <a href={'https://' + ingress.url} target="_blank" rel="noopener noreferrer">https://{ingress.url}
                        <svg xmlns="http://www.w3.org/2000/svg"
                          className="inline fill-current text-gray-500 hover:text-gray-700 ml-1 h-4 w-4"
                          viewBox="0 0 24 24">
                          <path d="M0 0h24v24H0z" fill="none" />
                          <path
                            d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                        </svg>
                        </a>
                      </p>
                      ) : null
                    }
                    {config &&
                        <>
                          <a href={'http://127.0.0.1:' + hostPort} target="_blank" rel="noopener noreferrer">http://127.0.0.1:{hostPort}
                            <svg xmlns="http://www.w3.org/2000/svg"
                              className="inline fill-current text-gray-500 hover:text-gray-700 mr-1 h-4 w-4"
                              viewBox="0 0 24 24">
                              <path d="M0 0h24v24H0z" fill="none" />
                              <path
                                d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                            </svg>
                          </a>
                          (port-forward)
                        </>
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
                  <p className="text-base text-gray-600">Deploy History</p>
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
          </div>
        </div>
      </div>
    </>
  )
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
      color = 'bg-green-200';
      pulsar = '';
      break;
    case 'PodInitializing':
    case 'ContainerCreating':
    case 'Pending':
      color = 'bg-blue-300';
      pulsar = 'animate-pulse';
      break;
    case 'Terminating':
      color = 'bg-gray-500';
      pulsar = 'animate-pulse';
      break;
    default:
      color = 'bg-red-600';
      pulsar = '';
      break;
  }

  return (
    <span className={`inline-block mr-1 mt-2 shadow-lg ${color} ${pulsar} font-bold px-2 cursor-default`} title={`${pod.name} - ${pod.status}`}>
      {pod.status}
    </span>
  );
}
