import React, { Component } from 'react';
import { format, formatDistance } from "date-fns";
import Releases from './releases';
import { InformationCircleIcon } from '@heroicons/react/solid'
import { Remarkable } from "remarkable";

export default class Pulse extends Component {
  constructor(props) {
    super(props);

    let reduxState = this.props.store.getState();
    this.state = {
      envs: reduxState.envs,
      releaseStatuses: reduxState.releaseStatuses,
      releaseHistorySinceDays: reduxState.settings.releaseHistorySinceDays,
      alerts: decorateKubernetesAlertsWithEnvAndRepo(reduxState.alerts, reduxState.connectedAgents),
      scmUrl: reduxState.settings.scmUrl,
      chartUpdatePullRequests: reduxState.pullRequests.chartUpdates,
    }

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ envs: reduxState.envs });
      this.setState({ releaseStatuses: reduxState.releaseStatuses });
      this.setState({ releaseHistorySinceDays: reduxState.settings.releaseHistorySinceDays });
      this.setState({ alerts: decorateKubernetesAlertsWithEnvAndRepo(reduxState.alerts, reduxState.connectedAgents) });
      this.setState({ scmUrl: reduxState.settings.scmUrl });
      this.setState({ chartUpdatePullRequests: reduxState.pullRequests.chartUpdates });
    });
  }

  render() {
    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <h1 className="text-3xl font-bold leading-tight text-gray-900">Pulse</h1>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 py-8 sm:px-0">
              {renderChartUpdatePullRequests(this.state.chartUpdatePullRequests)}
              {<AlertPanel
                alerts={this.state.alerts}
                history={this.props.history}
              />}
              <h3 className="text-2xl font-semibold leading-tight text-gray-900 mt-8 mb-8">Environments</h3>
              <div className="my-8">
                {this.state.envs.length > 0 ?
                  <div className="flow-root space-y-8">
                    {this.state.envs.map((env, idx) =>
                      <div key={idx}>
                        <Releases
                          gimletClient={this.props.gimletClient}
                          store={this.props.store}
                          env={env.name}
                          releaseHistorySinceDays={this.state.releaseHistorySinceDays}
                          releaseStatuses={this.state.releaseStatuses[env.name]}
                          scmUrl={this.state.scmUrl}
                          builtInEnv={env.builtIn}
                        />
                      </div>
                    )}
                  </div>
                  :
                  <p className="text-xs text-gray-800">You don't have any environments.</p>}
              </div>
            </div>
          </div>
        </main>
      </div>
    )
  }
}

export function emptyStateNoMatchingService() {
  return (
    <p className="text-base text-gray-800">No service matches the search</p>
  )
}

export function renderChartUpdatePullRequests(chartUpdatePullRequests) {
  if (JSON.stringify(chartUpdatePullRequests) === "{}") {
    return null
  }

  const prList = [];
  for (const [repoName, pullRequest] of Object.entries(chartUpdatePullRequests)) {
    prList.push(
      <li key={pullRequest.sha}>
        <a href={pullRequest.link} target="_blank" rel="noopener noreferrer">
          <span className="font-medium">{repoName}</span>: {pullRequest.title}
        </a>
      </li>)
  }

  return (
    <div className="rounded-md bg-blue-50 p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
        </div>
        <div className="ml-3 flex-1 text-blue-700 md:flex md:justify-between">
          <div className="text-xs flex flex-col">
            <span className="font-semibold text-sm">Helm chart version updates:</span>
            <ul className="list-disc list-inside text-xs ml-2">
              {prList}
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}

export function AlertPanel({ alerts, history, hideButton }) {
  if (!alerts) {
    return null;
  }

  if (alerts.length === 0) {
    return null;
  }

  const md = new Remarkable();

  return (
    <ul className="space-y-2 text-sm text-red-800 p-4">
      {alerts.map(alert => {
        return (
          <div key={`${alert.name} ${alert.objectName}`} className="flex bg-red-300 px-3 py-2 rounded relative">
            <div className="h-fit mb-8">
              <span className="text-sm">
                <p className="font-medium mb-2">
                  {alert.name} Alert {alert.status}
                </p>
                <div className="text-sm text-red-800">
                  <div className="prose-sm prose-headings:mb-1 prose-headings:mt-1 prose-p:mb-1 prose-code:bg-red-100 prose-code:p-1 prose-code:rounded text-red-900 w-full max-w-5xl" dangerouslySetInnerHTML={{ __html: md.render(alert.text) }} />
                </div>
              </span>
            </div>
            {!hideButton &&
              <>
                {alert.envName && <div className="absolute top-0 right-0 p-2 space-x-2 mb-6">
                  <span className="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-red-200">
                    {alert.envName}
                  </span>
                </div>}
                {alert.repoName && <div className="absolute bottom-0 right-0 p-2 space-x-2">
                  <button className="inline-flex items-center px-3 py-0.5 rounded-md text-sm font-medium bg-blue-400 text-slate-50"
                    onClick={() => history.push(`/repo/${alert.repoName}/${alert.envName}/${parseDeploymentName(alert.deploymentName)}`)}
                  >
                    Jump there
                  </button>
                </div>}
              </>}
            {dateLabel(alert.firedAt)}
            {dateLabel(alert.firedAt)}
          </div>
        )
      })}
    </ul>
  )
}

export function decorateKubernetesAlertsWithEnvAndRepo(alerts, connectedAgents) {
  alerts.forEach(alert => {
    const deploymentNamespace = alert.deploymentName.split("/")[0]
    const deploymentName = alert.deploymentName.split("/")[1]
    for (const env in connectedAgents) {
      connectedAgents[env].stacks.forEach(stack => {
        if (deploymentNamespace === stack.deployment.namespace && deploymentName === stack.deployment.name) {
          alert.envName = stack.env;
          alert.repoName = stack.repo;
        }
      })
    }
  })

  return alerts;
}

function dateLabel(lastSeen) {
  if (!lastSeen) {
    return null
  }

  const exactDate = format(lastSeen * 1000, 'h:mm:ss a, MMMM do yyyy')
  const dateLabel = formatDistance(lastSeen * 1000, new Date());

  return (
    <div
      className="text-xs text-red-700 absolute bottom-0 left-0 p-3"
      title={exactDate}
      target="_blank"
      rel="noopener noreferrer">
      {dateLabel} ago
    </div>
  );
}

export const parseDeploymentName = deployment => {
  return deployment.split("/")[1];
};
