import { useState, useEffect } from 'react';
import {
  ACTION_TYPE_ENVCONFIGS,
  ACTION_TYPE_REPO_METAS,
} from "../../redux/redux";
import { Env } from '../../components/env/env';
import { FunnelIcon } from '@heroicons/react/20/solid'
import MenuButton from '../../components/menuButton/menuButton';
import Dropdown from '../../components/dropdown/dropdown';
import { DeployStatusModal } from './deployStatus';
import DeployHandler from '../../deployHandler';
import { useLocation, useNavigate, useParams } from 'react-router-dom'

export default function Repo(props) {
  const { store, gimletClient } = props
  const { owner, repo, environment, deployment } = useParams()
  const repoName = `${owner}/${repo}`
  const navigate = useNavigate()
  const location = useLocation()

  const reduxState = store.getState();
  const [connectedAgents, setConnectedAgents] = useState(reduxState.connectedAgents)
  const [rolloutHistory, setRolloutHistory] = useState(reduxState.rolloutHistory)
  const [envConfigs, setEnvConfigs] = useState(reduxState.envConfigs[repoName])
  const [settings, setSettings] = useState(reduxState.settings)
  const [refreshQueue, setRefreshQueue] = useState(reduxState.repoRefreshQueue.filter(repo => repo === repoName).length)
  const [refreshQueueLength, setRefreshQueueLength] = useState(0)
  const [agents, setAgents] = useState(reduxState.settings.agents)
  const [envs, setEnvs] = useState(reduxState.envs)
  const [repoMetas, setRepoMetas] = useState(reduxState.repoMetas)
  const [fileInfos, setFileInfos] = useState(reduxState.fileInfos)
  const [alerts, setAlerts] = useState(reduxState.alerts)
  const [deployStatusModal, setDeployStatusModal] = useState(false)
  const [selectedEnv, setSelectedEnv] = useState(localStorage.getItem(repoName + "-selected-env") ?? "All Environments")
  const [appFilter, setAppFilter] = useState("")
  const deployHandler = new DeployHandler(owner, repo, gimletClient, store)

  store.subscribe(() => {
    const reduxState = store.getState();
    setConnectedAgents(reduxState.connectedAgents)
    setRolloutHistory(reduxState.rolloutHistory)
    setEnvConfigs(reduxState.envConfigs[repoName])
    setEnvs(reduxState.envs)
    setRepoMetas(reduxState.repoMetas)
    setFileInfos(reduxState.fileInfos)
    setSettings(reduxState.settings)
    setAlerts(reduxState.alerts)

    const queueLength = reduxState.repoRefreshQueue.filter(r => r === repoName).length
    setRefreshQueueLength(prevState => {
      if (prevState !== queueLength) {
        refreshConfigs(owner, repo);
      }
      return queueLength
    });
    setAgents(reduxState.settings.agents);
  });

  useEffect(() => {
    gimletClient.getRepoMetas(owner, repo)
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_REPO_METAS, payload: {
            repoMetas: data,
          }
        });
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    localStorage.setItem(repoName + "-selected-env", selectedEnv)    
  }, [selectedEnv]);

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

  const navigateToConfigEdit = (env, config) => {
    navigate(encodeURI(`/repo/${owner}/${repo}/envs/${env}/config/${config}/edit`))
  }

  const linkToDeployment = (env, deployment) => {
    navigate(`/repo/${owner}/${repo}/${env}/${deployment}?${location.search}`)
  }

  const stacksForRepo = envsForRepo(envs, connectedAgents, repoName);

  let repoRolloutHistory = undefined;
  if (rolloutHistory && rolloutHistory[repoName]) {
    repoRolloutHistory = rolloutHistory[repoName]
  }

  const envLabels = envs.map((env) => env.name)
  envLabels.unshift('All Environments')

  return (
    <div>
      {deployStatusModal && envConfigs &&
        <DeployStatusModal
          closeHandler={() => setDeployStatusModal(false)}
          owner={owner}
          repoName={repo}
          envConfigs={envConfigs}
          store={store}
          gimletClient={gimletClient}
        />
      }
      <header>
        <div className="max-w-7xl mx-auto pt-32 px-4 sm:px-6 lg:px-8">
          <div className='flex items-center space-x-2'>
            <AppFilter
              setFilter={setAppFilter}
            />
            <div className="w-96 capitalize">
              <Dropdown
                items={envLabels}
                value={selectedEnv}
                changeHandler={setSelectedEnv}
                buttonClass="capitalize"
              />
            </div>
            <MenuButton
              items={envs}
              handleClick={
                (envName) => navigate(encodeURI(`/repo/${owner}/${repo}/envs/${envName}/deploy`))}
            >
              New deployment..
            </MenuButton>
          </div>
        </div>
      </header>
      <main>
        <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
          <div className="pt-8 px-4 sm:px-0">
            <div>
              {envConfigs && Object.keys(stacksForRepo).sort().map((envName) =>
                {
                  const unselected = envName !== selectedEnv && selectedEnv !== "All Environments"
                  return unselected ? null :
                  <Env
                    key={envName}
                    env={stacksForRepo[envName]}
                    repoRolloutHistory={repoRolloutHistory}
                    envConfigs={envConfigs[envName]}
                    navigateToConfigEdit={navigateToConfigEdit}
                    linkToDeployment={linkToDeployment}
                    rollback={(env, app, rollbackTo) => {
                      setDeployStatusModal(true);
                      deployHandler.rollback(env, app, rollbackTo)
                    }}
                    owner={owner}
                    repoName={repo}
                    fileInfos={fileInfos.filter(fileInfo => fileInfo.envName === envName)}
                    releaseHistorySinceDays={settings.releaseHistorySinceDays}
                    gimletClient={gimletClient}
                    store={store}
                    envFromParams={environment}
                    deploymentFromParams={deployment}
                    settings={settings}
                    alerts={alerts}
                    appFilter={appFilter}
                  />
                }
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}

function AppFilter(props) {
  const { setFilter } = props;

  return (
    <div className="w-full">
      <div className="relative">
        <div className="absolute inset-y-0 left-0 flex items-center pl-3">
          <FunnelIcon className="filterIcon" aria-hidden="true" />
        </div>
        <input
          onChange={e => setFilter(e.target.value)}
          type="text"
          name="filter"
          id="filter"
          className="filter"
          placeholder="All Deployments..."
        />
      </div>
    </div>
  )
}

export function envsForRepo(envs, connectedAgents, repoName) {
  let envsForRepo = {};

  if (!connectedAgents || !envs) {
    return envsForRepo;
  }
  
  for (const env of envs) {
    envsForRepo[env.name] = {
      ...env,
      isOnline: isOnline(connectedAgents, env)
    };

    envsForRepo[env.name].stacks = connectedAgents[env.name]?.stacks
      ? connectedAgents[env.name].stacks.filter(service => service.repo === repoName)
      : []
  }

  return envsForRepo;
}

function isOnline(onlineEnvs, singleEnv) {
  return Object.keys(onlineEnvs)
      .map(env => onlineEnvs[env])
      .some(onlineEnv => {
          return onlineEnv.name === singleEnv.name
      })
};

