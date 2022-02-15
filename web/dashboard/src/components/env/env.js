import React, { Component } from "react";
import ServiceDetail from "../serviceDetail/serviceDetail";

export class Env extends Component {

  constructor(props) {
    super(props);
    this.state = {
      isClosed: false
    }
  }

  render() {
    const { searchFilter, envName, env, repoRolloutHistory, envConfigs, navigateToConfigEdit, rollback, repoName} = this.props;

    const renderedServices = renderServices(env.stacks, envConfigs, envName, repoRolloutHistory, navigateToConfigEdit, rollback);

    return (
      <div>
        <h4 className="flex items-stretch select-none text-xl font-medium capitalize leading-tight text-gray-900 my-4">
          {envName}
          <svg
            onClick={() => {
              this.setState((prevState) => {
                return {
                  isClosed: !prevState.isClosed
                }
              })
            }}

            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 cursor-pointer"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d={
                this.state.isClosed
                  ? "M9 5l7 7-7 7"
                  : "M19 9l-7 7-7-7"
              }
            />
          </svg>
        </h4>
        {this.state.isClosed ? null : (
          <div class="bg-white shadow divide-y divide-gray-200 p-4 sm:p-6 lg:p-8">
            {renderedServices.length > 0
              ? renderedServices
              : emptyState(searchFilter, envConfigs, navigateToConfigEdit, envName, repoName)}
          </div>
        )}
      </div>
    )
  }
}

function renderServices(stacks, envConfigs, envName, repoRolloutHistory, navigateToConfigEdit, rollback) {
  let services = [];

  let configsWeHave = [];
  if (envConfigs) {
    configsWeHave = envConfigs.map((config) => config.app);
  }

  let configsWeDeployed = [];
  // render services that are deployed on k8s
  services = stacks.map((stack) => {
    configsWeDeployed.push(stack.service.name);
    const configExists = configsWeHave.includes(stack.service.name)
    return (
      <ServiceDetail
        key={stack.service.name}
        stack={stack}
        rolloutHistory={appRolloutHistory(envName, stack.service.name, repoRolloutHistory)}
        rollback={rollback}
        envName={envName}
        navigateToConfigEdit={navigateToConfigEdit}
        configExists={configExists}
      />
    )
  })

  const configsWeHaventDeployed = configsWeHave.filter(config => !configsWeDeployed.includes(config));

  services.push(
    ...configsWeHaventDeployed.map(config => {
      return <ServiceDetail
        key={config}
        stack={{service: {
          name: config
        }}}
        rolloutHistory={appRolloutHistory(envName, config, repoRolloutHistory)}
        rollback={rollback}
        envName={envName}
        navigateToConfigEdit={navigateToConfigEdit}
        configExists={true}
      />
    }
    )
  )
  return services
}

function appRolloutHistory(envName, appName, repoRolloutHistory) {
  if (repoRolloutHistory) {
    let envRolloutHistory = repoRolloutHistory.find(env => env.name === envName)

    if (envRolloutHistory) {
      let appRolloutHistory = envRolloutHistory.apps.find(app => app.name === appName)
      return appRolloutHistory
    }
  }

  return []
}

function emptyState(searchFilter, envConfigs, navigateToConfigEdit, envName, repoName) {
  if (searchFilter !== '') {
    return emptyStateSearch()
  } else {
    if (!envConfigs) {
      return emptyStateDeployThisRepo(navigateToConfigEdit, envName, repoName);
    }
  }
}

function emptyStateSearch() {
  return <p className="text-xs text-gray-800">No service matches the search</p>
}

function emptyStateDeployThisRepo(navigateToConfigEdit, envName, repoName) {
  return <div
    target="_blank"
    rel="noreferrer"
    onClick={() => navigateToConfigEdit(envName, envName)}
    className="relative block w-full border-2 border-gray-300 border-dashed rounded-lg p-6 text-center hover:border-gray-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 cursor-pointer"
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className="mx-auto h-12 w-12 text-gray-400"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
    <div className="mt-2 block text-sm font-bold text-gray-500">
      Deploy this repository to <span className="capitalize">{envName}</span>
    </div>
  </div>
}
