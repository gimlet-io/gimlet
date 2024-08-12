import React, { Component } from 'react';
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

export default class Repo extends Component {
  constructor(props) {
    super(props);
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    let reduxState = this.props.store.getState();
    this.state = {
      connectedAgents: reduxState.connectedAgents,
      rolloutHistory: reduxState.rolloutHistory,
      envConfigs: reduxState.envConfigs[repoName],
      settings: reduxState.settings,
      refreshQueue: reduxState.repoRefreshQueue.filter(repo => repo === repoName).length,
      agents: reduxState.settings.agents,
      envs: reduxState.envs,
      repoMetas: reduxState.repoMetas,
      fileInfos: reduxState.fileInfos,
      alerts: reduxState.alerts,
      deployStatusModal: false,
      selectedEnv: localStorage.getItem(repoName + "-selected-env") ?? "All Environments",
      appFilter: ""
    }

    this.deployHandler = new DeployHandler(owner, repo, this.props.gimletClient, this.props.store)

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        connectedAgents: reduxState.connectedAgents,
        rolloutHistory: reduxState.rolloutHistory,
        envConfigs: reduxState.envConfigs[repoName],
        envs: reduxState.envs,
        repoMetas: reduxState.repoMetas,
        fileInfos: reduxState.fileInfos,
        scmUrl: reduxState.settings.scmUrl,
        alerts: reduxState.alerts,
      });

      const queueLength = reduxState.repoRefreshQueue.filter(r => r === repoName).length
      this.setState(prevState => {
        if (prevState.refreshQueueLength !== queueLength) {
          this.refreshConfigs(owner, repo);
        }
        return { refreshQueueLength: queueLength }
      });
      this.setState({ agents: reduxState.settings.agents });
    });

    this.navigateToConfigEdit = this.navigateToConfigEdit.bind(this)
    this.linkToDeployment = this.linkToDeployment.bind(this)
    this.setSelectedEnv = this.setSelectedEnv.bind(this)
    this.setAppFilter = this.setAppFilter.bind(this)
  }

  componentDidMount() {
    const { owner, repo } = this.props.match.params;

    this.props.gimletClient.getRepoMetas(owner, repo)
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_REPO_METAS, payload: {
            repoMetas: data,
          }
        });
      }, () => {/* Generic error handler deals with it */
    });

    this.props.gimletClient.getEnvConfigs(owner, repo)
      .then(envConfigs => {
        this.props.store.dispatch({
          type: ACTION_TYPE_ENVCONFIGS, payload: {
            owner: owner,
            repo: repo,
            envConfigs: envConfigs
          }
        });
      }, () => {/* Generic error handler deals with it */
    });
  }

  componentDidUpdate(prevProps, prevState) {
    if (prevState.selectedEnv !== this.state.selectedEnv) {
      const { owner, repo } = this.props.match.params;
      const repoName = `${owner}/${repo}`;

      localStorage.setItem(repoName + "-selected-env", this.state.selectedEnv)
    }
  }

  setSelectedEnv(selectedEnv) {
    this.setState({
      selectedEnv: selectedEnv
    })
  }

  setAppFilter(filter) {
    this.setState({
      appFilter: filter
    })
  }

  refreshConfigs(owner, repo) {
    this.props.gimletClient.getEnvConfigs(owner, repo)
      .then(envConfigs => {
        this.props.store.dispatch({
          type: ACTION_TYPE_ENVCONFIGS, payload: {
            owner: owner,
            repo: repo,
            envConfigs: envConfigs
          }
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  navigateToConfigEdit(env, config) {
    const { owner, repo } = this.props.match.params;
    this.props.history.push(encodeURI(`/repo/${owner}/${repo}/envs/${env}/config/${config}/edit`))
  }

  linkToDeployment(env, deployment) {
    const { owner, repo } = this.props.match.params;
    this.props.history.push({
      pathname: `/repo/${owner}/${repo}/${env}/${deployment}`,
      search: this.props.location.search
    })
  }

  fileMetasByEnv(envName) {
    return this.state.fileInfos.filter(fileInfo => fileInfo.envName === envName)
  }

  render() {
    const { owner, repo, environment, deployment } = this.props.match.params;
    const repoName = `${owner}/${repo}`
    const { envs, connectedAgents, rolloutHistory, settings, selectedEnv } = this.state;
    const { envConfigs, scmUrl, alerts, appFilter, deployStatusModal } = this.state;

    const stacksForRepo = envsForRepo(envs, connectedAgents, repoName);

    let repoRolloutHistory = undefined;
    if (rolloutHistory && rolloutHistory[repoName]) {
      repoRolloutHistory = rolloutHistory[repoName]
    }

    const envLabels = envs.map((env) => env.name)
    envLabels.unshift('All Environments')

    return (
      <div>
        {deployStatusModal && envConfigs !== undefined &&
          <DeployStatusModal
            closeHandler={() => this.setState({deployStatusModal: false})}
            owner={owner}
            repoName={repo}
            envConfigs={envConfigs}
            store={this.props.store}
            gimletClient={this.props.gimletClient}
          />
        }
        <header>
          <div className="max-w-7xl mx-auto pt-32 px-4 sm:px-6 lg:px-8">
            <div className='flex items-center space-x-2'>
              <AppFilter
                setFilter={this.setAppFilter}
              />
              <div className="w-96 capitalize">
                <Dropdown
                  items={envLabels}
                  value={selectedEnv}
                  changeHandler={this.setSelectedEnv}
                  buttonClass="capitalize"
                />
              </div>
              <MenuButton
                items={envs}
                handleClick={
                  (envName) => this.props.history.push(encodeURI(`/repo/${owner}/${repo}/envs/${envName}/deploy`))}
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
                    return unselected ? null : (
                    <Env
                      key={envName}
                      env={stacksForRepo[envName]}
                      repoRolloutHistory={repoRolloutHistory}
                      envConfigs={envConfigs[envName]}
                      navigateToConfigEdit={this.navigateToConfigEdit}
                      linkToDeployment={this.linkToDeployment}
                      rollback={(env, app, rollbackTo) => {
                        this.setState({deployStatusModal: true});
                        this.deployHandler.rollback(env, app, rollbackTo)
                      }}
                      owner={owner}
                      repoName={repo}
                      fileInfos={this.fileMetasByEnv(envName)}
                      releaseHistorySinceDays={settings.releaseHistorySinceDays}
                      gimletClient={this.props.gimletClient}
                      store={this.props.store}
                      envFromParams={environment}
                      deploymentFromParams={deployment}
                      scmUrl={scmUrl}
                      history={this.props.history}
                      alerts={alerts}
                      appFilter={appFilter}
                    />)}
                )}
              </div>
            </div>
          </div>
        </main>
      </div>
    )
  }
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

