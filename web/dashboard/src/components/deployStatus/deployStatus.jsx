import SimpleServiceDetail from "../serviceDetail/simpleServiceDetail";
import { Modal } from "./modal";
import React, { useState, useRef, useEffect } from 'react';
import { ArrowUpIcon, ArrowDownIcon, PlayIcon } from '@heroicons/react/outline'

export function DeployStatusModal(props) {
  const { closeHandler, owner, repoName, scmUrl } = props
  const ownerAndRepo = `${owner}/${repoName}`
  const { store, gimletClient } = props
  const { envConfigs } = props
  const logsEndRef = useRef();
  const topRef = useRef();

  const [imageBuildLogs, setImageBuildLogs] = useState("");
  store.subscribe(() => {
    if (imageBuildTrackingId) {
      const logs = store.getState().imageBuildLogs[imageBuildTrackingId]
      setImageBuildLogs(logs)
    }
  });
  const [gitopsCommits, setGitopsCommits] = useState(store.getState().gitopsCommits);
  store.subscribe(() => setGitopsCommits(store.getState().gitopsCommits));
  const [connectedAgents, setConnectedAgents] = useState(store.getState().connectedAgents);
  store.subscribe(() => setConnectedAgents(store.getState().connectedAgents));
  const [envs, setEnvs] = useState(store.getState().envs);
  store.subscribe(() => setEnvs(store.getState().envs));
  const [rolloutHistory, setRolloutHistory] = useState(store.getState().rolloutHistory[ownerAndRepo]);
  store.subscribe(() => setRolloutHistory(store.getState().rolloutHistory[ownerAndRepo]));
  const [settings, setSettings] = useState(store.getState().settings);
  store.subscribe(() => setSettings(store.getState().settings));
  const [runningDeploy, setRunningDeploy] = useState(store.getState().runningDeploy);
  store.subscribe(() => {
    const r = store.getState().runningDeploy
    setRunningDeploy(r)
    if (r) {
      setReleaseTrackingId(r.trackingId)
    }
  });
  const [runningImageBuild, setRunningImageBuild] = useState(store.getState().runningImageBuild);
  store.subscribe(() => {
    const r = store.getState().runningImageBuild
    setRunningImageBuild(r)
    if (r) {
      setImageBuildTrackingId(r.trackingId)
    }
  });
  const [releaseTrackingId, setReleaseTrackingId] = useState(
    store.getState().runningDeploy
      ? store.getState().runningDeploy.trackingId
      : undefined
  );
  const [imageBuildTrackingId, setImageBuildTrackingId] = useState(
    store.getState().runningImageBuild
      ? store.getState().runningImageBuild.trackingId
      : undefined
  );
  const [followLogs, setFollowLogs] = useState(true);

  useEffect(() => {
    if (followLogs) {
      logsEndRef.current && logsEndRef.current.scrollIntoView()
    }
  }, [imageBuildTrackingId, followLogs]);
  useEffect(() => {
    if (followLogs) {
      logsEndRef.current && logsEndRef.current.scrollIntoView()
    }
  }, [releaseTrackingId, followLogs]);
  useEffect(() => {
    if (followLogs) {
      logsEndRef.current && logsEndRef.current.scrollIntoView()
    }
  }, [imageBuildLogs?.logLines?.length, followLogs]);

  if (!runningDeploy) {
    return (<Loading />)
  }

  const env = runningDeploy.env
  const app = runningDeploy.app
  const key = runningDeploy.trackingId

  let stack = connectedAgents[env].stacks.find(s => s.service.name === app)
  const config = envConfigs[env].find((config) => config.app === app)

  if (!stack) { // for apps we haven't deployed yet
    stack={service:{name: app}}
  }

  const stackRolloutHistory = rolloutHistory && rolloutHistory[env] ? rolloutHistory[env][stack.service.name] : undefined

  return (
    <Modal closeHandler={closeHandler} key={`modal-${key}`}>
      <div className="h-full flex flex-col">
        <SimpleServiceDetail
          stack={stack}
          rolloutHistory={stackRolloutHistory}
          envName={env}
          owner={owner}
          repoName={repoName}
          config={config}
          releaseHistorySinceDays={settings.releaseHistorySinceDays}
          gimletClient={gimletClient}
          store={store}
          scmUrl={scmUrl}
          builtInEnv={envs.find(e => e.name === env).builtIn}
          // serviceAlerts={alerts[deployment]}
          logsEndRef={logsEndRef}
        />
        <Controls topRef={topRef} logsEndRef={logsEndRef} followLogs={followLogs} setFollowLogs={setFollowLogs} />
        <div
          className="overflow-y-auto flex-grow min-h-[50vh] bg-stone-900 text-gray-300 font-mono text-sm p-2"
          onScroll={evt => {
              if ((logsEndRef.current.offsetTop-window.innerHeight-100) > evt.target.scrollTop) {
                setFollowLogs(false)
                console.log('not visible')
              }
            }}
          >
          <DeployStatusPanel
            key={`panel-${key}`}
            runningDeploy={runningDeploy}
            runningImageBuild={runningImageBuild}
            scmUrl={scmUrl}
            envs={envs}
            gitopsCommits={gitopsCommits}
            imageBuildLogs={imageBuildLogs}
            logsEndRef={logsEndRef}
            topRef={topRef}
          />
        </div>
      </div>
    </Modal>
  )
}

function Controls(props) {
  const {topRef, logsEndRef, followLogs, setFollowLogs} = props

  return (
    <div className="text-end">
      <span className="isolate inline-flex rounded-md shadow-sm">
        <button
          type="button"
          onClick={() => topRef.current.scrollIntoView()}
          className="relative inline-flex items-center rounded-l-md bg-white px-1 py-1 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-300 focus:z-10"
        >
          <ArrowUpIcon className="h-4 w-4" aria-hidden="true" />
        </button>
        <button
          type="button"
          onClick={() => {
            if (!followLogs) {
              logsEndRef.current.scrollIntoView()
            }
            setFollowLogs(!followLogs)
          }}
          className={`relative -ml-px inline-flex items-center px-1 py-1 text-gray-400 ring-1 ring-inset ring-gray-300 ${followLogs ? 'bg-gray-300' : 'bg-white' } hover:bg-gray-300 focus:z-10`}
        >
          <PlayIcon className="h-4 w-4" aria-hidden="true" />
        </button>
        <button
          type="button"
          onClick={() => logsEndRef.current.scrollIntoView()}
          className="relative -ml-px inline-flex items-center rounded-r-md bg-white px-1 py-1 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-10"
        >
          <ArrowDownIcon className="h-4 w-4" aria-hidden="true" />
        </button>
      </span>
    </div>
  )
}

export function DeployStatusPanel(props) {
  const { runningDeploy, runningImageBuild } = props
  const { scmUrl, envs, gitopsCommits, imageBuildLogs, logsEndRef, topRef } = props

  const deployStatusWidget = runningDeploy.trackingId ? DeployStatus({runningDeploy, scmUrl, gitopsCommits, envs}) : null
  const imageBuildWidget = runningImageBuild ? ImageBuild(runningImageBuild.trackingId, imageBuildLogs) : null

  const key = runningDeploy.trackingId+'-'+runningImageBuild?.trackingId

  return (
    <>
      <p ref={topRef} />
      <DeployHeader
        key={`header-${key}`}
        scmUrl={scmUrl}
        runningDeploy={runningDeploy}
      />
      {!deployStatusWidget && !imageBuildWidget ? (<Loading />) : null}
      {imageBuildWidget}
      {deployStatusWidget}
      <p className='pb-12' ref={logsEndRef} />
    </>
  )
}

export function DeployStatus(props) {
  const { runningDeploy, scmUrl, gitopsCommits, envs } = props

  const gitopsRepo = envs.find(env => env.name === runningDeploy.env).appsRepo;
  const builtInEnv = envs.find(env => env.name === runningDeploy.env).builtIn;

  let gitopsWidget = (
    <div className="">
      <Loading/>
    </div>
  )
  let appliedWidget = null;

  if (runningDeploy.status === 'error') {
    gitopsWidget = (
      <div className="pt-4">
        <p className="text-red-500 font-semibold">
          Error
        </p>
        <p className="text-red-500 font-base">
          {runningDeploy.statusDesc}
        </p>
      </div>
    )
  }

  const hasResults = runningDeploy.results && runningDeploy.results.length !== 0;
  if (runningDeploy.status === 'processed' || hasResults) {
    gitopsWidget = gitopsWidgetFromResults(runningDeploy, gitopsRepo, scmUrl, builtInEnv);
    appliedWidget = appliedWidgetFromResults(runningDeploy, gitopsCommits, gitopsRepo, scmUrl, builtInEnv);  
  }

  return (
    <div key={`gitops-${runningDeploy.trackingId}`}>
        <div className="text-gray-100">
          <div className="flex">
            <div className="w-0 flex-1 justify-between">
              {gitopsWidget}
              <div className='pl-2'>{appliedWidget}</div>
            </div>
          </div>
        </div>
    </div>
  );
}

export function DeployHeader(props) {
  const {scmUrl, runningDeploy} = props

  return (
    <div className='pb-4'>
      {!runningDeploy.rollback &&
      <p className="text-yellow-100 font-semibold">
        Rolling out {runningDeploy.app}
      </p>
      }
      {runningDeploy.rollback &&
      <p className="text-yellow-100 font-semibold">
        Rolling back {runningDeploy.app}
      </p>
      }
      <p className="pl-2  ">
        🎯 {runningDeploy.env}
      </p>
      {!runningDeploy.rollback &&
      <p className="pl-2">
        <span>📎</span>
        <a
          href={`${scmUrl}/${runningDeploy.repo}/commit/${runningDeploy.sha}`}
          target="_blank" rel="noopener noreferrer"
          className='ml-2'
        >
          {runningDeploy.sha.slice(0, 6)}
        </a>
      </p>
      }
    </div>
  );
}

export function ImageBuild(trackingId, build) {
  if (!build) {
    return null
  }

  let statusText = (
    <div className="w-4/5 font-mono text-xs">
      {build.logLines.join("")}
    </div>
  )
  let instructionsText = null;
  if (build.status === "notBuilt") {
    instructionsText = <p>We could not build an image automatically. Please check our <a className="font-bold underline" target="_blank" rel="noreferrer" href='https://gimlet.io/docs/container-image-building'>documentation</a> to proceed."</p>
  } else if (build.status === "error") {
    instructionsText = "Could not build image, check server logs."
  }

  return (
    <div key={`logs-${trackingId}`}>
      <p className="text-yellow-100 pb-2 font-semibold">
        Building image
      </p>
      <div className="">
        <div className="whitespace-pre-wrap">{statusText}</div>
        <div className="pt-2 text-orange-600">{instructionsText}</div>
      </div>
    </div>
  );
}


export function Loading() {
  return (
    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none"
      viewBox="0 0 24 24">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4}></circle>
      <path className="opacity-75" fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
    </svg>
  )
}

function renderAppliedWidget(deployCommit, gitopsRepo, scmUrl, builtInEnv) {
  if (!deployCommit.sha) {
    return null;
  }

  let color = "text-yellow-300";
  let deployCommitStatus = "trailing";
  let deployCommitStatusIcon = <span className="h-4 w-4 rounded-full relative top-1 inline-block bg-yellow-400" />;

  if (deployCommit.status.includes("NotReady")) {
    deployCommitStatus = "applying";
  } else if (deployCommit.status.includes("Succeeded")) {
    color = "text-green-300";
    deployCommitStatus = "applied";
    deployCommitStatusIcon = <span>✅</span>;
  } else if (deployCommit.status.includes("Failed")) {
    color = "text-red-500";
    deployCommitStatus = deployCommit.statusDesc;
    deployCommitStatusIcon = <span>❗</span>;
  }

  return (
    <p key={deployCommit.sha} className={`font-semibold ${color}`}>
      {deployCommitStatusIcon}
      { !builtInEnv &&
      <a
        href={`${scmUrl}/${gitopsRepo}/commit/${deployCommit.sha}`}
        target="_blank" rel="noopener noreferrer"
        className='ml-1'
      >
        {deployCommit.sha?.slice(0, 6)}
      </a>
      }
      { builtInEnv &&
      <span className='ml-1'>{deployCommit.sha?.slice(0, 6)}</span>
      }
      <span className='ml-1'>{deployCommitStatus}</span>
    </p>
  )
}

function renderResult(result, gitopsRepo, scmUrl, builtInEnv) {
  if (result.hash && result.status === "success") {
    return (
      <div className="pl-2" key={result.hash}>
        <p className="font-semibold truncate" title={result.app}>
          {result.app}
          <span className='mx-1 align-middle'>✅</span>
          { !builtInEnv &&
          <a
            href={`${scmUrl}/${gitopsRepo}/commit/${result.hash}`}
            target="_blank" rel="noopener noreferrer" className="font-normal"
          >
            {result.hash.slice(0, 6)}
          </a>
          }
          { builtInEnv &&
          <span className="font-normal">{result.hash.slice(0, 6)}</span>
          }
        </p>
      </div>)
  } else if (result.status === "failure") {
    return (
      <div className="pl-2">
        <div className="grid grid-cols-2">
          <div>
            <p className="font-semibold truncate" title={result.app}>{result.app}</p>
          </div>
          <span className='mx-1 align-middle'>❌</span>
        </div>
        <p className="break-words text-red-500 font-normal">{`${result.status}: ${result.statusDesc}`}</p>
      </div>
    )
  }
}


function gitopsWidgetFromResults(deploy, gitopsRepo, scmUrl, builtInEnv) {
  return (
    <div className="">
      <p className="text-yellow-100 pt-4 font-semibold">
        Manifests written to git
      </p>
      {deploy.results.map(result => renderResult(result, gitopsRepo, scmUrl, builtInEnv))}
    </div>
  )
}

function appliedWidgetFromResults(deploy, gitopsCommits, gitopsRepo, scmUrl, builtInEnv) {
  const firstCommitOfEnv = gitopsCommits.length > 0 ? gitopsCommits.find((gitopsCommit) => gitopsCommit.env === deploy.env) : {};

  let deployCommit = {};
  deploy.results.forEach(result => {
    if (result.hash === firstCommitOfEnv.sha) {
      deployCommit = Object.assign({}, firstCommitOfEnv);
    }
  })

  return renderAppliedWidget(deployCommit, gitopsRepo, scmUrl, builtInEnv);
}
