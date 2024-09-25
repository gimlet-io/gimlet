import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import GitopsBootstrapWizard from './gitopsBootstrapWizard';
import ConnectCluster from './bootstrapGuide';
import Toggle from '../../components/toggle/toggle';
import {
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_ENVS
} from "../../redux/redux";
import { useNavigate } from 'react-router-dom'

export default function General(props) {
  const { gimletClient, store } = props;
  const { environment, scmUrl, isOnline, userToken } = props;

  const navigate = useNavigate()

  const deleteEnv = (envName) => {
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Deleting..."
      }
    });

    gimletClient.deleteEnvFromDB(envName)
      .then(() => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Environment deleted"
          }
        });
        refreshEnvs();
        navigate("/environments");
      }, err => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
      });
  }

  const refreshEnvs = () => {
    gimletClient.getEnvs()
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_ENVS,
          payload: data
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  const bootstrapGitops = (envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo) => {
    // setDisabled(true);
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Bootstrapping..."
      }
    });

    gimletClient.bootstrapGitops(envName, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo)
      .then(() => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Gitops environment bootstrapped"
          }
        });
        refreshEnvs();
        // setDisabled(false);
      }, (err) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
        // setDisabled(false);
      })
  }

  if (!environment.infraRepo || !environment.appsRepo) {
    return (
      <div className="w-full space-y-8">
        <GitopsBootstrapWizard environment={environment} bootstrap={bootstrapGitops} />
        <DeleteEnvCard environment={environment} deleteEnv={deleteEnv} />
      </div>
    )
  }

  if (!isOnline && !environment.ephemeral) {
    return (
      <div className="w-full space-y-8">
        <ConnectCluster envName={environment.name} token={userToken} />
        { !environment.builtIn && 
          <GitopRepositories environment={environment} scmUrl={scmUrl} />
        }
        <DeleteEnvCard environment={environment} deleteEnv={deleteEnv} />
      </div>
    )
  }

  return (
    <div className="w-full space-y-8">
      { !environment.builtIn && 
        <GitopRepositories environment={environment} scmUrl={scmUrl} />
      }
      <DeleteEnvCard environment={environment} deleteEnv={deleteEnv} />
    </div>
  )
}

function DeleteEnvCard(props) {
  const {environment, deleteEnv } = props

  return (
    <div className="w-full redCard">
      <div className="p-6 pb-4 items-center">
        <label htmlFor="environment" className="block font-medium">Delete Environment</label>
        <p className="text-sm text-neutral-800 dark:text-neutral-400 mt-4">
          The environment will be permanently deleted.
          <br /><br />
          The gitops repositories will remain intact, so as your deployed applications.
        </p>
      </div>
      <div className='flex items-center w-full learnMoreRed'>
        <div className='flex-grow'>
        </div>
        <button
          type="button"
          className="destructiveButton"
          onClick={() => {
            // eslint-disable-next-line no-restricted-globals
            confirm(`Are you sure you want to delete the ${environment.name} environment?`) &&
              deleteEnv(environment.name)
          }}
        >Delete</button>
      </div>
    </div>
  )
}

function GitopRepositories(props) {
  const {environment, scmUrl} = props

  const gitopsRepositories = [
    { name: environment.infraRepo, href: `${scmUrl}/${environment.infraRepo}` },
    { name: environment.appsRepo, href: `${scmUrl}/${environment.appsRepo}` }
  ];

  return (
    <div className="w-full card">
      <div className="p-6 pb-4 items-center">
        <label htmlFor="environment" className="block font-medium">
          Gitops Repositories
        </label>
        <div className="text-xs mt-4 font-mono">
          {gitopsRepositories.map((gitopsRepo) =>
          (
            <div className="flex" key={gitopsRepo.href}>
              {!environment.builtIn &&
                <a className="externalLink mb-1" href={gitopsRepo.href} target="_blank" rel="noreferrer">{gitopsRepo.name}
                  <ArrowTopRightOnSquareIcon className="externalLinkIcon ml-1" aria-hidden="true" />
                </a>
              }
              {environment.builtIn &&
                <div className="mb-1" href={gitopsRepo.href} target="_blank" rel="noreferrer">{gitopsRepo.name}</div>
              }
            </div>
          ))}
        </div>
        <div className="space-y-2 mt-2">
          <p className="mr-1 font-medium">Kustomization per app</p>
          <Toggle
            checked={environment.kustomizationPerApp}
            disabled
          />
          <p className="mr-1 font-medium">Separate environments by git repositories</p>
          <Toggle
            checked={environment.repoPerEnv}
            disabled
          />
        </div>
      </div>
      <div className='learnMoreBox'>
        Learn more about <a href="https://gimlet-documentation-home-page-revamp-emxxuioo.gimlet.app/docs/environment-settings/introduction#gitops-repositories" className='learnMoreLink'>Gitops Repositories <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
      </div>
    </div>
  )
}
