import { useState, useEffect } from 'react';
import { ArrowPathIcon } from '@heroicons/react/20/solid'
import { format, formatDistance } from "date-fns";
import { DeployStatusModal } from './deployStatus';
import DeployHandler from '../../deployHandler';
import { envsForRepo } from './repo'
import {
  ACTION_TYPE_ENVCONFIGS,
} from "../../redux/redux";
import ServiceDetail from '../../components/serviceDetail/serviceDetail';
import MenuButton from '../../components/menuButton/menuButton';
import { CommitWidget } from '../../components/commits/commits';
import { v4 as uuidv4 } from 'uuid';
import {produce} from 'immer';
import { useParams, useLocation, useHistory } from 'react-router-dom'

export function PreviewView(props) {
  const { store, gimletClient } = props;
  const reduxState = store.getState();

  const { owner, repo } = useParams();
  const repoName = `${owner}/${repo}`;
  let history = useHistory()

  const [pullRequests, setPullRequests] = useState()
  const [rolloutHistory, setRolloutHistory] = useState(reduxState.rolloutHistory?.[repoName])
  const [settings, setSettings] = useState(reduxState.settings)
  const [envConfigs, setEnvConfigs] = useState(reduxState.envConfigs[repoName])
  const [envs, setEnvs] = useState(reduxState.envs)
  const [connectedAgents, setConnectedAgents] = useState(reduxState.connectedAgents)
  const [branches, setBranches] = useState()
  // eslint-disable-next-line no-unused-vars
  const [refreshQueue, setRefreshQueue] = useState(reduxState.repoRefreshQueue.filter(repo => repo === repoName).length)
  const [renderId, setRenderId] = useState()
  const [deployStatusModal, setDeployStatusModal] = useState(false)

  store.subscribe(() => {
    const reduxState = store.getState()
    setEnvConfigs(reduxState.envConfigs[repoName])
    setEnvs(reduxState.envs)
    setConnectedAgents(reduxState.connectedAgents)
    setSettings(reduxState.settings)
    setRolloutHistory(reduxState.rolloutHistory?.[repoName])
    const queueLength = reduxState.repoRefreshQueue.filter(r => r === repoName).length;
    setRefreshQueue(prevQueueLength => {
      if (prevQueueLength !== queueLength) {
        setRenderId(uuidv4())
      }
      return queueLength
    })
  })

  useEffect(() => {
    gimletClient.getPullRequests(owner, repo)
      .then(data => {
        setPullRequests(data)
      }, () => {/* Generic error handler deals with it */
    });
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
        setBranches(data.filter(b => b !== 'main' && b !== 'master'))
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!envConfigs || !pullRequests || !branches) {
    return <SkeletonLoader />
  }

  const envNames = envs.map(e => e.name)
  const previewEnvConfig = Object.keys(envConfigs)
    .flatMap(envName => envConfigs[envName]).filter(c => envNames.includes(c.env)).find(c => c.preview)

  const branchesWithoutPullRequest = branches.filter(branch => !pullRequests.find(pr => pr.branch === branch));

  if (!previewEnvConfig) {
    return (
      <div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32">
          <h1 className="text-2xl leading-tight font-medium flex-grow">Previews</h1>
          <NoConfig items={envs} owner={owner} repo={repo} />
        </div>
        <div className="border-b border-neutral-200 dark:border-neutral-700 my-8"></div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center">
          <h1 className="text-2xl leading-tight font-medium flex-grow">Open Pull Requests</h1>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-2 pb-4 flex items-center">
          <div className='w-full card mt-4'>
            {pullRequests.length === 0 && <NoPullRequests />}
            <div className='divide-y dark:divide-neutral-700'>
            {
              pullRequests.map(pr => <Branch {...props} branch={pr.branch} pr={pr} key={pr.number} stacks={[]} settings={settings} />)
            }
            </div>
          </div>
        </div>
      </div>
    )
  }

  const stacksForRepo = envsForRepo(envs, connectedAgents, repoName);
  const stacks = stacksForRepo[previewEnvConfig.env]?.stacks
  const deployHandler = new DeployHandler(owner, repo, gimletClient, store)

  return (
    <div>
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
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32 flex items-center">
        <h1 className="text-2xl leading-tight font-medium flex-grow">Previews</h1>
        <button
            type="button"
            className='secondaryButton'
            onClick={() => history.push(encodeURI(`/repo/${owner}/${repo}/envs/${previewEnvConfig.env}/config/${repo}-preview/edit-preview`))}
          >
            Edit Preview Config
        </button>
      </div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center">
        <div className='flex font-light text-sm text-neutral-600 dark:text-neutral-300 pt-1'>
          <ArrowPathIcon className="h-4 mr-1 mt-0.5" aria-hidden="true" />
          Continuously deployed on pushes to branches
        </div>
      </div>
      <div className="border-b border-neutral-200 dark:border-neutral-700 my-8"></div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-4 flex items-center">
        <div className='w-full card mt-4'>
          {pullRequests.length === 0 && <NoPullRequests />}
          <div className='divide-y dark:divide-neutral-700'>
          {
            pullRequests.map((pr, idx) => <Branch
              {...props}
              renderId={renderId}
              branch={pr.branch}
              pr={pr}
              owner={owner}
              repo={repo}
              key={pr.number}
              stacks={stacks}
              config={previewEnvConfig}
              settings={settings}
              rolloutHistory={rolloutHistory}
              envs={envs}
              setDeployStatusModal={setDeployStatusModal}
              deployHandler={deployHandler}
            />)
          }
          {pullRequests.length !== 0 &&
            branchesWithoutPullRequest.map((branch, idx) => 
              <Branch
                {...props}
                renderId={renderId}
                branch={branch}
                owner={owner}
                repo={repo}
                key={idx}
                stacks={stacks}
                config={previewEnvConfig}
                settings={settings}
                rolloutHistory={rolloutHistory}
                envs={envs}
                setDeployStatusModal={setDeployStatusModal}
                deployHandler={deployHandler}
              />
            )
          }
          </div>
        </div>
      </div>
    </div>
  )
}

const NoConfig = (props) => {
  const { owner, repo } = props;

  let history = useHistory()

  return (
    <div className='w-full card p-4 mt-4'>
      <div className='items-center border-dashed border border-neutral-200 dark:border-neutral-700 rounded-md p-4 py-16'>
        <h3 className="mt-2 text-sm font-semibold text-center">No preview config</h3>
        <p className="mt-1 text-sm text-neutral-500 text-center">Get started by configuring preview deploys.</p>
        <div className="mt-6 text-center">
          <MenuButton {...props} handleClick={(envName) => history.push(encodeURI(`/repo/${owner}/${repo}/envs/${envName}/config/${repo}-preview/new-preview`))}>
            Configure Previews..
          </MenuButton>
        </div>
      </div>
    </div>
  );
}

const NoPullRequests = () => {
  return (
    <div className='flex items-center h-96 border-dashed border border-neutral-200 dark:border-neutral-700 rounded-md m-4'>
      <div className='mx-auto font-medium text-xl'>No Open Pull Requests</div>
    </div>
  );
}

const Branch = (props) => {
  const { owner, repo, deployment } = useParams();
  let location = useLocation()
  let history = useHistory()
  const { store, gimletClient, renderId } = props;
  const { branch, pr, stacks, envs, config, settings, rolloutHistory } = props
  const { setDeployStatusModal, deployHandler } = props

  const [latestCommitEvent, setLatestCommitEvent] = useState()
  const [latestCommit, setLatestCommit] = useState()

  store.subscribe(() => {
    const reduxState = store.getState()
    if (latestCommit && reduxState.commitEvents[latestCommit.sha]) {
      const events = Object.values(reduxState.commitEvents[latestCommit.sha])
      let latest = events[0]
      events.forEach(e => {
        if (e.created > latest.created) {
          latest = e
        }
      })
      setLatestCommitEvent(latest)
    }
  })

  useEffect(() => {
    if (!config) {
      return
    }

    gimletClient.getCommits(owner, repo, branch, "head")
      .then(data => {
        if (data.length > 0) {
          setLatestCommit(data[0])
        }
      }, () => {/* Generic error handler deals with it */
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [renderId]);

  useEffect(() => {
    const reduxState = store.getState()
    if (latestCommit && reduxState.commitEvents[latestCommit.sha]) {
      const events = Object.values(reduxState.commitEvents[latestCommit.sha])
      let latest = events[0]
      events.forEach(e => {
        if (e.created > latest.created) {
          latest = e
        }
      })
      setLatestCommit(produce(latestCommit, draft => {
        draft.lastEvent = latest
      }))
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [latestCommit]);

  useEffect(() => {
    if (latestCommit){
      setLatestCommit(produce(latestCommit, draft => {
        draft.lastEvent = latestCommitEvent
      }))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [latestCommitEvent]);

  const linkToDeployment = (env, deployment) => {
    history.push({
      pathname: `/repo/${owner}/${repo}/previews/${deployment}`,
      search: location.search
    })
  }

  const stack = stacks?.find(s => s.deployment?.branch === branch)

  const envNames = envs?.map(env => env["name"]);

  return (
    <div className='flex p-6'>
      <div>
        <svg data-testid="geist-icon" strokeLinejoin="round" className='text-neutral-800 dark:text-neutral-300 h-5 mt-1' viewBox="0 0 16 16">
          <path fillRule="evenodd" clipRule="evenodd" d="M4.75 1.75V1H3.25V1.75V9.09451C1.95608 9.42754 1 10.6021 1 12C1 13.6569 2.34315 15 4 15C5.42051 15 6.61042 14.0127 6.921 12.6869C8.37102 12.4872 9.7262 11.8197 10.773 10.773C11.8197 9.7262 12.4872 8.37102 12.6869 6.921C14.0127 6.61042 15 5.42051 15 4C15 2.34315 13.6569 1 12 1C10.3431 1 9 2.34315 9 4C9 5.37069 9.91924 6.52667 11.1749 6.8851C10.9929 7.94904 10.4857 8.9389 9.71231 9.71231C8.9389 10.4857 7.94904 10.9929 6.8851 11.1749C6.59439 10.1565 5.77903 9.35937 4.75 9.09451V1.75ZM13.5 4C13.5 4.82843 12.8284 5.5 12 5.5C11.1716 5.5 10.5 4.82843 10.5 4C10.5 3.17157 11.1716 2.5 12 2.5C12.8284 2.5 13.5 3.17157 13.5 4ZM4 13.5C4.82843 13.5 5.5 12.8284 5.5 12C5.5 11.1716 4.82843 10.5 4 10.5C3.17157 10.5 2.5 11.1716 2.5 12C2.5 12.8284 3.17157 13.5 4 13.5Z" fill="currentColor"></path>
        </svg>
      </div>
      <div className='w-full ml-4'>
        <div className='text-sm'>
          <div>
            <span className='font-medium'><a href={`${settings.scmUrl}/${owner}/${repo}/tree/${branch}`} target="_blank" rel="noopener noreferrer">{branch}</a></span>
          </div>
          {pr &&
          <div className='pt-2'>
            <div><a href={pr.link} target="_blank" rel="noopener noreferrer">#{pr.number} {pr.title}</a></div>
            <div className='text-neutral-800 dark:text-neutral-300 text-xs'>Opened by {pr.author} <span title={format(pr.created * 1000, 'h:mm:ss a, MMMM do yyyy')}>{formatDistance(pr.created * 1000, new Date())} ago</span></div>
          </div>
          }
          {latestCommit &&
            <ul className="-mb-4">
              <CommitWidget
                owner={owner}
                repo={repo}
                commit={latestCommit}
                envNames={envNames}
                scmUrl={settings.scmUrl}
                gimletClient={gimletClient}
                envs={envs}
                last={true}
              />
            </ul>
          }
        </div>
        {stack &&
          <div className="w-full flex items-center justify-between space-x-6 mt-4 bg-neutral-100 dark:bg-neutral-900 p-2 rounded-md">
            <ServiceDetail
              key={stack.service.name}
              stack={stack}
              rolloutHistory={rolloutHistory?.[config.env]?.[stack.service.name]}
              rollback={(env, app, rollbackTo) => {
                setDeployStatusModal(true);
                deployHandler.rollback(env, app, rollbackTo)
              }}
              environment={envs.find(e => e.name === config.env)}
              owner={owner}
              repoName={repo}
              linkToDeployment={linkToDeployment}
              config={config}
              releaseHistorySinceDays={settings.releaseHistorySinceDays}
              gimletClient={gimletClient}
              store={store} 
              deploymentFromParams={deployment}
              scmUrl={settings.scmUrl}
            />
          </div>
        }
      </div>
    </div>
  )
}

const SkeletonLoader = () => {
  return (
    <div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32">
        <h1 className="text-2xl leading-tight font-medium flex-grow">Previews</h1>
      </div>
      <div className="border-b border-neutral-200 dark:border-neutral-700 my-8"></div>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-4">
        <div role="status" className="w-full mt-4 h-[278px] bg-neutral-200 dark:bg-neutral-800 rounded-lg animate-pulse">
          <span className="sr-only">Loading...</span>
        </div>
      </div>
    </div>
  );
}
