import React, {Component} from 'react';
import ServiceCard from "../../components/serviceCard/serviceCard";
import {PlusIcon} from '@heroicons/react/solid'

export default class Services extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      envs: reduxState.envs,
      search: reduxState.search,
      agents: reduxState.settings.agents
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({envs: reduxState.envs});
      this.setState({search: reduxState.search});
      this.setState({agents: reduxState.settings.agents});
    });

    this.navigateToRepo = this.navigateToRepo.bind(this);
  }

  navigateToRepo(repo) {
    this.props.history.push(`/repo/${repo}`)
  }

  render() {
    let {envs, search, agents} = this.state;

    let filteredEnvs = {};
    for (const envName of Object.keys(envs)) {
      const env = envs[envName];
      filteredEnvs[env.name] = {name: env.name, stacks: env.stacks};
      if (search.filter !== '') {
        filteredEnvs[env.name].stacks = env.stacks.filter((service) => {
          return service.service.name.includes(search.filter) ||
            (service.deployment !== undefined && service.deployment.name.includes(search.filter)) ||
            (service.ingresses !== undefined && service.ingresses.filter((ingress) => ingress.url.includes(search.filter)).length > 0)
        })
      }
    }

    const emptyState = search.filter !== '' ?
      emptyStateNoMatchingService()
      :
      emptyStateNoServices();

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <h1 className="text-3xl font-bold leading-tight text-gray-900">Services</h1>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 py-8 sm:px-0">
              <div>
                {agents.length === 0 && emptyStateNoAgents()}
                {Object.keys(filteredEnvs).map((envName) => {
                  const env = filteredEnvs[envName];
                  const renderedServices = env.stacks.map((service) => {
                    return (
                      <li key={service.name} className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200">
                        <ServiceCard
                          service={service}
                          navigateToRepo={this.navigateToRepo}
                        />
                      </li>
                    )
                  })

                  return (
                    <div>
                      <h4 className="text-xl font-medium capitalize leading-tight text-gray-900 my-4">{envName}</h4>

                      <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                        {renderedServices.length > 0 ? renderedServices : emptyState}
                      </ul>
                    </div>
                  )
                })
                }
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

function emptyStateNoServices() {
  return (
    <p className="text-xs text-gray-800">No services</p>
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
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M5.636 18.364a9 9 0 010-12.728m12.728 0a9 9 0 010 12.728m-9.9-2.829a5 5 0 010-7.07m7.072 0a5 5 0 010 7.07M13 12a1 1 0 11-2 0 1 1 0 012 0z"/>
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
          <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true"/>
          Connect Gimlet Agent
        </a>
      </div>
    </div>
  )
}
