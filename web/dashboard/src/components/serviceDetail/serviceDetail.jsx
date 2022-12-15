import React, { Component, useEffect, useState, useRef } from 'react';
import { RolloutHistory } from "../rolloutHistory/rolloutHistory";
import Emoji from "react-emoji-render";
import { XIcon } from '@heroicons/react/solid'
import { KubernetesEventsAlertBox } from '../../views/pulse/pulse';
import {
  ACTION_TYPE_ROLLOUT_HISTORY,
  ACTION_TYPE_CLEAR_PODLOGS
} from "../../redux/redux";


function ServiceDetail(props) {
  const { stack, rolloutHistory, rollback, envName, owner, repoName, navigateToConfigEdit, linkToDeployment, configExists, fileName, releaseHistorySinceDays, gimletClient, store, kubernetesEvents, deploymentFromParams } = props;
  const ref = useRef(null);

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

  return (
    <>
      <PodLogsOverlay
        closeLogsOverlayHandler={closeLogsOverlayHandler}
        namespace={logsOverlayNamespace}
        svc={logsOverlayService}
        visible={logsOverlayVisible}
        store={store}
      />
      <div className="w-full flex items-center justify-between space-x-6">
        <div className="flex-1">
          <h3 ref={ref} className="flex text-lg font-bold">
            {stack.service.name}
            {configExists &&
              <>
                <a href={`https://github.com/${owner}/${repoName}/blob/main/.gimlet/${fileName}`} target="_blank" rel="noopener noreferrer">
                  <svg xmlns="http://www.w3.org/2000/svg"
                    className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="16" height="16"
                    viewBox="0 0 24 24">
                    <path d="M0 0h24v24H0z" fill="none" />
                    <path
                      d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                  </svg>
                </a>
                <span onClick={() => linkToDeployment(envName, stack.service.name)}>
                  <svg
                    className="cursor-pointer inline text-gray-500 hover:text-gray-700 ml-1 h-5 w-5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1">
                    </path>
                  </svg>
                </span>
                <span onClick={() => navigateToConfigEdit(envName, stack.service.name)}>
                  <svg
                    className="cursor-pointer inline text-gray-500 hover:text-gray-700 ml-1  h-5 w-5"
                    xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </span>
              </>
            }
          </h3>
          {<div className="px-3 py-4">
            <KubernetesEventsAlertBox
              kubernetesEvents={kubernetesEvents}
              hideButton
            />
          </div>}
          <div className="my-2 mb-4 sm:my-4 sm:mb-6">
            <RolloutHistory
              env={envName}
              app={stack.service.name}
              rollback={rollback}
              appRolloutHistory={rolloutHistory}
              releaseHistorySinceDays={releaseHistorySinceDays}
            />
          </div>
          <div className="flex flex-wrap text-sm">
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
                setLogsOverlayVisible={setLogsOverlayVisible}
                setLogsOverlayNamespace={setLogsOverlayNamespace}
                setLogsOverlayService={setLogsOverlayService}
              />
            </div>
            <div className="flex-1 min-w-full md:min-w-0" />
          </div>
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
  render() {
    const { deployment, service, repo, gimletClient, setLogsOverlayVisible, setLogsOverlayNamespace, setLogsOverlayService } = this.props;

    if (deployment === undefined) {
      return null;
    }

    return (
      <div className="bg-gray-100 p-2 mb-1 border rounded-sm border-blue-200, text-gray-500 relative">
        <button className="text-xs text-gray-400 absolute top-0 right-0 p-2"
          onClick={() => {
            setLogsOverlayVisible(true)
            setLogsOverlayNamespace(deployment.namespace);
            setLogsOverlayService(service.name);
            gimletClient.podLogsRequest(deployment.namespace, service.name);
          }}
        >app logs</button>
        <span className="text-xs text-gray-400 absolute bottom-0 right-0 p-2">deployment</span>
        <div className="mb-1">
          <p className="truncate">{deployment.commitMessage && <Emoji text={deployment.commitMessage} />}</p>
          <p className="text-xs italic"><a href={`https://github.com/${repo}/commit/${deployment.sha}`} target="_blank"
            rel="noopener noreferrer">{deployment.sha.slice(0, 6)}</a></p>
        </div>
        <p className="text-xs truncate w-9/12">{deployment.namespace}/{deployment.name}</p>
        {
          deployment.pods && deployment.pods.map((pod) => (
            <Pod key={pod.name} pod={pod} />
          ))
        }
      </div>
    );
  }
}

export default ServiceDetail;

const PodLogsOverlay = ({ visible, namespace, svc, closeLogsOverlayHandler, store }) => {
  let reduxState = store.getState();
  const pod = namespace + "/" + svc;

  const [logs, setLogs] = useState(reduxState.podLogs[pod])

  store.subscribe(() => {
    setLogs(reduxState.podLogs[pod])
  });

  const logsEndRef = useRef(null);

  useEffect(() => {
    logsEndRef.current.scrollIntoView();
  }, [logs]);

  return (
    <div
      className={(visible ? "visible" : "invisible") + " fixed flex inset-0 z-10 bg-gray-500 bg-opacity-75"}
      onClick={() => { closeLogsOverlayHandler(namespace, svc) }}
    >
      <div className="flex self-center items-center justify-center w-full p-8 h-4/5">
        <div className="transform flex flex-col overflow-hidden bg-white rounded-xl h-4/5 max-h-full w-4/5 p-6"
          onClick={e => { e.stopPropagation() }}
        >
          <div className="absolute top-0 right-0 p-1.5">
            <button
              className="rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none"
              onClick={() => {
                closeLogsOverlayHandler(namespace, svc);
              }}
            >
              <span className="sr-only">Close</span>
              <XIcon className="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div className="h-full relative overflow-y-auto p-4 bg-slate-800 rounded-lg">
            {logs?.split('\n').map((line, idx) => <p key={idx} className='font-mono text-xs text-yellow-200'>{line}</p>)}
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
