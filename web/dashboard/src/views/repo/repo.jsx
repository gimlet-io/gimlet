import React, { Component } from 'react';
import { Spinner } from '../repositories/repositories';
import {
  ACTION_TYPE_BRANCHES,
  ACTION_TYPE_ENVCONFIGS,
  ACTION_TYPE_COMMITS,
  ACTION_TYPE_DEPLOY,
  ACTION_TYPE_DEPLOY_STATUS,
  ACTION_TYPE_IMAGEBUILD,
  ACTION_TYPE_IMAGEBUILD_STATUS,
  ACTION_TYPE_REPO_METAS,
  ACTION_TYPE_ROLLOUT_HISTORY,
  ACTION_TYPE_REPO_PULLREQUESTS,
  ACTION_TYPE_RELEASE_STATUSES,
} from "../../redux/redux";
import Commits from "../../components/commits/commits";
import Dropdown from "../../components/dropdown/dropdown";
import { Env } from '../../components/env/env';
import TenantSelector from './tenantSelector';
import RefreshButton from '../../components/refreshButton/refreshButton';
import { DeployStatusModal } from '../../components/deployStatus/deployStatus';

export default class Repo extends Component {
  constructor(props) {
    super(props);
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    const queryParams = new URLSearchParams(this.props.location.search)

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      connectedAgents: reduxState.connectedAgents,
      search: reduxState.search,
      rolloutHistory: reduxState.rolloutHistory,
      commits: reduxState.commits,
      branches: reduxState.branches,
      envConfigs: reduxState.envConfigs[repoName],
      selectedBranch: queryParams.get("branch") ?? '',
      selectedTenant: queryParams.get("tenant") ?? '',
      settings: reduxState.settings,
      refreshQueue: reduxState.repoRefreshQueue.filter(repo => repo === repoName).length,
      agents: reduxState.settings.agents,
      envs: reduxState.envs,
      repoMetas: reduxState.repoMetas,
      fileInfos: reduxState.fileInfos,
      pullRequests: reduxState.pullRequests.configChanges[repoName],
      alerts: reduxState.alerts,
      deployStatusModal: false,
      unselectedEnvs: [],
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        connectedAgents: reduxState.connectedAgents,
        rolloutHistory: reduxState.rolloutHistory,
        commits: reduxState.commits,
        branches: reduxState.branches,
        envConfigs: reduxState.envConfigs[repoName],
        envs: reduxState.envs,
        repoMetas: reduxState.repoMetas,
        fileInfos: reduxState.fileInfos,
        pullRequests: reduxState.pullRequests.configChanges[repoName],
        scmUrl: reduxState.settings.scmUrl,
        alerts: reduxState.alerts,
      });

      const queueLength = reduxState.repoRefreshQueue.filter(r => r === repoName).length
      this.setState(prevState => {
        if (prevState.refreshQueueLength !== queueLength) {
          this.refreshBranches(owner, repo);
          this.refreshCommits(owner, repo, prevState.selectedBranch);
          this.refreshConfigs(owner, repo);
        }
        return { refreshQueueLength: queueLength }
      });
      this.setState({ agents: reduxState.settings.agents });
    });

    this.branchChange = this.branchChange.bind(this)
    this.deploy = this.deploy.bind(this)
    this.rollback = this.rollback.bind(this)
    this.checkDeployStatus = this.checkDeployStatus.bind(this)
    this.navigateToConfigEdit = this.navigateToConfigEdit.bind(this)
    this.linkToDeployment = this.linkToDeployment.bind(this)
    this.newConfig = this.newConfig.bind(this)
    this.setSelectedTenant = this.setSelectedTenant.bind(this)
    this.refreshCommits = this.refreshCommits.bind(this)
    this.setUnselectedEnvs = this.setUnselectedEnvs.bind(this)
  }

  componentDidMount() {
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    if (JSON.parse(localStorage.getItem(repoName + "-unselected-envs"))) {
      const unselectedEnvs =JSON.parse(localStorage.getItem(repoName + "-unselected-envs"));
      this.setState({ unselectedEnvs: unselectedEnvs });
    }

    this.props.gimletClient.getRepoMetas(owner, repo)
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_REPO_METAS, payload: {
            repoMetas: data,
          }
        });
      }, () => {/* Generic error handler deals with it */
    });

    this.getPullRequests(owner, repo)

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

    this.props.gimletClient.getBranches(owner, repo)
      .then(data => {
        let defaultBranch = 'main'
        for (let branch of data) {
          if (branch === "master") {
            defaultBranch = "master";
          }
        }

        if (this.state.selectedBranch === "") this.branchChange(defaultBranch)
        return data;
      })
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_BRANCHES, payload: {
            owner: owner,
            repo: repo,
            branches: data
          }
        });
      }, () => {/* Generic error handler deals with it */
    });
  }

  refreshBranches(owner, repo) {
    this.props.gimletClient.getBranches(owner, repo)
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_BRANCHES, payload: {
            owner: owner,
            repo: repo,
            branches: data
          }
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  refreshCommits(owner, repo, branch) {
    this.props.gimletClient.getCommits(owner, repo, branch, "head")
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_COMMITS, payload: {
            owner: owner,
            repo: repo,
            commits: data
          }
        });
      }, () => {/* Generic error handler deals with it */
      });
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

  branchChange(newBranch) {
    if (newBranch === '') {
      return
    }

    const { owner, repo } = this.props.match.params;
    const { selectedBranch } = this.state;

    if (newBranch !== selectedBranch) {
      this.setState({ selectedBranch: newBranch });

      this.props.gimletClient.getCommits(owner, repo, newBranch, "head")
        .then(data => {
          this.props.store.dispatch({
            type: ACTION_TYPE_COMMITS, payload: {
              owner: owner,
              repo: repo,
              commits: data
            }
          });
        }, () => {/* Generic error handler deals with it */
        });
    }
  }

  checkDeployStatus(trackingId) {
    const { owner, repo } = this.props.match.params;

    this.props.gimletClient.getDeployStatus(trackingId)
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_DEPLOY_STATUS, payload: {
            trackingId: trackingId,
            status: data.status,
            statusDesc: data.statusDesc,
            results: data.results,
          }
        });

        if (data.status === "new") {
          setTimeout(() => {
            this.checkDeployStatus(trackingId);
          }, 1000);
        }

        if (data.status === "processed") {
          let gitopsCommitsApplied = true;
          const numberOfResults = data.results.length;

          if (numberOfResults > 0) {
            const latestGitopsHashMetadata = data.results[0];
            if (latestGitopsHashMetadata.gitopsCommitStatus === "N/A") { // poll until all gitops writes are applied
              gitopsCommitsApplied = false;
              setTimeout(() => {
                this.checkDeployStatus(trackingId);
              }, 1000);
            }
          }
          if (gitopsCommitsApplied) {
            for (const result of data.results) {
              setTimeout(() => {
                this.props.gimletClient.getRolloutHistoryPerApp(owner, repo, result.env, result.app)
                .then(data => {
                    this.props.store.dispatch({
                      type: ACTION_TYPE_ROLLOUT_HISTORY, payload: {
                        owner: owner,
                        repo: repo,
                        env: result.env,
                        app: result.app,
                        releases: data,
                      }
                    });
                  }, () => {/* Generic error handler deals with it */ }
                  );

                this.props.gimletClient.getReleases(result.env, 10)
                  .then(data => {
                    this.props.store.dispatch({
                      type: ACTION_TYPE_RELEASE_STATUSES,
                      payload: {
                        envName: result.env,
                        data: data,
                      }
                    });
                  }, () => {/* Generic error handler deals with it */
                  })
                }, 300);
            }
          }
        }
      }, () => {/* Generic error handler deals with it */
      });
  }

  checkImageBuildStatus(trackingId) {
    this.props.gimletClient.getDeployStatus(trackingId)
      .then(data => {
        const triggeredDeployRequestID = data.results && data.results.length > 0 ? data.results[0].triggeredDeployRequestID : undefined
        this.props.store.dispatch({
          type: ACTION_TYPE_IMAGEBUILD_STATUS, payload: {
            triggeredDeployRequestID: triggeredDeployRequestID,
            status: data.status,
            statusDesc: data.statusDesc,
            results: data.results,
          }
        });

        if (data.status === "new") {
          setTimeout(() => {
            this.checkImageBuildStatus(trackingId);
          }, 2000);
        }

        if (data.type === "imageBuild" && data.status === "success") {
          const triggeredReleaseId = data.results[0].triggeredDeployRequestID
          this.checkDeployStatus(triggeredReleaseId);
        }
      }, () => {/* Generic error handler deals with it */
      });
  }

  deploy(target, sha, repo) {
    this.setState({deployStatusModal: true});
    this.props.gimletClient.deploy(target.artifactId, target.env, target.app, this.state.selectedTenant)
      .then(data => {
        const trackingId = data.id
        if (data.type === 'imageBuild') {
          this.props.store.dispatch({
            type: ACTION_TYPE_IMAGEBUILD, payload: {
              trackingId: trackingId
            }
          });
          this.props.store.dispatch({
            type: ACTION_TYPE_DEPLOY, payload: {
              repo: repo,
              env: target.env,
              app: target.app,
              sha: sha
            }
          });
          setTimeout(() => {
            this.checkImageBuildStatus(trackingId);
          }, 2000);
        } else {
          this.props.store.dispatch({type: ACTION_TYPE_IMAGEBUILD, payload: undefined});
          this.props.store.dispatch({
            type: ACTION_TYPE_DEPLOY, payload: {
              trackingId: trackingId,
              repo: repo,
              env: target.env,
              app: target.app,
              sha: sha
            }
          });
          setTimeout(() => {
            this.checkDeployStatus(trackingId);
          }, 1000);
        }
      }, () => {/* Generic error handler deals with it */
      });
  }

  triggerCommitSync(owner, repo) {
    this.props.gimletClient.triggerCommitSync(owner, repo)
  }

  getPullRequests(owner, repo) {
    this.props.gimletClient.getPullRequests(owner, repo)
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_REPO_PULLREQUESTS, payload: {
            data: data,
            repoName: `${owner}/${repo}`
          }
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  rollback(env, app, rollbackTo, e) {
    this.setState({deployStatusModal: true});
    const target = {
      rollback: true,
      app: app,
      env: env,
    };
    this.props.gimletClient.rollback(env, app, rollbackTo)
      .then(data => {
        const trackingId = data.id;
        this.props.store.dispatch({
          type: ACTION_TYPE_DEPLOY, payload: {
            rollback: true,
            trackingId: trackingId,
            // repo: repo,
            env: target.env,
            app: target.app,
          }
        });
        setTimeout(() => {
          this.checkDeployStatus(trackingId);
        }, 1000);
      }, () => {/* Generic error handler deals with it */
      });
  }

  navigateToConfigEdit(env, config) {
    const { owner, repo } = this.props.match.params;
    window.location.replace(encodeURI(`/repo/${owner}/${repo}/envs/${env}/config/${config}`));
  }

  linkToDeployment(env, deployment) {
    const { owner, repo } = this.props.match.params;
    this.props.history.push({
      pathname: `/repo/${owner}/${repo}/${env}/${deployment}`,
      search: this.props.location.search
    })
  }

  newConfig(env, config) {
    const { owner, repo } = this.props.match.params;
    this.props.history.push(encodeURI(`/repo/${owner}/${repo}/envs/${env}/config/${config}/new`));
  }

  fileMetasByEnv(envName) {
    return this.state.fileInfos.filter(fileInfo => fileInfo.envName === envName)
  }

  setSelectedTenant(tenant) {
    this.setState({ selectedTenant: tenant });
    const queryParam = tenant === "" ? tenant : `?tenant=${tenant}`

    this.props.history.push({
      pathname: this.props.location.pathname,
      search: queryParam
    })
  }

  tenantsFromConfigs(envConfigs) {
    let tenants = [];

    if (!envConfigs) {
      return tenants;
    }

    for (const configs of Object.values(envConfigs)) {
      configs.forEach(config => {
        const tenantName = config.tenant.name;
        if (tenantName && !tenants.includes(tenantName)) {
          tenants.push(tenantName);
        }
      });
    }
    return tenants;
  }

  filteredConfigsByTenant(envConfigs, selectedTenant) {
    if (!envConfigs || !selectedTenant) {
      return envConfigs;
    }

    const filteredEnvs = envConfigs.filter(envConfig => envConfig.tenant.name === selectedTenant);

    if (filteredEnvs.length === 0) {
      return undefined;
    }

    return filteredEnvs;
  }

  setUnselectedEnvs(unselectedEnvs) {
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;
    localStorage.setItem(repoName + "-unselected-envs", JSON.stringify(unselectedEnvs))
    this.setState({
      unselectedEnvs: unselectedEnvs
    })
  }

  render() {
    const { owner, repo, environment, deployment } = this.props.match.params;
    const repoName = `${owner}/${repo}`
    let { envs, connectedAgents, rolloutHistory, commits, pullRequests, settings, deployStatusModal, unselectedEnvs } = this.state;
    const { branches, selectedBranch, envConfigs, scmUrl, alerts } = this.state;

    let decoratedEnvs = envsForRepo(envs, connectedAgents, repoName);

    let repoRolloutHistory = undefined;
    if (rolloutHistory && rolloutHistory[repoName]) {
      repoRolloutHistory = rolloutHistory[repoName]
    }

    return (
      <div>
        {deployStatusModal && envConfigs !== undefined && 
        <DeployStatusModal
          closeHandler={()=> this.setState({deployStatusModal: false})}
          envs={decoratedEnvs}
          owner={owner}
          repoName={repo}
          envConfigs={envConfigs}
          scmUrl={scmUrl}
          store={this.props.store}
          gimletClient={this.props.gimletClient}
        />
        }
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className='flex-1'>
              <div className="flex justify-between">
                <h1 className="text-3xl font-bold leading-tight text-gray-900">{repoName}
                  <a href={`${scmUrl}/${owner}/${repo}`} target="_blank" rel="noopener noreferrer">
                    <svg xmlns="http://www.w3.org/2000/svg"
                      className="inline fill-current text-gray-500 hover:text-gray-700 ml-1 h-4 w-4"
                      viewBox="0 0 24 24">
                      <path d="M0 0h24v24H0z" fill="none" />
                      <path
                        d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                    </svg>
                  </a>
                  {/* {this.ciConfigAndShipperStatuses(repoName)} */}
                </h1>
                <RefreshButton
                  refreshFunc={() => {
                    this.triggerCommitSync(owner, repo);
                    this.getPullRequests(owner, repo);
                  }}
                />
              </div>
              <button className="text-gray-500 hover:text-gray-700" onClick={() => this.props.history.push("/repositories")}>
                &laquo; back
              </button>
            </div>
            <TenantSelector
              tenants={this.tenantsFromConfigs(envConfigs)}
              selectedTenant={this.state.selectedTenant}
              setSelectedTenant={this.setSelectedTenant}
            />
            {decoratedEnvs && <div className='py-4'>
              <EnvSelector
                envs={decoratedEnvs}
                unselectedEnvs={unselectedEnvs}
                setUnselectedEnvs={this.setUnselectedEnvs}
              />
            </div>
            }
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 sm:px-0">
              <div>
                {envConfigs && Object.keys(decoratedEnvs).sort().map((envName) =>
                  {
                    const unselected = unselectedEnvs.includes(envName)

                    return unselected ? null : (
                    <Env
                      key={envName}
                      env={decoratedEnvs[envName]}
                      repoRolloutHistory={repoRolloutHistory}
                      envConfigs={this.filteredConfigsByTenant(envConfigs[envName], this.state.selectedTenant)}
                      navigateToConfigEdit={this.navigateToConfigEdit}
                      linkToDeployment={this.linkToDeployment}
                      newConfig={this.newConfig}
                      rollback={this.rollback}
                      owner={owner}
                      repoName={repo}
                      fileInfos={this.fileMetasByEnv(envName)}
                      pullRequests={pullRequests?.[envName]}
                      releaseHistorySinceDays={settings.releaseHistorySinceDays}
                      gimletClient={this.props.gimletClient}
                      store={this.props.store}
                      envFromParams={environment}
                      deploymentFromParams={deployment}
                      scmUrl={scmUrl}
                      history={this.props.history}
                      alerts={alerts}
                    />)}
                )}

                {Object.keys(branches).length !== 0 &&
                  <div className="bg-gray-50 rounded-lg p-4 sm:p-6 lg:p-8 mt-8 relative">
                    <div className="w-64 mb-4 lg:mb-8">
                      <Dropdown
                        items={branches[repoName]}
                        value={selectedBranch}
                        changeHandler={(newBranch) => this.branchChange(newBranch)}
                      />
                    </div>
                    {commits &&
                      <Commits
                        commits={commits[repoName]}
                        envs={envs}
                        connectedAgents={decoratedEnvs}
                        deployHandler={this.deploy}
                        repo={repo}
                        gimletClient={this.props.gimletClient}
                        store={this.props.store}
                        owner={owner}
                        branch={this.state.selectedBranch}
                        scmUrl={scmUrl}
                        tenant={this.state.selectedTenant}
                      />
                    }
                  </div>}
                {(!envConfigs || !commits) && <Spinner />}
              </div>
            </div>
          </div>
        </main>
      </div>
    )
  }
}

/*
  Takes all envs from Kubernetes
  and finds the relevant stacks for the repo for each environment
  then filters the relevant stacks further with the search box filter
*/
function stacks(connectedAgents, envName) {
  for (const agentName of Object.keys(connectedAgents)) {
    const agent = connectedAgents[agentName];
    if (agentName === envName) {
      return agent.stacks;
    }
  }
  return [];
}

function envsForRepo(envs, connectedAgents, repoName) {
  let decoratedEnvs = {};

  if (!connectedAgents || !envs) {
    return decoratedEnvs;
  }
  
  // iterate through all Kubernetes envs
  for (const env of envs) {
    decoratedEnvs[env.name] = {
      name: env.name,
      builtIn: env.builtIn,
      isOnline: isOnline(connectedAgents, env)
    };

    // find all stacks that belong to this repo
    decoratedEnvs[env.name].stacks = stacks(connectedAgents, env.name).filter((service) => {
      return service.repo === repoName
    });
  }

  return decoratedEnvs;
}

function isOnline(onlineEnvs, singleEnv) {
  return Object.keys(onlineEnvs)
      .map(env => onlineEnvs[env])
      .some(onlineEnv => {
          return onlineEnv.name === singleEnv.name
      })
};

function EnvSelector(props) {
  const { envs, unselectedEnvs, setUnselectedEnvs } = props

  return (
    <div className='space-x-1'>
    {envs && Object.keys(envs).sort().map(envName => {
      const unselected = unselectedEnvs.includes(envName)

      return (
        <button key={envName} className={(unselected ? "bg-gray-50 text-gray-600" : "text-blue-50 bg-blue-600") + " select-none capitalize rounded-full px-3"}
        onClick={() => unselected
          ? setUnselectedEnvs(unselectedEnvs.filter(i => i !== envName))
          : unselectedEnvs.push(envName) && setUnselectedEnvs(unselectedEnvs)}
        >
          {envName}
        </button>
      )
      })}
    </div>
  )
}
