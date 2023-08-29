import React, { Component } from "react";
import { v4 as uuidv4 } from 'uuid';
import ServiceDetail from "../serviceDetail/serviceDetail";
import { parseDeploymentName } from "../../views/pulse/pulse";
import { InformationCircleIcon } from '@heroicons/react/solid'

export class Env extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isClosed: localStorage.getItem([this.props.env.name]) === "true"
    }
  }

  componentDidMount() {
    if (this.props.env.name === this.props.envFromParams) {
      this.setState({ isClosed: false });
    }
  }

  render() {
    const { env, repoRolloutHistory, envConfigs, navigateToConfigEdit, linkToDeployment, newConfig, rollback, owner, repoName, fileInfos, pullRequests, releaseHistorySinceDays, gimletClient, store, kubernetesAlerts, deploymentFromParams, scmUrl, history } = this.props;

    const renderedServices = renderServices(env.stacks, envConfigs, env.name, repoRolloutHistory, navigateToConfigEdit, linkToDeployment, rollback, owner, repoName, fileInfos, releaseHistorySinceDays, gimletClient, store, kubernetesAlerts, deploymentFromParams, scmUrl, env.builtIn);

    return (
      <div>
        <h4 className="relative flex items-stretch select-none text-xl font-medium capitalize leading-tight text-gray-900 my-4">
          {env.name}
          <span title={env.isOnline ? "Connected" : "Disconnected"}>
            <svg className={(env.isOnline ? "text-green-400" : "text-red-400") + " inline fill-current ml-1"} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
              <path
                d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
              />
            </svg>
          </span>
          <svg
            onClick={() => {
              localStorage.setItem(this.props.env.name, !this.state.isClosed);
              this.setState((prevState) => {
                return {
                  isClosed: !prevState.isClosed
                }
              })
            }}

            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 cursor-pointer absolute right-0"
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
              {!env.isOnline && connectEnvCard(history)}
              {renderedServices.length === 10 &&
                <span className="text-xs text-blue-700">Displaying at most 10 application configurations per environment.</span>
              }
              { env.isOnline && renderedServices.length !== 0 &&
                <>
                  {renderedServices}
                  <h4 className="text-xs cursor-pointer text-gray-500 hover:text-gray-700"
                    onClick={() => {
                      const newAppName = `${repoName}-${uuidv4().slice(0, 4)}`
                      newConfig(env.name, newAppName)
                    }}>
                    New deployment configuration
                  </h4>
                </>
              }
              { env.isOnline && renderedServices.length === 0 && emptyStateDeployThisRepo(newConfig, env.name, repoName) }
            </div>
          </>
        )}
      </div>
    )
  }
}

function renderServices(
  stacks,
  envConfigs,
  envName,
  repoRolloutHistory,
  navigateToConfigEdit,
  linkToDeployment,
  rollback,
  owner,
  repoName,
  fileInfos,
  releaseHistorySinceDays,
  gimletClient,
  store,
  kubernetesAlerts,
  deploymentFromParams,
  scmUrl,
  builtInEnv) {
  let services = [];

  let configsWeHave = [];
  if (envConfigs) {
    configsWeHave = envConfigs.map((config) => config.app);
  }

  const filteredStacks = stacks;//.filter(stack => configsWeHave.includes(stack.service.name));

  let configsWeDeployed = [];
  // render services that are deployed on k8s
  services = filteredStacks.map((stack) => {
    configsWeDeployed.push(stack.service.name);
    const configExists = configsWeHave.includes(stack.service.name)
    let config = undefined;
    if (configExists) {
      config = envConfigs.find((config) => config.app === stack.service.name)
    }

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
        config={config}
        releaseHistorySinceDays={releaseHistorySinceDays}
        gimletClient={gimletClient}
        store={store}
        kubernetesAlerts={kubernetesAlertsByDeploymentName(kubernetesAlerts, stack.service.name)}
        deploymentFromParams={deploymentFromParams}
        scmUrl={scmUrl}
        builtInEnv={builtInEnv}
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

function connectEnvCard(history) {
  return (
    <div className="rounded-md bg-blue-50 p-4">
    <div className="flex">
      <div className="flex-shrink-0">
        <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
      </div>
      <div className="ml-3">
        <h3 className="text-sm font-medium text-blue-800">Environment disconnected</h3>
        <div className="mt-2 text-sm text-blue-700">
          This environment is disconnected.<br />
          <button
            className="font-medium"
            onClick={() => {history.push("/environments");return true}}
          >
            Click to connect this environment to a cluster on the Environments page.
          </button>
        </div>
      </div>
    </div>
    </div>
  );
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

function emptyStateDeployThisRepo(newConfig, envName, repoName) {
  return <div
    target="_blank"
    rel="noreferrer"
    onClick={() => {
      newConfig(envName, repoName)
    }}
    className="relative block w-full border-2 border-gray-300 border-dashed rounded-lg p-6 text-center hover:border-pink-400 cursor-pointer text-gray-500 hover:text-pink-500"
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className="mx-auto h-12 w-12"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
    <div className="mt-2 block text-sm font-bold">
      Add deployment configuration
    </div>
  </div>
}
