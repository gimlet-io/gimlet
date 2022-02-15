import React, { Component } from 'react';
import { Pod } from "../serviceCard/serviceCard";
import { RolloutHistory } from "../rolloutHistory/rolloutHistory";
import Emoji from "react-emoji-render";

function ServiceDetail(props) {
  const { stack, rolloutHistory, rollback, envName, navigateToConfigEdit, configExists } = props;

  return (
    <div class="w-full flex items-center justify-between space-x-6">
      <div class="flex-1 truncate">
        <h3 class="flex text-lg font-bold">
          {stack.service.name}
          {configExists &&
            <span onClick={() => navigateToConfigEdit(envName, stack.service.name)}>
              <svg
                className="cursor-pointer inline text-gray-500 hover:text-gray-700 ml-1  h-5 w-5"
                xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </span>
          }
        </h3>
        <div class="my-2 mb-4 sm:my-4 sm:mb-6">
          <RolloutHistory
            env={stack.env}
            app={stack.service.name}
            rollback={rollback}
            appRolloutHistory={rolloutHistory}
          />
        </div>
        <div class="flex flex-wrap text-sm">
          <div class="flex-1 min-w-full md:min-w-0">
            {stack.ingresses ? stack.ingresses.map((ingress) => <Ingress ingress={ingress} />) : null}
          </div>
          <div class="flex-1 md:ml-2 min-w-full md:min-w-0">
            <Deployment
              envName={stack.env}
              repo={stack.repo}
              deployment={stack.deployment}
            />
          </div>
          <div class="flex-1 min-w-full md:min-w-0" />
        </div>
      </div>
    </div>
  )
}

class Ingress extends Component {
  render() {
    const { ingress } = this.props;

    if (ingress === undefined) {
      return null;
    }

    return (
      <div class="bg-gray-100 p-2 mb-1 border rounded-sm border-gray-200 text-gray-500 relative">
        <span class="text-xs text-gray-400 absolute bottom-0 right-0 p-2">ingress</span>
        <div class="mb-1"><a href={'https://' + ingress.url} target="_blank" rel="noopener noreferrer">{ingress.url}</a>
        </div>
        <p class="text-xs">{ingress.namespace}/{ingress.name}</p>
      </div>
    );
  }
}

class Deployment extends Component {
  render() {
    const { deployment, repo } = this.props;

    if (deployment === undefined) {
      return null;
    }

    return (
      <div class="bg-gray-100 p-2 mb-1 border rounded-sm border-blue-200, text-gray-500 relative">
        <span class="text-xs text-gray-400 absolute bottom-0 right-0 p-2">deployment</span>
        <p class="mb-1">
          <p class="truncate">{deployment.commitMessage && <Emoji text={deployment.commitMessage} />}</p>
          <p class="text-xs italic"><a href={`https://github.com/${repo}/commit/${deployment.sha}`} target="_blank"
            rel="noopener noreferrer">{deployment.sha.slice(0, 6)}</a></p>
        </p>
        <p class="text-xs">{deployment.namespace}/{deployment.name}</p>
        {
          deployment.pods && deployment.pods.map((pod) => (
            <Pod key={pod.name} pod={pod} />
          ))
        }
      </div>
    );
  }

}

export default ServiceDetail;
