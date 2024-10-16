import { useNavigate, useParams } from "react-router-dom";
import ServiceDetail from "../serviceDetail/serviceDetail";
import { InformationCircleIcon } from '@heroicons/react/20/solid';

export function Env(props) {
  const { store, gimletClient } = props
  const { env, repoRolloutHistory, envConfigs, rollback, fileInfos } = props;
  const { releaseHistorySinceDays, settings, alerts, appFilter } = props;

  return (
    <div>
      <h4 className="relative flex items-stretch select-none text-xl font-medium capitalize leading-tight text-neutral-900 dark:text-neutral-200 my-4">
        {env.name}
        <span title={env.isOnline ? "Connected" : "Disconnected"}>
          <svg className={(env.isOnline ? "text-green-400 dark:text-teal-600" : "text-red-400") + " inline fill-current ml-1"} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
            <path
              d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
            />
          </svg>
        </span>
      </h4>
      <div className="space-y-4">
        {!env.isOnline &&
          <ConnectEnvCard env={env} trial={settings.trial} />
        }
        <Services
          store={store}
          gimletClient={gimletClient}
          envName={env.name}
          stacks={env.stacks}
          ephemeralEnv={env.ephemeral}
          builtInEnv={env.builtIn}
          envConfigs={envConfigs}
          repoRolloutHistory={repoRolloutHistory}
          rollback={rollback}
          fileInfos={fileInfos}
          releaseHistorySinceDays={releaseHistorySinceDays}
          settings={settings}
          alerts={alerts}
          appFilter={appFilter}
        />
      </div>
    </div>
  )
}

function Services(props) {
  const { envName, ephemeralEnv, builtInEnv, stacks, envConfigs, repoRolloutHistory, rollback, fileInfos, releaseHistorySinceDays, gimletClient, store, settings, alerts, appFilter } = props
  const { owner, repo, deployment } = useParams()
  const repoName = `${owner}/${repo}`

  console.log(repoRolloutHistory)

  let configsWeHave = [];
  if (envConfigs) {
    configsWeHave = envConfigs
      .filter((config) => !config.preview)
      .map((config) => config.app);
  }

  const filteredStacks = stacks
    .filter(stack => stack.service.name.includes(appFilter)) // app filter
    .filter(stack => stack.deployment?.branch === "") // filter preview deploys from this view

  let configsWeDeployed = [];
  let services = [];

  // services that are deployed on k8s
  services = filteredStacks.map((stack) => {
    configsWeDeployed.push(stack.service.name);
    const config = envConfigs.find((config) => config.app === stack.service.name)
    return {stack, config}
  })

  const configsWeHaventDeployed = configsWeHave.filter(config => !configsWeDeployed.includes(config) && config.includes(appFilter));

  services.push(
    ...configsWeHaventDeployed.map(config => ({
      stack: {service: {name: config}},
      config: config
    }))
  )

  if (services.length >= 10) {
    return services.slice(0, 10);
  }

  return (
    <>
    {services.length === 10 &&
      <span className="text-xs text-blue-700">Displaying at most 10 application configurations per environment.</span>
    }
    {services.length === 0 && emptyStateDeployThisRepo(navigate, envName, owner, repoName) }
    {services.map(({stack, config}) => (
      <div key={'sc-'+stack.service.name} className="w-full flex items-center justify-between space-x-6 p-4 card">
        <ServiceDetail
          key={'sc-'+stack.service.name}
          store={store}
          gimletClient={gimletClient}
          stack={stack}
          rolloutHistory={repoRolloutHistory?.[envName]?.[stack.service.name]}
          rollback={rollback}
          envName={envName}
          ephemeralEnv={ephemeralEnv}
          fileName={fileName(fileInfos, stack.service.name)}
          config={config}
          releaseHistorySinceDays={releaseHistorySinceDays}
          scmUrl={settings.scmUrl}
          serviceAlerts={alerts[deployment]}
          builtInEnv={builtInEnv}
        />
      </div>
    ))}
    </>
  )
}

function fileName(fileInfos, appName) {
  if (fileInfos.find(fileInfo => fileInfo.appName === appName)) {
    return fileInfos.find(fileInfo => fileInfo.appName === appName).fileName;
  }
}

function ConnectEnvCard(props) {
  let { env, trial } = props
  const navigate = useNavigate()

  let expired = false
  let stuck = false
  let startingUp = false
  if (trial) {
    const expiringAt = new Date(env.expiry * 1000);
    expired = expiringAt < new Date()
    const sevenDays = 604800000
    const age = new Date() - (expiringAt-sevenDays)
    stuck = age > 5*60*1000
    startingUp = !expired && !stuck
  }

  const color = startingUp ? 'blue' : 'red'

  return (
    <div className={`rounded-md bg-${color}-100 p-4`}>
    <div className="flex">
      <div className="flex-shrink-0">
        <InformationCircleIcon className={`h-5 w-5 text-${color}-400`} aria-hidden="true" />
      </div>
      <div className="ml-3">
        <h3 className={`text-sm font-bold text-${color}-800`}>Environment disconnected</h3>
        <div className={`mt-2 text-sm text-${color}-800`}>
          {startingUp &&
          <>
            This environment is still initializing and should be up in a couple of minutes.
            <svg className="animate-spin h-3 w-3 text-black inline ml-1" xmlns="http://www.w3.org/2000/svg" fill="none"
              viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4}></circle>
              <path className="opacity-75" fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg><br />
            <button className='font-bold cursor-pointer'
              onClick={() => {navigate(`/env/${env.name}`);return true}}
            >
              Click to learn more about this environment
            </button>
          </>
          }
          {expired &&
          <>
            This environment is disconnected.
            Your trial expired, so the underlying Kubernetes cluster is stopped.<br />
            <button className='font-bold cursor-pointer'
              onClick={() => {navigate(`/env/${env.name}`);return true}}
            >
              Click to purchase a license.
            </button>
          </>
          }
          {!expired && stuck &&
          <>
            This environment is disconnected as
            your trial Kubernetes cluster is stuck - please contact support.
          </>
          }
          {!trial &&
          <>
            This environment is disconnected.<br />
            <button className='font-bold cursor-pointer'
              onClick={() => {navigate(`/env/${env.name}`);return true}}
            >
              Click to connect this environment to a cluster on the Environments page.
            </button>
          </>
          }
        </div>
      </div>
    </div>
    </div>
  );
}

function emptyStateDeployThisRepo(navigate, envName, owner, repo) {
  return (
    <div className='card w-full p-4 mt-4'>
      <div className='items-center border-dashed border border-neutral-200 dark:border-neutral-700 rounded-md p-4 py-16'>
        <h3 className="mt-2 text-sm font-semibold text-center">No Deployments</h3>
        <p className="mt-1 text-sm text-neutral-500 dark:text-neutral-400 text-center">Get started by configuring a new deployment.</p>
        <div className="mt-6 text-center">
          <button
            onClick={() => navigate(encodeURI(`/repo/${owner}/${repo}/envs/${envName}/deploy`))}
            className="primaryButton px-8 capitalize">
            New Deployment to {envName}
          </button>
        </div>
      </div>
    </div>
  )
}
