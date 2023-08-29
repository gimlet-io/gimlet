export function DeployStatus(
  deploy,
  scmUrl,
  gitopsCommits,
  envs
  ) {

  const gitopsRepo = envs.find(env => env.name === deploy.env).appsRepo;
  const builtInEnv = envs.find(env => env.name === deploy.env).builtIn;

  let gitopsWidget = (
    <div className="">
      <Loading/>
    </div>
  )
  let appliedWidget = null;

  if (deploy.status === 'error') {
    gitopsWidget = (
      <div className="">
        <p className="text-red-500 font-semibold">
          Gitops write failed
        </p>
        <p className="text-red-500 font-base">
          {deploy.statusDesc}
        </p>
      </div>
    )
  }

  const hasResults = deploy.results && deploy.results.length !== 0;
  if (deploy.status === 'processed' || hasResults) {
    gitopsWidget = gitopsWidgetFromResults(deploy, gitopsRepo, scmUrl, builtInEnv);
    appliedWidget = appliedWidgetFromResults(deploy, gitopsCommits, deploy.env, gitopsRepo, scmUrl, builtInEnv);  
  }

  return (
    <div className="">
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

export function deployHeader(scmUrl, deploy) {
  return (
    <>
      {!deploy.rollback &&
      <p className="text-yellow-100 font-semibold">
        Rolling out {deploy.app}
      </p>
      }
      {deploy.rollback &&
      <p className="text-yellow-100 font-semibold">
        Rolling back {deploy.app}
      </p>
      }
      <p className="pl-2  ">
        🎯 {deploy.env}
      </p>
      {!deploy.rollback &&
      <p className="pl-2">
        <span>📎</span>
        <a
          href={`${scmUrl}/${deploy.repo}/commit/${deploy.sha}`}
          target="_blank" rel="noopener noreferrer"
          className='ml-1'
        >
          {deploy.sha.slice(0, 6)}
        </a>
      </p>
      }
    </>
  );
}

export function ImageBuild(build, logsEndRef) {
  let statusIcon = '⏳';
  let statusText = (
    <div className="w-4/5 font-mono text-xs">
      {build.logLines.join()}
    </div>
  )
  let instructionsText = null;
  if (build.status === "success") {
    statusIcon = '✅';
  } else if (build.status === "notBuilt") {
    statusIcon = '😟';
    instructionsText = <p>We could not build an image automatically. Please check our <a className="font-bold underline" target="_blank" rel="noreferrer" href='https://gimlet.io/docs/container-image-building'>documentation</a> to proceed."</p>
  } else if (build.status === "error") {
    statusIcon = '❗';
    statusText = "Could not build image, check server logs."
  }

  return (
    <>
      <p className="text-yellow-100 pt-4 pb-2 font-semibold">
        Building image {statusIcon}
      </p>
      <div className="px-2">
      <div className="p-4 h-48 overflow-y-scroll bg-gray-900">
        <div className="whitespace-pre-wrap">{statusText}</div>
        <p ref={logsEndRef} />
      </div>
      <div className="pt-2 text-orange-600">{instructionsText}</div>
      </div>
    </>
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

function appliedWidgetFromResults(deploy, gitopsCommits, env, gitopsRepo, scmUrl, builtInEnv) {
  const firstCommitOfEnv = gitopsCommits.length > 0 ? gitopsCommits.find((gitopsCommit) => gitopsCommit.env === env) : {};

  let deployCommit = {};
  deploy.results.forEach(result => {
    if (result.hash === firstCommitOfEnv.sha) {
      deployCommit = Object.assign({}, firstCommitOfEnv);
    }
  })

  return renderAppliedWidget(deployCommit, gitopsRepo, scmUrl, builtInEnv);
}
