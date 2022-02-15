import React from 'react';
import './serviceCard.css';
import * as PropTypes from "prop-types";
import Emoji from "react-emoji-render";

function ServiceCard(props) {
  const {service, navigateToRepo} = props;

  return (
    <div className="w-full flex items-center justify-between p-6 space-x-6 cursor-pointer"
         onClick={() => navigateToRepo(service.repo)}>
      <div className="flex-1 truncate">
        <p className="text-sm font-bold">{service.service.namespace}/{service.service.name}
          <span
            className="flex-shrink-0 inline-block px-2 py-0.5 mx-1 text-green-800 text-xs font-medium bg-green-100 rounded-full">
            {service.env}
          </span>
        </p>
        <div className="flex items-center space-x-3">
          <h3 className="text-gray-500 mb-2 text-xs font-medium truncate">{service.repo}</h3>
        </div>
        <Deployment
          envName={service.env}
          deployment={service.deployment}
          repo={service.repo}
        />
      </div>
    </div>
  )
}

export function Deployment(props) {
  const {deployment, repo} = props;

  if (!deployment) {
    return null;
  }

  return (
    <div>
      <p className="mb-1">
        <p className="truncate">{deployment.message}</p>
        <p className="truncate text-xs"><a href={`https://github.com/${repo}/commit/${deployment.sha}`} target="_blank"
                                    rel="noopener noreferrer" onClick={(e) => {
                                      e.stopPropagation();
                                      return true
                                    }}>{deployment.commitMessage && <Emoji text={deployment.commitMessage}/>}</a></p>
        <p className="text-xs italic">
        <a href={`https://github.com/${repo}/commit/${deployment.sha}`} target="_blank"
                                       rel="noopener noreferrer" onClick={(e) => {
                                        e.stopPropagation();
                                        return true
                                      }}>
            {deployment.sha.slice(0, 6)}
          </a>
        </p>
      </p>
      {deployment.pods.map((pod) => (
        <Pod key={pod.name} pod={pod}/>
      ))
      }
    </div>
  );
}

Deployment.propTypes =
  {
    deployment: PropTypes.any,
  }
;

export function Pod(props) {
  const {pod} = props;

  let color;
  let pulsar;
  switch (pod.status) {
    case 'Running':
      color = 'text-blue-200';
      pulsar = '';
      break;
    case 'PodInitializing':
    case 'ContainerCreating':
    case 'Pending':
      color = 'text-blue-100';
      pulsar = 'pulsar-green';
      break;
    case 'Terminating':
      color = 'text-blue-800';
      pulsar = 'pulsar-gray';
      break;
    default:
      color = 'text-red-600';
      pulsar = '';
      break;
  }

  return (
    <span className="inline-block w-4 mr-1 mt-2">
      <svg viewBox="0 0 1 1"
           className={`fill-current ${color} ${pulsar}`}>
        <g>
          <title>{pod.name} - {pod.status}</title>
          <rect width="1" height="1"/>
        </g>
      </svg>
    </span>
  );
}

export default ServiceCard;
