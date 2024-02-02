import jp from 'jsonpath'
import { Logs } from './logs'
import { Describe } from './describe'

export function CompactService(props) {
  const { service, capacitorClient, store } = props;
  const deployment = service.deployment;

  return (
    <div className="w-full flex items-center justify-between space-x-6 bg-white pb-6 rounded-lg border border-neutral-300 shadow-lg">
      <div className="flex-1">
        <h3 className="flex text-lg font-bold rounded p-4">
          <span className="cursor-pointer">{deployment.metadata.name}</span>
          <div className="flex items-center ml-auto space-x-2">
            {deployment &&
              <>
                <Logs
                  capacitorClient={capacitorClient}
                  store={store}
                  deployment={deployment}
                  containers={podContainers(service.pods)}
                />
                <Describe
                  capacitorClient={capacitorClient}
                  deployment={deployment}
                />
              </>
            }
          </div>
        </h3>
        <div>
          <div className="grid grid-cols-12 mt-4 px-4">
            <div className="col-span-5 space-y-4">
              <div>
                <p className="text-base text-neutral-600">Pods</p>
                {
                  service.pods.map((pod) => (
                    <Pod key={pod.metadata.name} pod={pod} />
                  ))
                }
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function Pod(props) {
  const {pod} = props;

  let color;
  let pulsar;
  switch (pod.status.phase) {
    case 'Running':
      color = 'bg-green-200';
      pulsar = '';
      break;
    case 'PodInitializing':
    case 'ContainerCreating':
    case 'Pending':
      color = 'bg-blue-300';
      pulsar = 'animate-pulse';
      break;
    case 'Terminating':
      color = 'bg-neutral-500';
      pulsar = 'animate-pulse';
      break;
    default:
      color = 'bg-red-600';
      pulsar = '';
      break;
  }

  return (
    <span className={`inline-block mr-1 mt-2 shadow-lg ${color} ${pulsar} font-bold px-2 cursor-default`} title={`${pod.metadata.name} - ${pod.status.phase}`}>
      {pod.status.phase}
    </span>
  );
}

function podContainers(pods) {
  const containers = [];

  pods.forEach((pod) => {
    const podName = jp.query(pod, '$.metadata.name')[0];

    const initContainerNames = jp.query(pod, '$.spec.initContainers[*].name');
    initContainerNames.forEach((initContainerName) => {
      containers.push(`${podName}/${initContainerName}`);
    });

    const containerNames = jp.query(pod, '$.spec.containers[*].name');
    containerNames.forEach((containerName) => {
      containers.push(`${podName}/${containerName}`);
    });
  });

  return containers;
}
