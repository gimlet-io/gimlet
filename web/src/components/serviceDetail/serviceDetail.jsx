import { useEffect, useState, useRef } from 'react';
import { RolloutHistory } from "../rolloutHistory/rolloutHistory";
import Emoji from "react-emoji-render";
import { Fragment } from 'react'
import { Menu, Transition } from '@headlessui/react'
import { ArrowPathIcon } from '@heroicons/react/20/solid'
import {
  ACTION_TYPE_ROLLOUT_HISTORY,
} from "../../redux/redux";
import { copyToClipboard } from '../../views/settings/settings';
import { usePostHog } from 'posthog-js/react'
import Timeline from './timeline';
import { AlertPanel } from './alert';
import { Logs } from '../../views/footer/logs';
import { Describe } from '../../views/footer/capacitor/Describe';
import { ArrowTopRightOnSquareIcon, LinkIcon } from '@heroicons/react/24/solid';
import { toast } from 'react-toastify';
import { InProgress, Success, Error } from '../../popUpWindow';

function ServiceDetail(props) {
  const { store, gimletClient } = props;
  const { owner, repoName } = props;
  const { environment } = props;
  const { stack, rolloutHistory, rollback, navigateToConfigEdit, linkToDeployment, config, fileName, releaseHistorySinceDays, deploymentFromParams, scmUrl, serviceAlerts } = props;
  const ref = useRef(null);
  const posthog = usePostHog()
  const [pullRequests, setPullRequests] = useState()

  const progressToastId = useRef(null);

  const configExists = config !== undefined

  useEffect(() => {
    if (deploymentFromParams === stack.service.name) {
      window.scrollTo({ behavior: 'smooth', top: ref.current.offsetTop })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deploymentFromParams, stack.service.name]);

  useEffect(() => {
    if (config) {
      gimletClient.getRolloutHistoryPerApp(owner, repoName, environment.name, config.app)
        .then(data => {
          store.dispatch({
            type: ACTION_TYPE_ROLLOUT_HISTORY, payload: {
              owner: owner,
              repo: repoName,
              env: environment.name,
              app: config.app,
              releases: data,
            }
          });
        }, () => {/* Generic error handler deals with it */ });
      
      gimletClient.getConfigChangePullRequestsPerConfig(owner, repoName, environment.name, config.app)
        .then(data => {
          setPullRequests(data)
        })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [isCopied, setCopied] = useState(false)

  const handleCopyClick = () => {
    setCopied(true);

    setTimeout(() => {
      setCopied(false);
    }, 2000);
  };

  const deleteAppInstance = () => {
    progressToastId.current = toast(<InProgress header="Deleting application instance..."/>, { autoClose: false });

    gimletClient.deleteAppInstance(environment.name, stack.service.name)
      .then(() => {
        toast.update(progressToastId.current, {
          render: <Success header="Application instance deleted"/>,
          className: "bg-green-50 shadow-lg p-2",
          bodyClassName: "p-2",
        });
      }, (err) => {
        toast.update(progressToastId.current, {
          render: <Error header="Error" message={err.statusText} />,
          className: "bg-red-50 shadow-lg p-2",
          bodyClassName: "p-2",
          progressClassName: "!bg-red-200",
          autoClose: 5000
        });
      });
  }

  const silenceAlert = (object, hours) => {
    var date = new Date();
    date.setHours(date.getHours() + hours);

    progressToastId.current = toast(<InProgress header="Silencing alerts..." />, { autoClose: false });

    gimletClient.silenceAlert(object, date.toISOString())
      .then(() => {
        toast.update(progressToastId.current, {
          render: <Success header="Silenced" />,
          className: "bg-green-50 shadow-lg p-2",
          bodyClassName: "p-2",
          autoClose: 3000,
        });
      }, (err) => {
        toast.update(progressToastId.current, {
          render: <Error header="Error" message={err.statusText} />,
          className: "bg-red-50 shadow-lg p-2",
          bodyClassName: "p-2",
          progressClassName: "!bg-red-200",
          autoClose: 5000
        });
      });
  }

  const deployment = stack.deployment;
  const repo = stack.repo;

  let hostPort = "<host-port>"
  let appPort = "<app-port>"
  if (config && config.values) {
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
    <div className="flex-1 text-neutral-900 dark:text-neutral-200">
      <h3 ref={ref} className="flex text-lg font-bold rounded">
        <span>{stack.service.name}</span>
        <button onClick={() => linkToDeployment(environment.name, stack.service.name)} target="_blank" rel="noopener noreferrer">
          <LinkIcon className="serviceLinkIcon ml-1" aria-hidden="true" />
        </button>
        {configExists &&
          <a href={`${scmUrl}/${owner}/${repoName}/blob/main/.gimlet/${encodeURIComponent(fileName)}`} target="_blank" rel="noopener noreferrer">
            <ArrowTopRightOnSquareIcon className="serviceLinkIcon ml-1" aria-hidden="true" />
          </a>
        }
        <div className="flex ml-auto space-x-1">
          {deployment &&
            <>
              <Logs
                capacitorClient={gimletClient}
                store={store}
                namespace={deployment.namespace}
                deployment={deployment.name}
                containers={podContainers(deployment.pods)}
              />
              <Describe
                capacitorClient={gimletClient}
                store={store}
                namespace={deployment.namespace}
                deployment={deployment.name}
                pods={deployment.pods}
              />
            </>
          }
          <Menu as="div" className="relative flex grow">
            {isCopied && (
              <div className="absolute -right-12 -top-10">
                <div className="text-sm font-medium p-2 bg-indigo-600 select-none text-white inline-block rounded">
                  Copied!
                </div>
              </div>
            )}
            <Menu.Button className="transparentBtn p-1.5">
              <span className="sr-only">Open options</span>
              <svg className="size-4" strokeLinejoin="round" viewBox="0 0 16 16"><path fillRule="evenodd" clip-rce="evenodd" d="M4 8C4 8.82843 3.32843 9.5 2.5 9.5C1.67157 9.5 1 8.82843 1 8C1 7.17157 1.67157 6.5 2.5 6.5C3.32843 6.5 4 7.17157 4 8ZM9.5 8C9.5 8.82843 8.82843 9.5 8 9.5C7.17157 9.5 6.5 8.82843 6.5 8C6.5 7.17157 7.17157 6.5 8 6.5C8.82843 6.5 9.5 7.17157 9.5 8ZM13.5 9.5C14.3284 9.5 15 8.82843 15 8C15 7.17157 14.3284 6.5 13.5 6.5C12.6716 6.5 12 7.17157 12 8C12 8.82843 12.6716 9.5 13.5 9.5Z" fill="currentColor"></path></svg>
            </Menu.Button>
            <Transition
              as={Fragment}
              enter="transition ease-out duration-100"
              enterFrom="transform opacity-0 scale-95"
              enterTo="transform opacity-100 scale-100"
              leave="transition ease-in duration-75"
              leaveFrom="transform opacity-100 scale-100"
              leaveTo="transform opacity-0 scale-95"
            >
              <Menu.Items className="absolute right-0 top-6 z-10 mt-2 w-56 origin-top-right rounded-md bg-neutral-800 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
                <div className="py-1">
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        className={`${configExists && active && 'bg-neutral-600'} ${configExists ? 'text-neutral-100' : 'text-neutral-500'} block px-4 py-2 text-sm w-full text-left font-normal`}
                        onClick={() => {
                          if (configExists) {
                            posthog?.capture('Env config edit pushed')
                            navigateToConfigEdit(environment.name, config.app)
                          }
                        }}
                      >
                        Edit{!configExists && ' (no config)'}
                      </button>
                    )}
                  </Menu.Item>
                  {deployment &&
                    <>
                      <Menu.Item>
                        {({ active }) => (
                          <button
                            className={`${active && 'bg-neutral-600'} text-neutral-100 block px-4 py-2 text-sm w-full text-left font-normal`}
                            onClick={() => {
                              // eslint-disable-next-line no-restricted-globals
                              confirm(`Are you sure you want to restart deployment ${deployment.name}?`) &&
                                gimletClient.restartDeploymentRequest(deployment.namespace, deployment.name)
                            }}
                          >
                            Restart
                          </button>
                        )}
                      </Menu.Item>
                      <Menu.Item>
                        {({ active }) => (
                          <button
                            className={`${active && 'bg-neutral-600'} text-neutral-100 block px-4 py-2 text-sm w-full text-left font-normal`}
                            onClick={() => {
                              // eslint-disable-next-line no-restricted-globals
                              confirm(`Are you sure you want to delete the ${stack.service.name} application instance?`) &&
                                deleteAppInstance()
                            }}
                          >
                            Delete deployed instance
                          </button>
                        )}
                      </Menu.Item>
                      {!environment.ephemeral &&
                        <Menu.Item>
                          {({ active }) => (
                            <button
                              className={`${config && active && 'bg-neutral-600'} ${config ? 'text-neutral-100' : 'text-neutral-500'} block px-4 py-2 text-sm w-full text-left font-normal`}
                              onClick={() => {
                                if (config) {
                                  copyToClipboard(`kubectl port-forward deploy/${deployment.name} -n ${deployment.namespace} ${hostPort}:${appPort}`);
                                  handleCopyClick();
                                }
                              }}
                            >
                              Copy port-forward command
                              {!config && ' (no config)'}
                            </button>
                          )}
                        </Menu.Item>
                      }
                    </>
                  }
                </div>
              </Menu.Items>
            </Transition>
          </Menu>
        </div>
      </h3>
      {deployment && config && <DeployIndicator deploy={config.values && config.values.deploy} owner={owner} repo={repoName} branch={deployment.branch} />}
      {pullRequests && pullRequests.length !== 0 &&
        <PullRequests items={pullRequests} />
      }
      <AlertPanel
        alerts={serviceAlerts?.filter(alert => alert.status === "Firing")}
        silenceAlert={silenceAlert}
        hideButton
      />
      <div className="grid grid-cols-12 py-4">
        <div className="col-span-5 border-r space-y-4">
          { deployment &&
          <>
          <div>
            <p className="serviceCardLabel">Pods</p>
            {
              deployment.pods && deployment.pods.map((pod) => (
                <Pod key={pod.name} pod={pod} />
              ))
            }
          </div>
          <div>
            <p className="serviceCardLabel">Version</p>
            <p>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512" className="h4 w-4 inline fill-current"><path d="M320 336a80 80 0 1 0 0-160 80 80 0 1 0 0 160zm156.8-48C462 361 397.4 416 320 416s-142-55-156.8-128H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H163.2C178 151 242.6 96 320 96s142 55 156.8 128H608c17.7 0 32 14.3 32 32s-14.3 32-32 32H476.8z"/></svg>
              <span className="text-xs pl-2 font-mono"><a href={`${scmUrl}/${repo}/commit/${deployment.sha}`} target="_blank" rel="noopener noreferrer">{deployment.sha.slice(0, 8)}</a></span>
              <span className="pl-2 text-sm font-normal">{deployment.commitMessage && <Emoji text={deployment.commitMessage} />}</span>
            </p>
          </div>
          </>
          }
          {stack.osca && Object.keys(stack.osca.links).length !== 0 &&
            <div>
              <p className="serviceCardLabel">Links</p>
              <div className="text-sm mt-2 flex">
                {
                  Object.keys(stack.osca.links).map((k, idx, ar) => {
                    return (
                      <div key={k}>
                        <a href={stack.osca.links[k]} rel="noreferrer" target="_blank" className="externalLink capitalize">
                          {k}
                          <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" />
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
            <p className="serviceCardLabel">Address</p>
            <div className="text-sm">
              <div className="relative">
              {stack.service.name}.{stack.service.namespace}.svc.cluster.local
              </div>
              {config && !environment.ephemeral &&
                <>
                  <a className="externalLink" href={'http://127.0.0.1:' + hostPort} target="_blank" rel="noopener noreferrer">http://127.0.0.1:{hostPort}
                    <ArrowTopRightOnSquareIcon className="externalLinkIcon mr-1" aria-hidden="true" />
                  </a>
                  (port-forward)
                </>
              }
              {stack.ingresses ? stack.ingresses.map((ingress) =>
                <p className="externalLink font-bold" key={`${ingress.namespace}/${ingress.name}`}>
                  <a href={'https://' + ingress.url} target="_blank" rel="noopener noreferrer">https://{ingress.url}
                  <ArrowTopRightOnSquareIcon className="externalLinkIcon ml-1" aria-hidden="true" />
                  </a>
                </p>
                ) : null
              }
            </div>
          </div>
          }
          { deployment &&
          <div>
            <p className="serviceCardLabel">Health</p>
            <div className="text-neutral-900 text-sm pb-2">
              <Timeline alerts={serviceAlerts} />
            </div>
          </div>
          }
          {config &&
          <div>
            <p className="serviceCardLabel">Deploy History</p>
            <div className="text-neutral-900 text-sm pt-2">
              <RolloutHistory
                env={environment.name}
                app={config.app}
                rollback={rollback}
                appRolloutHistory={rolloutHistory}
                releaseHistorySinceDays={releaseHistorySinceDays}
                scmUrl={scmUrl}
                builtInEnv={environment.builtIn}
              />
            </div>
          </div>
          }
        </div>
      </div>
    </div>
  )
}

export default ServiceDetail;

export function Pod(props) {
  const {pod} = props;

  let textColor;
  let color;
  let pulsar;
  switch (pod.status) {
    case 'Running':
      textColor = 'text-green-900 dark:text-teal-400'
      color = 'bg-green-300 dark:bg-teal-600';
      pulsar = '';
      break;
    case 'PodInitializing':
    case 'ContainerCreating':
    case 'Pending':
      textColor = 'text-blue-900'
      color = 'bg-blue-300';
      pulsar = 'animate-pulse';
      break;
    case 'Terminating':
      color = 'bg-neutral-500';
      pulsar = 'animate-pulse';
      break;
    default:
      textColor = 'text-neutral-900 dark:text-red-400'
      color = 'bg-red-600 dark:bg-red-800';
      pulsar = '';
      break;
  }

  return (
    <span className={`inline-block mr-1 mt-2 shadow-lg ${textColor} ${color} ${pulsar} font-bold px-2 cursor-default`} title={`${pod.name} - ${pod.status}`}>
      {pod.status}
    </span>
  );
}

export function podContainers(pods) {
  const containers = [];
  pods?.forEach((pod) => {
    pod.containers?.forEach(container => {
      containers.push(`${pod.name}/${container.name}`);
    })
  });

  return containers;
}

function DeployIndicator(props) {
  const { deploy, owner, repo, branch } = props;

  let indicator;
  switch (deploy && deploy.event) {
    case "push":
      indicator = <span><ArrowPathIcon className="h-4 mr-1 mt-0.5" aria-hidden="true" />Continuously deployed on {deploy.branch}</span>;
      break;
    case "tag":
      indicator = <span>Deployed on {deploy.tag} git tags </span>;
      break;
    default:
      indicator = <span><a href={`/repo/${owner}/${repo}/commits?branch=${branch}`}>Deploy manually</a></span>;
  }

  return <p className="align-top text-xs font-medium">{indicator}</p>
}

function PullRequests(props) {
  const { items } = props;

  const listItems = items.map(pullRequest =>
    <li key={pullRequest.sha}>
      <a href={pullRequest.link} target="_blank" rel="noopener noreferrer">
        {`#${pullRequest.number} ${pullRequest.title}`}
      </a>
    </li>
  );

  return (
    <div className="pt-2">
      <div className="bg-neutral-100 dark:bg-neutral-900 rounded w-full inline-grid items-center mx-auto p-3">
        <span className="font-medium text-xs">Pull Requests:</span>
        <ul className="list-disc list-inside text-xs ml-2">{listItems}</ul>
      </div>
    </div>
  )
};
