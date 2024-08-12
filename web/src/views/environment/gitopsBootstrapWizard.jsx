import { useState, useEffect } from 'react';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import KustomizationPerApp from './kustomizationPerApp';
import SeparateEnvironments from './separateEnvironments';

export default function GitopsBootstrapWizard(props) {
  const { environment, bootstrap } = props

  const [repoPerEnv, setRepoPerEnv] = useState(true)
  const [infraRepo, setInfraRepo] = useState(`gitops-${environment.name}-infra`)
  const [appsRepo, setAppsRepo] = useState(`gitops-${environment.name}-apps`)
  const [kustomizationPerApp, setKustomizationPerApp] = useState(true)

  useEffect(() => {
    if (repoPerEnv) {
      setInfraRepo(`gitops-${environment.name}-infra`)
      setAppsRepo(`gitops-${environment.name}-apps`)
    } else {
      setInfraRepo(`gitops-infra`)
      setAppsRepo(`gitops-apps`)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [repoPerEnv]);

  return (
    <div className="w-full card">
      <div className="p-6 pb-4 items-center">
        <label htmlFor="environment" className="block font-medium">Bootstrap Gitops Repositories</label>
        <div className="my-4">
          <p className="max-w-4xl text-sm text-neutral-800 dark:text-neutral-400">
            To initialize this environment, we will create git repositories to store infrastructure and application deployments.
            Bootstrap the gitops repository to get started.
          </p>
          <KustomizationPerApp
            kustomizationPerApp={kustomizationPerApp}
            setKustomizationPerApp={setKustomizationPerApp}
          />
          <SeparateEnvironments
            repoPerEnv={repoPerEnv}
            setRepoPerEnv={setRepoPerEnv}
            infraRepo={infraRepo}
            appsRepo={appsRepo}
            setInfraRepo={setInfraRepo}
            setAppsRepo={setAppsRepo}
            envName={environment.name}
          />
        </div>
      </div>
      <div className='learnMoreBox flex items-center'>
        <div className='flex-grow'>
          Just use the defaults to get started, or learn more about <a href="https://gimlet.io" className='learnMoreLink'>Gitops Repositories <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
        </div>
        <button
          type="button"
          onClick={() => bootstrap(environment.name, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo)}
          className="primaryButton px-4"
        >Bootstrap Gitops Repositories</button>
      </div>
    </div>
  )
}
