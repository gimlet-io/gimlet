import React, { Component } from 'react';
import { format, formatDistance } from "date-fns";
import { nanoid } from 'nanoid';
import Releases from './releases';

export default class Pulse extends Component {
  constructor(props) {
    super(props);

    let reduxState = this.props.store.getState();
    this.state = {
      envs: reduxState.envs,
      releaseStatuses: reduxState.releaseStatuses,
      releaseHistorySinceDays: reduxState.settings.releaseHistorySinceDays,
      kubernetesEvents: decorateKubernetesEventsWithEnvAndRepo(reduxState.kubernetesEvents, reduxState.connectedAgents)
    }

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ envs: reduxState.envs });
      this.setState({ releaseStatuses: reduxState.releaseStatuses });
      this.setState({ releaseHistorySinceDays: reduxState.settings.releaseHistorySinceDays });
      this.setState({ kubernetesEvents: decorateKubernetesEventsWithEnvAndRepo(reduxState.kubernetesEvents, reduxState.connectedAgents) });
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
              {<KubernetesEventsAlertBox
                kubernetesEvents={this.state.kubernetesEvents}
                history={this.props.history}
              />}
              <h3 className="text-2xl font-semibold leading-tight text-gray-900 mt-16 mb-8">Environments</h3>
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

export function KubernetesEventsAlertBox({ kubernetesEvents, history, hideButton }) {
  if (kubernetesEvents.length === 0) {
    return null;
  }

  return (
    <ul className="space-y-2 text-sm text-red-800">
      {kubernetesEvents.map(event => {
        const exactDate = format(event.lastSeen * 1000, 'h:mm:ss a, MMMM do yyyy')
        const dateLabel = formatDistance(event.lastSeen * 1000, new Date());

        return (
          <div key={nanoid()} className="flex bg-red-300 px-3 py-2 rounded relative">
            <div className="h-fit mb-8">
              <span className="text-sm">
                <p className="font-medium lowercase mb-2">
                  {event.object} {event.reason}
                </p>
                <p>
                  {event.message}
                </p>
              </span>
            </div>
            {!hideButton &&
              <>
                {event.envName && <div className="absolute top-0 right-0 p-2 space-x-2 mb-6">
                  <span className="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-red-200">
                    {event.envName}
                  </span>
                </div>}
                {event.repoName && <div className="absolute bottom-0 right-0 p-2 space-x-2">
                  <button className="inline-flex items-center px-3 py-0.5 rounded-md text-sm font-medium bg-blue-400 text-slate-50"
                    onClick={() => history.push(`/repo/${event.repoName}/${event.envName}/${event.deploymentName}`)}
                  >
                    Jump there
                  </button>
                </div>}
              </>}
            <div
              className="text-xs text-red-700 absolute bottom-0 left-0 p-3"
              title={exactDate}
              target="_blank"
              rel="noopener noreferrer">
              {dateLabel} ago
            </div>
          </div>
        )
      })}
    </ul>
  )
}

export function decorateKubernetesEventsWithEnvAndRepo(kubernetesEvents, connectedAgents) {
  kubernetesEvents.forEach(event => {
    for (const env in connectedAgents) {
      connectedAgents[env].stacks.forEach(stack => {
        if (event.deploymentNamespace === stack.deployment.namespace && event.deploymentName === stack.deployment.name) {
          event.envName = stack.env;
          event.repoName = stack.repo;
        }
      })
    }
  })

  return kubernetesEvents;

}
