import React, { useState, useRef, useEffect } from 'react';
import { RolloutHistory } from "../rolloutHistory/rolloutHistory";
import Emoji from "react-emoji-render";
import { copyToClipboard } from '../../views/settings/settings';
import Timeline from './timeline';
import { Logs } from '../../views/footer/logs';
import { Describe } from '../../views/footer/capacitor/Describe';
import { Pod } from './serviceDetail'
import { ACTION_TYPE_ROLLOUT_HISTORY } from '../../redux/redux'
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

function SimpleServiceDetail(props) {
  const { store, gimletClient } = props
  const { owner, repoName, config, newApp } = props
  const { stack, envName, rolloutHistory, releaseHistorySinceDays, scmUrl, builtInEnv, serviceAlerts, logsEndRef } = props;
  const ref = useRef(null);

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
        logsEndRef.current.scrollIntoView({block: "nearest", inline: "nearest"});
      }, () => {/* Generic error handler deals with it */ });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [isCopied, setCopied] = useState(false)

  const handleCopyClick = () => {
    setCopied(true);

    setTimeout(() => {
      setCopied(false);
    }, 2000);
  };

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
      <div className="w-full flex items-center justify-between space-x-6 pb-2 rounded-lg">
        <div className="flex-1">
          <h3 ref={ref} className="flex text-lg font-bold rounded px-2 py-2">
            <span>{stack.service.name}</span>
            <div className="flex items-center ml-auto space-x-1">
              {deployment &&
                <>
                  <Logs
                    capacitorClient={gimletClient}
                    store={store}
                    namespace={deployment.namespace}
                    deployment={deployment.name}
                    pods={deployment.pods}
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
            </div>
          </h3>
          <div>
            <div className="grid grid-cols-12 px-2">
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
                    <div className="text-sm mt-2">
                      {
                        Object.keys(stack.osca.links).map((k, idx, ar) => {
                          return (
                            <div key={k}>
                              <a href={stack.osca.links[k]} rel="noreferrer" target="_blank" className="capitalize externalLink">
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
                    { !newApp &&
                    <button
                      onClick={() => {
                        copyToClipboard(`kubectl port-forward deploy/${deployment.name} -n ${deployment.namespace} ${hostPort}:${appPort}`);
                        handleCopyClick();
                      }}
                      className="absolute right-0 transparentBtn">
                      Port-forward command
                    </button>
                    }
                    {isCopied && (
                      <div className="absolute -right-5 -top-10">
                        <div className="p-2 bg-indigo-600 select-none text-white inline-block rounded">
                          Copied!
                        </div>
                      </div>
                    )}
                    </div>
                    {stack.ingresses ? stack.ingresses.map((ingress) =>
                      <p className="externalLink font-bold" key={`${ingress.namespace}/${ingress.name}`}>
                        <a href={'https://' + ingress.url} target="_blank" rel="noopener noreferrer">https://{ingress.url}
                        <ArrowTopRightOnSquareIcon className="externalLinkIcon ml-1" aria-hidden="true" />
                        </a>
                      </p>
                      ) : null
                    }
                    {config &&
                        <>
                          <a className="externalLink" href={'http://127.0.0.1:' + hostPort} target="_blank" rel="noopener noreferrer">http://127.0.0.1:{hostPort}
                          <ArrowTopRightOnSquareIcon className="externalLinkIcon mr-1" aria-hidden="true" />
                          </a>
                          (port-forward)
                        </>
                      }
                  </div>
                </div>
                }
                { deployment && !newApp &&
                <div>
                  <p className="serviceCardLabel">Health</p>
                  <div className="text-neutral-900 text-sm">
                    <Timeline alerts={serviceAlerts} />
                  </div>
                </div>
                }
                { !newApp &&
                <div>
                  <p className="text-base text-neutral-600">Deploy History</p>
                  <div className="text-neutral-900 text-sm pt-2">
                    <RolloutHistory
                      env={envName}
                      app={stack.service.name}
                      appRolloutHistory={rolloutHistory}
                      releaseHistorySinceDays={releaseHistorySinceDays}
                      scmUrl={scmUrl}
                      builtInEnv={builtInEnv}
                    />
                  </div>
                </div>
                }
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

export default SimpleServiceDetail;
