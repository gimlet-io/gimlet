import React, { Component } from 'react';
import { PlusIcon } from '@heroicons/react/solid';
import Releases from './releases';

export default class Pulse extends Component {
  constructor(props) {
    super(props);

    let reduxState = this.props.store.getState();
    this.state = {
      envs: reduxState.envs,
      releaseStatuses: reduxState.releaseStatuses,
      releaseHistorySinceDays: reduxState.settings.releaseHistorySinceDays
    }

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ envs: reduxState.envs });
      this.setState({ releaseStatuses: reduxState.releaseStatuses });
      this.setState({ releaseHistorySinceDays: reduxState.settings.releaseHistorySinceDays });
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
              <div>
                {this.state.envs.length > 0 ?
                  <div className="flow-root">
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


export function emptyStateNoAgents() {
  return (
    <div className="text-center mt-8 mb-16">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="mx-auto h-12 w-12 text-gray-400"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M5.636 18.364a9 9 0 010-12.728m12.728 0a9 9 0 010 12.728m-9.9-2.829a5 5 0 010-7.07m7.072 0a5 5 0 010 7.07M13 12a1 1 0 11-2 0 1 1 0 012 0z" />
      </svg>
      <h3 className="mt-2 text-sm font-medium text-gray-900">No connected Gimlet Agents</h3>
      <p className="mt-1 text-sm text-gray-500">Get started by installing the Gimlet Agent on a Kubernetes
        environment.</p>
      <div className="mt-6">
        <a
          href="https://gimlet.io"
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          Connect Gimlet Agent
        </a>
      </div>
    </div>
  )
}
