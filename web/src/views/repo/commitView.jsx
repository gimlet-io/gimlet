import { useState, useEffect } from 'react';
import { DeployStatusModal } from './deployStatus';
import Commits from "../../components/commits/commits";
import Dropdown from "../../components/dropdown/dropdown";
import {
  ACTION_TYPE_ENVCONFIGS,
} from "../../redux/redux";
import DeployHandler from '../../deployHandler';
import { useHistory, useLocation, useParams } from 'react-router-dom'
import {produce} from 'immer';

export function CommitView(props) {
  const { store, gimletClient } = props;
  const reduxState = store.getState();
  const location = useLocation()
  const history = useHistory();

  const { owner, repo } = useParams();
  const repoName = `${owner}/${repo}`;
  const queryParams = new URLSearchParams(location.search)

  const [deployStatusModal, setDeployStatusModal] = useState(false)
  const [commits, setCommits] = useState()
  const [commitEvents, setCommitEvents] = useState(reduxState.commitEvents)
  const [branches, setBranches] = useState()
  const [selectedBranch, setSelectedBranch] = useState(queryParams.get("branch") ?? '')
  const [settings, setSettigns] = useState(reduxState.settings)
  const [envConfigs, setEnvConfigs] = useState(reduxState.envConfigs[repoName])
  const [envs, setEnvs] = useState(reduxState.envs)
  const [connectedAgents, setConnectedAgents] = useState(reduxState.connectedAgents)
  // eslint-disable-next-line no-unused-vars
  const [refreshQueue, setRefreshQueue] = useState(reduxState.repoRefreshQueue.filter(repo => repo === repoName).length)

  store.subscribe(() => {
    const reduxState = store.getState()
    setEnvConfigs(reduxState.envConfigs[repoName])
    setEnvs(reduxState.envs)
    setConnectedAgents(reduxState.connectedAgents)
    setSettigns(reduxState.settings)

    const queueLength = reduxState.repoRefreshQueue.filter(r => r === repoName).length;
    setRefreshQueue(prevQueueLength => {
      if (prevQueueLength !== queueLength) {
        refreshBranches(owner, repo);
        refreshCommits(owner, repo, selectedBranch);
        refreshConfigs(owner, repo);
      }
      return queueLength
    })

    setCommitEvents(reduxState.commitEvents)
  })

  useEffect(() => {
    if (commits) {
      commits.forEach(c => {
        if (!reduxState.commitEvents[c.sha]) {
          return
        }
        const commitEvents = Object.values(reduxState.commitEvents[c.sha])
        let latest = commitEvents[0]
        commitEvents.forEach(e => {
          if (e.created > latest.created) {
            latest = e
          }
        })
        if (latest.created > c.lastEvent.created ||
          latest.status !== c.lastEvent.status // handle status updates on an existing event, not just newer events
        ) {
          const updated = produce(commits, draft => {
            var idx
            draft.forEach((e, index) => {
              if (e.sha === c.sha) {
                idx = index
              }
            });
            draft[idx].lastEvent = latest;
          })
          setCommits(updated)
        }
      })
    }
  }, [commitEvents]);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    params.set('branch', selectedBranch);
    history.replace({ pathname: location.pathname, search: params.toString() });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedBranch]);

  useEffect(() => {
    gimletClient.getCommits(owner, repo, selectedBranch, "head")
      .then(data => {
        setCommits(data)
      }, () => {/* Generic error handler deals with it */
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedBranch]);

  useEffect(() => {
    gimletClient.getEnvConfigs(owner, repo)
      .then(envConfigs => {
        store.dispatch({
          type: ACTION_TYPE_ENVCONFIGS, payload: {
            owner: owner,
            repo: repo,
            envConfigs: envConfigs
          }
        });
      }, () => {/* Generic error handler deals with it */
    });
    gimletClient.getBranches(owner, repo)
      .then(data => {
        let defaultBranch = 'main'
        for (let branch of data) {
          if (branch === "master") {
            defaultBranch = "master";
          }
        }

        if (selectedBranch === "") {
          setSelectedBranch(defaultBranch)
        }
        setBranches(data)
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const deployHandler = new DeployHandler(owner, repo, gimletClient, store)

  const refreshBranches = (owner, repo) => {
    gimletClient.getBranches(owner, repo)
      .then(data => {
        let defaultBranch = 'main'
        for (let branch of data) {
          if (branch === "master") {
            defaultBranch = "master";
          }
        }

        if (selectedBranch === "") {
          setSelectedBranch(defaultBranch)
        }
        setBranches(data)
      })
  }

  const refreshCommits = (owner, repo, branch) => {
    gimletClient.getCommits(owner, repo, branch, "head")
      .then(data => {
        setCommits(data)
      }, () => {/* Generic error handler deals with it */
      });
  }

  const refreshConfigs = (owner, repo) => {
    gimletClient.getEnvConfigs(owner, repo)
      .then(envConfigs => {
        store.dispatch({
          type: ACTION_TYPE_ENVCONFIGS, payload: {
            owner: owner,
            repo: repo,
            envConfigs: envConfigs
          }
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  const fetchNextCommitsWidgets = () => {
    const lastCommit = commits[commits.length - 1]

    gimletClient.getCommits(owner, repo, selectedBranch, lastCommit.sha)
      .then(data => {
        const uniqueCommits = data.filter(commit => (
          !commits.some(existingCommit => existingCommit.sha === commit.sha)
        ));

        setCommits((prevCommits) => [...prevCommits, ...uniqueCommits]);
      }, () => {/* Generic error handler deals with it */
      });
  }

  if (!envConfigs || !commits) {
    return <SkeletonLoader />
  }

  return (
    <div className='max-w-7xl mx-auto pt-32 px-4 sm:px-6 lg:px-8'>
      {deployStatusModal && envConfigs !== undefined &&
        <DeployStatusModal
          closeHandler={() => setDeployStatusModal(false)}
          owner={owner}
          repoName={repo}
          envConfigs={envConfigs}
          store={store}
          gimletClient={gimletClient}
        />
      }
      <div className="mb-4 lg:mb-8 flex">
        <div className='flex-grow'>
          <div className='w-96'>
            <Dropdown
              items={branches}
              value={selectedBranch}
              changeHandler={setSelectedBranch}
            />
          </div>
        </div>
        {settings.instance === "" &&
        <button
            type="button"
            className='secondaryButton'
            onClick={() => {
              gimletClient.triggerCommitSync(owner, repo)
            }}
          >
            Refresh
        </button>
        }
      </div>
      <div className="card p-4">
        <Commits
          commits={commits}
          envs={envs}
          connectedAgents={connectedAgents}
          deployHandler={(target, sha, repo) => {
            setDeployStatusModal(true);
            deployHandler.deploy(target, sha, repo)
          }}
          repo={repo}
          gimletClient={gimletClient}
          store={store}
          owner={owner}
          fetchNextCommitsWidgets={fetchNextCommitsWidgets}
          scmUrl={settings.scmUrl}
        />
      </div>
    </div>
  )
}

const SkeletonLoader = () => {
  return (
    <div className='max-w-7xl mx-auto pt-32 px-4 sm:px-6 lg:px-8 animate-pulse'>
      <div className="w-96 mb-4 lg:mb-8">
        <div
          className="bg-white dark:bg-neutral-800 mt-1 relative w-full border border-neutral-200 dark:border-neutral-700 rounded-md h-9 pl-3 pr-10 py-2"
          id="headlessui-listbox-button-4"
          type="button"
          aria-haspopup="listbox"
          aria-expanded="false"
          data-headlessui-state=""
        >
        </div>
      </div>
      <div className="card p-4">
        <ul className="-mb-4">
          <li>
            <div className="relative pl-2 py-4 rounded">
              <span className="absolute top-4 left-6 -ml-px h-full w-0.5 bg-neutral-200 dark:bg-neutral-400" aria-hidden="true"></span>
              <div className="relative flex items-start space-x-3">
                <svg className="w-8 h-8 text-neutral-200 dark:text-neutral-400 relative bg-white dark:bg-neutral-800" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M10 0a10 10 0 1 0 10 10A10.011 10.011 0 0 0 10 0Zm0 5a3 3 0 1 1 0 6 3 3 0 0 1 0-6Zm0 13a8.949 8.949 0 0 1-4.951-1.488A3.987 3.987 0 0 1 9 13h2a3.987 3.987 0 0 1 3.951 3.512A8.949 8.949 0 0 1 10 18Z"/>
                </svg>
                <div className="w-full max-w-4xl space-y-3">
                <div className="h-2 bg-neutral-400 dark:bg-neutral-600 rounded w-5/5"></div>
                  <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
                  <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/5"></div>
                </div>
              </div>
            </div>
          </li>
          <li>
            <div className="relative pl-2 py-4 rounded">
              <div className="relative flex items-start space-x-3">
                <svg className="w-8 h-8 text-neutral-200 dark:text-neutral-400 relative" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M10 0a10 10 0 1 0 10 10A10.011 10.011 0 0 0 10 0Zm0 5a3 3 0 1 1 0 6 3 3 0 0 1 0-6Zm0 13a8.949 8.949 0 0 1-4.951-1.488A3.987 3.987 0 0 1 9 13h2a3.987 3.987 0 0 1 3.951 3.512A8.949 8.949 0 0 1 10 18Z"/>
                </svg>
                <div className="w-full max-w-4xl space-y-3">
                  <div className="h-2 bg-neutral-400 dark:bg-neutral-600 rounded w-5/5"></div>
                  <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-2/5"></div>
                  <div className="h-2 bg-neutral-300 dark:bg-neutral-500 rounded w-1/5"></div>
                </div>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </div>
  );
}
