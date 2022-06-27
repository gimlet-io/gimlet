import React, { Component } from 'react';

import {
  ACTION_TYPE_BRANCHES,
  ACTION_TYPE_ENVCONFIGS,
  ACTION_TYPE_COMMITS,
  ACTION_TYPE_DEPLOY,
  ACTION_TYPE_DEPLOY_STATUS,
  ACTION_TYPE_REPO_METAS,
  ACTION_TYPE_ROLLOUT_HISTORY,
  ACTION_TYPE_ADD_ENVCONFIG,
} from "../../redux/redux";
import { Commits } from "../../components/commits/commits";
import Dropdown from "../../components/dropdown/dropdown";
import { emptyStateNoAgents } from "../services/services";
import { Env } from '../../components/env/env';

export default class Repo extends Component {
  constructor(props) {
    super(props);
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      connectedAgents: reduxState.connectedAgents,
      search: reduxState.search,
      rolloutHistory: reduxState.rolloutHistory,
      commits: reduxState.commits,
      branches: reduxState.branches,
      envConfigs: reduxState.envConfigs[repoName],
      selectedBranch: '',
      settings: reduxState.settings,
      refreshQueue: reduxState.repoRefreshQueue.filter(repo => repo === repoName).length,
      agents: reduxState.settings.agents,
      envs: reduxState.envs,
      repoMetas: reduxState.repoMetas,
      fileInfos: reduxState.fileInfos,
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        connectedAgents: reduxState.connectedAgents,
        search: reduxState.search,
        rolloutHistory: reduxState.rolloutHistory,
        commits: reduxState.commits,
        branches: reduxState.branches,
        envConfigs: reduxState.envConfigs[repoName],
        envs: reduxState.envs,
        repoMetas: reduxState.repoMetas,
        fileInfos: reduxState.fileInfos,
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
    this.newConfig = this.newConfig.bind(this)
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

    this.props.gimletClient.getRolloutHistory(owner, repo)
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_ROLLOUT_HISTORY, payload: {
            owner: owner,
            repo: repo,
            releases: data
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

    this.props.gimletClient.getBranches(owner, repo)
      .then(data => {
        let defaultBranch = 'main'
        for (let branch of data) {
          if (branch === "master") {
            defaultBranch = "master";
          }
        }

        this.branchChange(defaultBranch)
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
    this.props.gimletClient.getCommits(owner, repo, branch)
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

      this.props.gimletClient.getCommits(owner, repo, newBranch)
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

  checkDeployStatus(deployRequest) {
    const { owner, repo } = this.props.match.params;

    this.props.gimletClient.getDeployStatus(deployRequest.trackingId)
      .then(data => {
        deployRequest.status = data.status;
        deployRequest.statusDesc = data.statusDesc;
        deployRequest.gitopsHashes = data.gitopsHashes;
        this.props.store.dispatch({
          type: ACTION_TYPE_DEPLOY_STATUS, payload: deployRequest
        });

        if (data.status === "new") {
          setTimeout(() => {
            this.checkDeployStatus(deployRequest);
          }, 500);
        }

        if (data.status === "processed") {
          let gitopsCommitsApplied = true;
          const numberOfGitopsHashes = data.gitopsHashes.length;
          if (numberOfGitopsHashes > 0) {
            const latestGitopsHashMetadata = data.gitopsHashes[0];
            if (latestGitopsHashMetadata.status === 'N/A') { // poll until all gitops writes are applied
              gitopsCommitsApplied = false;
              setTimeout(() => {
                this.checkDeployStatus(deployRequest);
              }, 500);
            }
          }
          if (gitopsCommitsApplied) {
            this.props.gimletClient.getRolloutHistory(owner, repo)
              .then(data => {
                this.props.store.dispatch({
                  type: ACTION_TYPE_ROLLOUT_HISTORY, payload: {
                    owner: owner,
                    repo: repo,
                    releases: data
                  }
                });
              }, () => {/* Generic error handler deals with it */
              });
          }
        }
      }, () => {/* Generic error handler deals with it */
      });
  }

  deploy(target, sha, repo) {
    this.props.gimletClient.deploy(target.artifactId, target.env, target.app)
      .then(data => {
        target.sha = sha;
        target.trackingId = data.trackingId;
        setTimeout(() => {
          this.checkDeployStatus(target);
        }, 500);
      }, () => {/* Generic error handler deals with it */
      });

    target.sha = sha;
    target.repo = repo;
    this.props.store.dispatch({
      type: ACTION_TYPE_DEPLOY, payload: target
    });
  }

  rollback(env, app, rollbackTo, e) {
    const target = {
      rollback: true,
      app: app,
      env: env,
    };
    this.props.gimletClient.rollback(env, app, rollbackTo)
      .then(data => {
        target.trackingId = data.trackingId;
        setTimeout(() => {
          this.checkDeployStatus(target);
        }, 500);
      }, () => {/* Generic error handler deals with it */
      });

    this.props.store.dispatch({
      type: ACTION_TYPE_DEPLOY, payload: target
    });
  }

  navigateToConfigEdit(env, config) {
    const { owner, repo } = this.props.match.params;
    this.props.history.push(`/repo/${owner}/${repo}/envs/${env}/config/${config}`);
  }

  newConfig(env, config) {
    const { owner, repo } = this.props.match.params;

    this.props.store.dispatch({
      type: ACTION_TYPE_ADD_ENVCONFIG, payload: {
        repo: `${owner}/${repo}`,
        env: env,
        envConfig: {
          app: config,
          env: env
        },
      }
    });
    this.props.history.push(`/repo/${owner}/${repo}/envs/${env}/config/${config}/new`);
  }

  ciConfigAndShipperStatuses(repoName) {
    const { repoMetas } = this.state;
    let shipperColor = "text-gray-500";
    let ciConfigColor = "text-gray-500";
    let ciConfig = "";

    if (repoMetas.githubActions) {
      ciConfigColor = "text-green-500";
      ciConfig = ".github/workflows"
    } else if (repoMetas.circleCi) {
      ciConfigColor = "text-green-500";
      ciConfig = ".circleci"
    }
    if (repoMetas.hasShipper) {
      shipperColor = "text-green-500";
    }

    return (
      <>
        {repoMetas.githubActions || repoMetas.circleCi ?
          <a href={`https://github.com/${repoName}/tree/main/${ciConfig}`} target="_blank" rel="noopener noreferrer">
            <svg xmlns="http://www.w3.org/2000/svg" className={`inline ml-1 h-4 w-4 ${ciConfigColor}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              <title>{repoMetas.githubActions || repoMetas.circleCi ? "This repository has CI config" : "This repository doesn't have CI config"}</title>
            </svg>
          </a>
          :
          <svg xmlns="http://www.w3.org/2000/svg" className={`inline ml-1 h-4 w-4 ${ciConfigColor}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            <title>{repoMetas.githubActions || repoMetas.circleCi ? "This repository has CI config" : "This repository doesn't have CI config"}</title>
          </svg>}
        <svg xmlns="http://www.w3.org/2000/svg" className={`inline ml-1 h-4 w-4 ${shipperColor}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
          <title>{repoMetas.hasShipper ? "This repository has shipper" : "This repository doesn't have shipper"}</title>
        </svg>
      </>)
  }

  fileMetasByEnv(envName) {
    return this.state.fileInfos.filter(fileInfo => fileInfo.envName === envName)
  }

  render() {
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`
    let { envs, connectedAgents, search, rolloutHistory, commits, agents } = this.state;
    const { branches, selectedBranch, envConfigs } = this.state;

    let filteredEnvs = envsForRepoFilteredBySearchFilter(envs, connectedAgents, repoName, search.filter);

    let repoRolloutHistory = undefined;
    if (rolloutHistory && rolloutHistory[repoName]) {
      repoRolloutHistory = rolloutHistory[repoName]
    }

    if (!this.state.envConfigs) {
      return null
    }

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <h1 className="text-3xl font-bold leading-tight text-gray-900">{repoName}
              <a href={`https://github.com/${owner}/${repo}`} target="_blank" rel="noopener noreferrer">
                <svg xmlns="http://www.w3.org/2000/svg"
                  className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
                  viewBox="0 0 24 24">
                  <path d="M0 0h24v24H0z" fill="none" />
                  <path
                    d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                </svg>
              </a>
              {this.ciConfigAndShipperStatuses(repoName)}
            </h1>
            <button className="text-gray-500 hover:text-gray-700" onClick={() => this.props.history.goBack()}>
              &laquo; back
            </button>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 py-8 sm:px-0">
              <div>
                {agents.length === 0 &&
                  <div className="mt-8 mb-16">
                    {emptyStateNoAgents()}
                  </div>
                }

                {Object.keys(filteredEnvs).sort().map((envName) =>
                  <Env
                    searchFilter={search.filter}
                    envName={envName}
                    env={filteredEnvs[envName]}
                    repoRolloutHistory={repoRolloutHistory}
                    envConfigs={envConfigs[envName]}
                    navigateToConfigEdit={this.navigateToConfigEdit}
                    newConfig={this.newConfig}
                    rollback={this.rollback}
                    owner={owner}
                    repoName={repo}
                    fileInfos={this.fileMetasByEnv(envName)}
                  />
                )
                }

                <div className="bg-gray-50 shadow p-4 sm:p-6 lg:p-8 mt-8 relative">
                  <div className="w-64 mb-4 lg:mb-8">
                    {branches &&
                      <Dropdown
                        items={branches[repoName]}
                        value={selectedBranch}
                        changeHandler={(newBranch) => this.branchChange(newBranch)}
                      />
                    }
                  </div>
                  {commits &&
                    <Commits
                      commits={commits[repoName]}
                      connectedAgents={filteredEnvs}
                      rolloutHistory={repoRolloutHistory}
                      deployHandler={this.deploy}
                      repo={repoName}
                    />
                  }
                </div>
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

function envsForRepoFilteredBySearchFilter(envs, connectedAgents, repoName, searchFilter) {
  let filteredEnvs = {};

  // iterate through all Kubernetes envs
  for (const env of envs) {
    filteredEnvs[env.name] = { name: env.name };

    // find all stacks that belong to this repo
    filteredEnvs[env.name].stacks = stacks(connectedAgents, env.name).filter((service) => {
      return service.repo === repoName
    });

    // applpy search box filter
    if (searchFilter !== '') {
      filteredEnvs[env.name].stacks = filteredEnvs[env.name].stacks.filter((service) => {
        return service.service.name.includes(searchFilter) ||
          (service.deployment !== undefined && service.deployment.name.includes(searchFilter)) ||
          (service.ingresses !== undefined && service.ingresses.filter((ingress) => ingress.url.includes(searchFilter)).length > 0)
      })
    }
  }

  return filteredEnvs;
}
