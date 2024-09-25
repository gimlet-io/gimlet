import { format, formatDistance } from "date-fns";
import { Remarkable } from "remarkable";
import { Menu } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/24/solid';
import { useNavigate } from "react-router-dom";

export function AlertPanel({ alerts, hideButton, silenceAlert }) {
  const navigate = useNavigate()
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
                  <button className="inline-flex items-center px-3 py-0.5 rounded-md text-sm font-medium bg-blue-400 text-neutral-50"
                    onClick={() => navigate(`/repo/${alert.repoName}/${alert.envName}/${parseDeploymentName(alert.deploymentName)}`)}
                  >
                    Jump there
                  </button>
                </div>}
              </>}
            <div className="absolute top-0 right-0 p-2 space-x-2 mb-6">
              <SilenceWidget
                alert={alert} 
                silenceAlert={silenceAlert}
              />
            </div>
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

const SilenceWidget = ({ alert, silenceAlert }) => {
  const currentUnix = (new Date().getTime() / 1000).toFixed(0)
  if (alert.silencedUntil && alert.silencedUntil > currentUnix) {
    return silenceUntilDateLabel(alert.silencedUntil)
  }

  const object =`${alert.deploymentName}-${alert.type}`
  const silenceOptions = [
    { title: 'for 2 hours', hours: 2 },
    { title: 'for 24 hours', hours: 24 },
    { title: 'for 1 week', hours: 24 * 7 },
  ]

  return (
    <Menu as="span" className="relative">
      <div className="inline-flex rounded-md shadow-sm">
        <Menu.Button className="inline-flex items-center rounded-l-md rounded-r-md bg-red-400 p-2 hover:bg-red-500 text-white space-x-1">
          <p className="text-sm font-semibold">Silence</p>
          <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
        </Menu.Button>
      </div>

      <Menu.Items className="absolute right-0 z-10 mt-2 w-72 origin-top-right divide-y divide-neutral-200 overflow-hidden rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
        {silenceOptions.map((option) => (
          <Menu.Item
            key={option.title}
            className={'bg-red-400 text-white select-none text-sm'}
            value={option}
          >
            {({ active }) => (
              <button
                onClick={() => {
                  // eslint-disable-next-line no-restricted-globals
                  confirm(`Are you sure you want to silence ${alert.deploymentName} ${alert.type} alerts ${option.title}?`) &&
                  silenceAlert(object, option.hours);
                }}
                className={(
                  active ? 'bg-red-500 text-neutral-100' : 'text-neutral-100') +
                  ' block px-4 py-2 text-sm w-full text-left'
                }
              >
                {option.title}
              </button>
            )}
          </Menu.Item>
        ))}
      </Menu.Items>
    </Menu>
  )
}

function silenceUntilDateLabel(silenceUntil) {
  const exactDate = format(silenceUntil * 1000, 'h:mm:ss a, MMMM do yyyy')

  return (
    <div
      className="text-xs text-red-700 p-3"
      target="_blank"
      rel="noopener noreferrer">
      Silenced until: {exactDate}
    </div>
  );
}
