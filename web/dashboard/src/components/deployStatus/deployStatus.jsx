import {Component, Fragment} from 'react'
import {Transition} from '@headlessui/react'
import {XIcon} from '@heroicons/react/solid'
import {ACTION_TYPE_CLEAR_DEPLOY_STATUS} from "../../redux/redux";

export default class DeployStatus extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      runningDeploys: reduxState.runningDeploys,
      envs: reduxState.envs,
      gitopsCommits: reduxState.gitopsCommits
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ runningDeploys: reduxState.runningDeploys });
      this.setState({ envs: reduxState.envs });
      this.setState({ gitopsCommits: reduxState.gitopsCommits });
    });
  }

  render() {
    const {runningDeploys, envs} = this.state;

    if (runningDeploys.length === 0) {
      return null;
    }

    const deploy = runningDeploys[0];
    const gitopsRepo = envs.find(env => env.name === deploy.env).appsRepo;

    console.log("RENDER")
    console.log(envs)
    console.log(gitopsRepo)

    let gitopsWidget = (
      <div className="mt-2">
        <Loading/>
      </div>
    )
    let appliedWidget = null;

    if (deploy.status === 'error') {
      gitopsWidget = (
        <div className="mt-2">
          <p className="text-red-500 font-semibold">
            Gitops write failed
          </p>
          <p className="text-red-500 font-base">
            {deploy.statusDesc}
          </p>
        </div>
      )
    }

    // Feature 'Add results to deploy status', gitopsHashes is deprecated, will be removed!
    if (deploy.results) {
      gitopsWidget = gitopsWidgetFromResults(deploy, gitopsRepo);
      appliedWidget = appliedWidgetFromResults(deploy, this.state.gitopsCommits, deploy.env, gitopsRepo);
    } else {
      const hasGitopsHashes = deploy.gitopsHashes && deploy.gitopsHashes.length !== 0;
      if (deploy.status === 'processed' || hasGitopsHashes) {
        gitopsWidget = gitopsWidgetFromGitopsHashes(deploy, gitopsRepo);
        appliedWidget = appliedWidgetFromGitopsHashes(deploy, this.state.gitopsCommits, deploy.env, gitopsRepo);
      }
    }

    return (
      <>
        <div
          aria-live="assertive"
          className="fixed inset-0 flex items-end px-4 py-6 pointer-events-none sm:p-6 sm:items-start"
        >
          <div className="w-full flex flex-col items-center space-y-4 sm:items-end">
            <Transition
              show={runningDeploys.length > 0}
              as={Fragment}
              enter="transform ease-out duration-300 transition"
              enterFrom="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
              enterTo="translate-y-0 opacity-100 sm:translate-x-0"
              leave="transition ease-in duration-100"
              leaveFrom="opacity-100"
              leaveTo="opacity-0"
            >
              <div
                className="max-w-lg w-full bg-gray-800 text-gray-100 text-sm shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5 overflow-hidden">
                <div className="p-4">
                  <div className="flex">
                    <div className="w-0 flex-1 justify-between">
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
                          href={`https://github.com/${deploy.repo}/commit/${deploy.sha}`}
                          target="_blank" rel="noopener noreferrer"
                          className='ml-1'
                        >
                          {deploy.sha.slice(0, 6)}
                        </a>
                      </p>
                      }
                      {gitopsWidget}
                      <div className='pl-2 mt-4'>{appliedWidget}</div>
                    </div>
                    <div className="ml-4 flex-shrink-0 flex items-start">
                      <button
                        className="rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none"
                        onClick={() => {
                          this.props.store.dispatch({
                            type: ACTION_TYPE_CLEAR_DEPLOY_STATUS, payload: {}
                          });
                        }}
                      >
                        <span className="sr-only">Close</span>
                        <XIcon className="h-5 w-5" aria-hidden="true"/>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </>
    )
  }
}

function Loading() {
  return (
    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none"
      viewBox="0 0 24 24">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4}></circle>
      <path className="opacity-75" fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
    </svg>
  )
}

function renderAppliedWidget(deployCommit, gitopsRepo) {
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
      <a
        href={`https://github.com/${gitopsRepo}/commit/${deployCommit.sha}`}
        target="_blank" rel="noopener noreferrer"
        className='ml-1'
      >
        {deployCommit.sha?.slice(0, 6)}
      </a>
      <span className='ml-1'>{deployCommitStatus}</span>
    </p>
  )
}

function renderResult(result, gitopsRepo) {
  if (result.hash && result.status === "success") {
    return (
      <div className="pl-2 mb-2">
        <p className="font-semibold truncate mb-1" title={result.app}>
          {result.app}
          <span className='mx-1 align-middle'>✅</span>
          <a
            href={`https://github.com/${gitopsRepo}/commit/${result.hash}`}
            target="_blank" rel="noopener noreferrer"
          >
            {result.hash.slice(0, 6)}
          </a>
        </p>
      </div>)
  } else if (result.status === "failure") {
    return (
      <div className="pl-2 mb-2">
        <div className="grid grid-cols-2">
          <div>
            <p className="font-semibold truncate mb-1" title={result.app}>{result.app}</p>
          </div>
          <span className='mx-1 align-middle'>❌</span>
        </div>
        <p className="break-words text-red-500 font-semibold">{`${result.status}: ${result.statusDesc}`}</p>
      </div>
    )
  }
}


function gitopsWidgetFromResults(deploy, gitopsRepo) {
  if (deploy.status !== 'processed') {
    return (
      <div className="mt-2">
        <Loading />
      </div>);
  }

  return (
    <div className="mt-2">
      <p className="text-yellow-100 font-semibold">
        Manifests written to git
      </p>
      {deploy.results.map(result => renderResult(result, gitopsRepo))}
    </div>
  )
}

function gitopsWidgetFromGitopsHashes(deploy, gitopsRepo) {
  return (
    <div className="mt-2">
      <p className="text-yellow-100 font-semibold">
        Manifests written to git
      </p>
      {deploy.gitopsHashes.map(hashStatus => (
        <p key={hashStatus.hash} className="pl-2">
          <span>📋</span>
          <a
            href={`https://github.com/${gitopsRepo}/commit/${hashStatus.hash}`}
            target="_blank" rel="noopener noreferrer"
            className='ml-1'
          >
            {hashStatus.hash.slice(0, 6)}
          </a>
        </p>
      ))}
    </div>
  )
}

function appliedWidgetFromResults(deploy, gitopsCommits, env, gitopsRepo) {
  const firstCommitOfEnv = gitopsCommits.filter((gitopsCommit) => gitopsCommit.env === env)[0];

  let deployCommit = {};
  deploy.results.forEach(result => {
    if (result.hash === firstCommitOfEnv.sha) {
      deployCommit = Object.assign({}, firstCommitOfEnv);
    }
  })

  return renderAppliedWidget(deployCommit, gitopsRepo);
}

function appliedWidgetFromGitopsHashes(deploy, gitopsCommits, env, gitopsRepo) {
  const firstCommitOfEnv = gitopsCommits.filter((gitopsCommit) => gitopsCommit.env === env)[0];

  let deployCommit = {};
  deploy.gitopsHashes.forEach(gitopsHash => {
    if (gitopsHash.hash === firstCommitOfEnv.sha) {
      deployCommit = Object.assign({}, firstCommitOfEnv);
    }
  })

  return renderAppliedWidget(deployCommit, gitopsRepo);
}
