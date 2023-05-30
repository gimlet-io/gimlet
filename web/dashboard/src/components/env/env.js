import React, { Component } from "react";
import { v4 as uuidv4 } from 'uuid';
import ServiceDetail from "../serviceDetail/serviceDetail";
import { parseDeploymentName } from "../../views/pulse/pulse";

export class Env extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isClosed: localStorage.getItem([this.props.envName]) === "true"
    }
  }

  componentDidMount() {
    if (this.props.envName === this.props.envFromParams) {
      this.setState({ isClosed: false });
    }
  }

  render() {
    const { searchFilter, envName, env, repoRolloutHistory, envConfigs, navigateToConfigEdit, linkToDeployment, newConfig, rollback, owner, repoName, fileInfos, pullRequests, releaseHistorySinceDays, gimletClient, store, kubernetesAlerts, deploymentFromParams, scmUrl } = this.props;

    const renderedServices = renderServices(env.stacks, envConfigs, envName, repoRolloutHistory, navigateToConfigEdit, linkToDeployment, rollback, owner, repoName, fileInfos, releaseHistorySinceDays, gimletClient, store, kubernetesAlerts, deploymentFromParams, scmUrl);

    return (
      <div>
        <h4 className="flex items-stretch select-none text-xl font-medium capitalize leading-tight text-gray-900 my-4">
          {envName}
          <svg
            onClick={() => {
              localStorage.setItem(this.props.envName, !this.state.isClosed);
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
          <>
            {renderPullRequests(pullRequests)}
            <div className="bg-white shadow p-4 sm:p-6 lg:p-8 space-y-4">
              {renderedServices.length === 10 &&
              <span className="text-xs text-blue-700">Displaying at most 10 application configurations per environment.</span>}
              {renderedServices.length > 0
                ?
                <>
                  {renderedServices}
                  <h4 className="text-xs cursor-pointer text-gray-500 hover:text-gray-700"
                    onClick={() => {
                      const newAppName = `${repoName}-${uuidv4().slice(0, 4)}`
                      newConfig(envName, newAppName)
                    }}>
                    Add app config
                  </h4>
                </>
                : emptyState(searchFilter, envConfigs, newConfig, envName, repoName)}

            </div>
          </>
        )}
      </div>
    )
  }
}

function renderServices(stacks, envConfigs, envName, repoRolloutHistory, navigateToConfigEdit, linkToDeployment, rollback, owner, repoName, fileInfos, releaseHistorySinceDays, gimletClient, store, kubernetesAlerts, deploymentFromParams, scmUrl) {
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
        rolloutHistory={repoRolloutHistory?.[envName]?.[stack.service.name]}
        rollback={rollback}
        envName={envName}
        owner={owner}
        repoName={repoName}
        fileName={fileName(fileInfos, stack.service.name)}
        navigateToConfigEdit={navigateToConfigEdit}
        linkToDeployment={linkToDeployment}
        configExists={configExists}
        releaseHistorySinceDays={releaseHistorySinceDays}
        gimletClient={gimletClient}
        store={store}
        kubernetesAlerts={kubernetesAlertsByDeploymentName(kubernetesAlerts, stack.service.name)}
        deploymentFromParams={deploymentFromParams}
        scmUrl={scmUrl}
      />
    )
  })

  if (services.length >= 10) {
    return services.slice(0, 10);
  }

  const configsWeHaventDeployed = configsWeHave.filter(config => !configsWeDeployed.includes(config));

  services.push(
    ...configsWeHaventDeployed.map(config => {
      return <ServiceDetail
        key={config}
        stack={{
          service: {
            name: config
          }
        }}
        rolloutHistory={repoRolloutHistory?.[envName]?.[config]}
        rollback={rollback}
        envName={envName}
        owner={owner}
        repoName={repoName}
        fileName={fileName(fileInfos, config)}
        navigateToConfigEdit={navigateToConfigEdit}
        linkToDeployment={linkToDeployment}
        configExists={true}
        releaseHistorySinceDays={releaseHistorySinceDays}
        gimletClient={gimletClient}
        store={store}
        kubernetesAlerts={kubernetesAlertsByDeploymentName(kubernetesAlerts, config)}
        deploymentFromParams={deploymentFromParams}
        scmUrl={scmUrl}
      />
    }
    )
  )
  return services.slice(0, 10)
}

function kubernetesAlertsByDeploymentName(kubernetesAlerts, deploymentName) {
  return kubernetesAlerts.filter(event => parseDeploymentName(event.deploymentName) === deploymentName)
}

function fileName(fileInfos, appName) {
  if (fileInfos.find(fileInfo => fileInfo.appName === appName)) {
    return fileInfos.find(fileInfo => fileInfo.appName === appName).fileName;
  }
}

function emptyState(searchFilter, envConfigs, newConfig, envName, repoName) {
  if (!envConfigs) {
    return emptyStateDeployThisRepo(newConfig, envName, repoName);
  }

  if (searchFilter !== '') {
    return emptyStateSearch()
  }
}

function emptyStateSearch() {
  return <p className="text-xs text-gray-800">No service matches the search</p>
}

function emptyStateDeployThisRepo(newConfig, envName, repoName) {
  return <div
    target="_blank"
    rel="noreferrer"
    onClick={() => {
      newConfig(envName, repoName)
    }}
    className="relative block w-full border-2 border-gray-300 border-dashed rounded-lg p-6 text-center hover:border-gray-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 cursor-pointer"
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className="mx-auto h-12 w-12 text-gray-400"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
    <div className="mt-2 block text-sm font-bold text-gray-500">
      Deploy this repository to <span className="capitalize">{envName}</span>
    </div>
  </div>
}

export function renderPullRequests(pullRequests) {
  if (!pullRequests || pullRequests.length === 0) {
    return null
  }

  return (
    <div className="bg-indigo-600 rounded-t-lg">
      <div className="text-white inline-grid items-center mx-auto py-3 px-3 sm:px-6 lg:px-8">
        <span className="font-bold text-sm">Pull Requests:</span>
        <ul className="list-disc list-inside text-xs ml-2">
          {pullRequests.map(pullRequest =>
            <li key={pullRequest.sha}>
              <a href={pullRequest.link} target="_blank" rel="noopener noreferrer">
                {`#${pullRequest.number} ${pullRequest.title}`}
              </a>
            </li>)}
        </ul>
      </div>
    </div>
  )
};
