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
      settings: reduxState.settings
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({runningDeploys: reduxState.runningDeploys});
      this.setState({settings: reduxState.settings});
    });
  }

  render() {
    const {runningDeploys, settings} = this.state;
    const gitopsRepo = settings.gitopsRepo;

    if (runningDeploys.length === 0) {
      return null;
    }

    const deploy = runningDeploys[0];

    let gitopsWidget = (
      <div className="mt-2">
        <Loading/>
      </div>
    )
    let appliedWidget = null;

    if (deploy.status === 'error') {
      gitopsWidget = (
        <div class="mt-2">
          <p className="text-red-500 font-semibold">
            Gitops write failed
          </p>
          <p className="text-red-500 font-base">
            {deploy.statusDesc}
          </p>
        </div>
      )
    }

    if (deploy.gitopsHashes && deploy.gitopsHashes.length !== 0) {
      gitopsWidget = (
        <div class="mt-2">
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
      appliedWidget = (
        <Loading/>
      )
    }

    if (deploy.gitopsHashes) {
      const numberOfGitopsHashes = deploy.gitopsHashes.length;
      if (numberOfGitopsHashes > 0) {
        const latestGitopsHashMetadata = deploy.gitopsHashes[0];
        if (latestGitopsHashMetadata.status !== 'N/A') {
          appliedWidget = deploy.gitopsHashes.map(hashStatus => {
            if (hashStatus.status !== 'Progressing' &&
              hashStatus.status !== 'N/A') {
              return (
                <p key={hashStatus.hash} className="font-semibold text-red-500">
                  <span>❗</span>
                  <a
                    href={`https://github.com/${gitopsRepo}/commit/${hashStatus.hash}`}
                    target="_blank" rel="noopener noreferrer"
                    class='ml-1'
                  >
                    {hashStatus.hash.slice(0, 6)}
                  </a>
                  <span class='ml-1 block'>{hashStatus.statusDesc}</span>
                </p>
              )
            } else {
              return (
                <p key={hashStatus.hash} className="font-semibold text-green-300">
                  <span>✅</span>
                  <a
                    href={`https://github.com/${gitopsRepo}/commit/${hashStatus.hash}`}
                    target="_blank" rel="noopener noreferrer"
                    class='ml-1'
                  >
                    {hashStatus.hash.slice(0, 6)}
                  </a>
                  <span class='ml-1'>applied</span>
                </p>
              )
            }
          })
        }
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
                      <p class="pl-2  ">
                        🎯 {deploy.env}
                      </p>
                      {!deploy.rollback &&
                      <p class="pl-2">
                        <span>📎</span>
                        <a
                          href={`https://github.com/${deploy.repo}/commit/${deploy.sha}`}
                          target="_blank" rel="noopener noreferrer"
                          class='ml-1'
                        >
                          {deploy.sha.slice(0, 6)}
                        </a>
                      </p>
                      }
                      {gitopsWidget}
                      <div class='pl-2 mt-4'>{appliedWidget}</div>
                    </div>
                    <div className="ml-4 flex-shrink-0 flex">
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

function

Loading() {
  return (
    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none"
         viewBox="0 0 24 24">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
      <path className="opacity-75" fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
    </svg>
  )
}
