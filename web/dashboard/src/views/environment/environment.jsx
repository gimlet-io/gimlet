import React, { Component } from 'react';
import {
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWERRORLIST,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_ENVUPDATED,
  ACTION_TYPE_SAVE_ENV_PULLREQUEST,
  ACTION_TYPE_ENV_PULLREQUESTS,
  ACTION_TYPE_ENVSPINNEDOUT,
  ACTION_TYPE_ENVS
} from "../../redux/redux";
import { InformationCircleIcon } from '@heroicons/react/solid'
import SeparateEnvironments from './separateEnvironments';
import KustomizationPerApp from './kustomizationPerApp';
import BootstrapGuide from './bootstrapGuide';
import StackUI from './stack-ui';
import DeleteButton from './deleteButton';

export default class EnvironmentView extends Component {
  constructor(props) {
    super(props);
    let reduxState = this.props.store.getState();
    const { env } = this.props.match.params;

    this.state = {
      connectedAgents: reduxState.connectedAgents,
      environment: findEnv(reduxState.envs, env),
      user: reduxState.user,
      disabled: false,
      gitopsUpdatePullRequests: reduxState.pullRequests.gitopsUpdates[env],
      scmUrl: reduxState.settings.scmUrl,
      settings: reduxState.settings,
      errors: {},
      kustomizationPerApp: false,
      repoPerEnv: true,
      infraRepo: "gitops-infra",
      appsRepo: "gitops-apps",
    };
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        connectedAgents: reduxState.connectedAgents,
        environment: findEnv(reduxState.envs, env),
        user: reduxState.user,
        popupWindow: reduxState.popupWindow,
        gitopsUpdatePullRequests: reduxState.pullRequests.gitopsUpdates[env],
        scmUrl: reduxState.settings.scmUrl,
        settings: reduxState.settings
      });

      if (!this.state.stack) {
        this.setLocalStackConfig(this.state.environment)
      }
    });

    this.setValues = this.setValues.bind(this)
    this.setRepoPerEnv = this.setRepoPerEnv.bind(this)
    this.setKustomizationPerApp = this.setKustomizationPerApp.bind(this)
    this.setInfraRepo = this.setInfraRepo.bind(this)
    this.setAppsRepo = this.setAppsRepo.bind(this)
    this.delete = this.delete.bind(this)
    this.validationCallback = this.validationCallback.bind(this)
  }

  componentDidMount() {
    const { gimletClient, store } = this.props;

    gimletClient.getPullRequestsFromInfraRepo()
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_ENV_PULLREQUESTS,
          payload: data
        });
      }, () => {/* Generic error handler deals with it */
      })
  }

  setLocalStackConfig(env) {
    if (env && env.stackConfig) {
      this.setState({
        stack: Object.assign({}, env.stackConfig.config),
        stackNonDefaultValues: Object.assign({}, env.stackConfig.config),
      })
    }
  }

  isOnline(onlineEnvs, singleEnv) {
    return Object.keys(onlineEnvs)
      .map(env => onlineEnvs[env])
      .some(onlineEnv => {
        return onlineEnv.name === singleEnv.name
      })
  };

  configurationTab() {
    const { environment, scmUrl } = this.state;

    if (!environment.infraRepo || !environment.appsRepo) {
      return this.gitopsBootstrapWizard();
    }

    const isRepoPerEnvEnabled = environment.repoPerEnv ? "enabled" : "disabled";
    const isKustomizationPerAppEnabled = environment.kustomizationPerApp ? "enabled" : "disabled";
    const gitopsRepositories = [
      { name: environment.infraRepo, href: `${scmUrl}/${environment.infraRepo}` },
      { name: environment.appsRepo, href: `${scmUrl}/${environment.appsRepo}` }
    ];

    return (
      <div className="mt-4 text-sm text-gray-500 my-4 bg-white overflow-hidden shadow rounded-lg divide-gray-200 px-4 py-5 sm:px-6">
        <div className="space-y-1">
          <span className="flex"><p className="mr-1 font-semibold text-gray-600">Kustomization per app setting</p> is {isKustomizationPerAppEnabled} for this environment.</span>
          <span className="flex"><p className="mr-1 font-semibold text-gray-600">Separate environments by git repositories setting</p> is {isRepoPerEnvEnabled} for this environment.</span>
        </div>
        <div className="mt-4 mb-1">
          <span className="font-bold text-gray-600">Gitops repositories</span>
          <div className="ml-4 mt-1 font-mono">
            {gitopsRepositories.map((gitopsRepo) =>
            (
              <div className="flex" key={gitopsRepo.href}>
                {!environment.builtIn &&
                  <a className="mb-1 hover:text-gray-600" href={gitopsRepo.href} target="_blank" rel="noreferrer">{gitopsRepo.name}
                    <svg xmlns="http://www.w3.org/2000/svg"
                      className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
                      viewBox="0 0 24 24">
                      <path d="M0 0h24v24H0z" fill="none" />
                      <path
                        d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                    </svg>
                  </a>
                }
                {environment.builtIn &&
                  <div className="mb-1" href={gitopsRepo.href} target="_blank" rel="noreferrer">{gitopsRepo.name}</div>
                }
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  saveComponents() {
    const { gimletClient, store } = this.props;
    const { errors, environment, stackNonDefaultValues } = this.state;

    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Saving components..."
      }
    });

    for (const variable of Object.keys(errors)) {
      if (errors[variable] !== null) {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERRORLIST, payload: {
            header: "Error",
            errorList: errors
          }
        });
        return false
      }
    }

    gimletClient.saveInfrastructureComponents(environment.name, stackNonDefaultValues)
      .then((data) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Pull request was created",
            link: data.createdPr.link
          }
        });
        store.dispatch({
          type: ACTION_TYPE_SAVE_ENV_PULLREQUEST, payload: {
            envName: data.envName,
            createdPr: data.createdPr
          }
        });
        store.dispatch({
          type: ACTION_TYPE_ENVUPDATED, name: environment.name, payload: data.stackConfig
        });
      }, (err) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
        this.resetPopupWindowAfterThreeSeconds()
      })
  }

  refreshEnvs() {
    this.props.gimletClient.getEnvs()
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_ENVS,
          payload: data
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  delete(envName) {
    this.props.store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Deleting..."
      }
    });

    this.props.gimletClient.deleteEnvFromDB(envName)
      .then(() => {
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Environment deleted"
          }
        });
        this.refreshEnvs();
        this.props.history.push("/environments");
      }, err => {
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
      });
  }

  spinOutBuiltInEnv() {
    const { gimletClient, store } = this.props;

    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Converting environment..."
      }
    });
    gimletClient.spinOutBuiltInEnv()
      .then((data) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Environment was converted",
          }
        });
        store.dispatch({
          type: ACTION_TYPE_ENVSPINNEDOUT,
          payload: data
        });
      }, (err) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
        this.resetPopupWindowAfterThreeSeconds()
      })
  }

  resetPopupWindowAfterThreeSeconds() {
    const { store } = this.props;
    setTimeout(() => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  };

  builtInEnvInfo() {
    return (
      <div className="rounded-md bg-blue-50 p-4 my-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-blue-800">This is a built-in environment</h3>
            <div className="mt-2 text-sm text-blue-700">
              Gimlet made this environment for you so you can quickly start deploying.<br />
              To make edits to it,
              <span className="cursor-pointer underline font-medium pl-1 text-blue-800"
                onClick={() => {
                  // eslint-disable-next-line no-restricted-globals
                  confirm(`Are you sure you want to convert to a gitops environment?`) &&
                    this.spinOutBuiltInEnv()
                }}
              >
                convert it to a gitops environment
              </span> first.
              <br />
              By doing so, Gimlet will create two git repositories to host the infrastructure and application manifests.
            </div>
          </div>
        </div>
      </div>
    );
  }

  setValues = (variable, values, nonDefaultValues) => {
    this.setState((prevState) => ({
      stack: { ...prevState.stack, [variable]: values },
      stackNonDefaultValues: { ...prevState.stackNonDefaultValues, [variable]: nonDefaultValues },
    }));
  }

  validationCallback(variable, validationErrors) {
    if (validationErrors !== null) {
      validationErrors = validationErrors.filter(error => error.keyword !== 'oneOf');
      validationErrors = validationErrors.filter(error => error.dataPath !== '.enabled');
    }

    this.setState((prevState) => ({
      errors: { ...prevState.errors, [variable]: validationErrors },
    }));
  }

  infrastructureComponentsTab() {
    const { environment, settings, stack } = this.state;

    if (!settings.provider || settings.provider === "") {
      return (
        <div className="rounded-md bg-blue-50 p-4 mb-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">Integrate Github</h3>
              <div className="mt-2 text-sm text-blue-700">
                This view will load your git repositories once you integrated Github.<br />
                <button
                  className="font-medium"
                  onClick={() => { this.props.history.push("/settings"); return true }}
                >
                  Click to integrate Github on the Settings page.
                </button>
              </div>
            </div>
          </div>
        </div>
      )
    }

    if (environment.builtIn) {
      return this.builtInEnvInfo();
    }

    return (
      <div className="relative mt-4 text-gray-700">
        <div className='absolute right-0 top-0 pointer-events-none z-10 py-1'>
          <button
            onClick={() => this.saveComponents()}
            disabled={environment.builtIn}
            className={(environment.builtIn ? 'bg-gray-600 cursor-default' : 'bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700') + ` inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150 ml-auto tracking-wide pointer-events-auto`}>
            Save components
          </button>
        </div>
        <StackUI
          stack={stack}
          stackDefinition={environment.stackDefinition}
          setValues={this.setValues}
          validationCallback={this.validationCallback}
        />
      </div>
    )
  }

  setKustomizationPerApp(value) {
    this.setState({ kustomizationPerApp: value });
  }

  setRepoPerEnv(value) {
    this.setState({ repoPerEnv: value });
  }

  setInfraRepo(value) {
    this.setState({ infraRepo: value });
  }

  setAppsRepo(value) {
    this.setState({ appsRepo: value });
  }

  bootstrapGitops() {
    const { environment, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo } = this.state;
    const { gimletClient, store } = this.props;

    if (this.state.disabled) {
      return
    }
    this.setState({ disabled: true });
    store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Bootstrapping..."
      }
    });

    gimletClient.bootstrapGitops(environment.name, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo)
      .then(() => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Gitops environment bootstrapped"
          }
        });
        this.refreshEnvs();
        this.setState({ disabled: false });
        this.resetPopupWindowAfterThreeSeconds()
      }, (err) => {
        store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
        this.setState({ disabled: false });
        this.resetPopupWindowAfterThreeSeconds()
      })
  }

  gitopsBootstrapWizard() {
    const { environment, disabled, repoPerEnv, kustomizationPerApp, infraRepo, appsRepo } = this.state;
    if (repoPerEnv && infraRepo === "gitops-infra") {
      this.setInfraRepo(`gitops-${environment.name}-infra`);
    }
    if (repoPerEnv && appsRepo === "gitops-apps") {
      this.setAppsRepo(`gitops-${environment.name}-apps`);
    }
    if (!repoPerEnv && infraRepo === `gitops-${environment.name}-infra`) {
      this.setInfraRepo("gitops-infra");
    }
    if (!repoPerEnv && appsRepo === `gitops-${environment.name}-apps`) {
      this.setAppsRepo("gitops-apps");
    }

    return (
      <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-gray-200 px-4 py-5 sm:px-6">
        <div className="mt-2 pb-4 border-b border-gray-200">
          <h3 className="text-lg leading-6 font-medium text-gray-900">Bootstrap gitops repository</h3>
          <p className="mt-2 max-w-4xl text-sm text-gray-500">
            To initialize this environment, bootstrap the gitops repository first
          </p>
        </div>
        <KustomizationPerApp
          kustomizationPerApp={kustomizationPerApp}
          setKustomizationPerApp={this.setKustomizationPerApp}
        />
        <SeparateEnvironments
          repoPerEnv={repoPerEnv}
          setRepoPerEnv={this.setRepoPerEnv}
          infraRepo={infraRepo}
          appsRepo={appsRepo}
          setInfraRepo={this.setInfraRepo}
          setAppsRepo={this.setAppsRepo}
          envName={environment.name}
        />
        <div className="p-0 flow-root mt-8">
          <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
            <button
              onClick={() => this.bootstrapGitops(environment.name, repoPerEnv, kustomizationPerApp)}
              disabled={disabled}
              className="bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700 inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150"
            >
              Bootstrap gitops repository
            </button>
          </span>
        </div>
      </div>
    )
  }

  render() {
    let { environment, connectedAgents, user, gitopsUpdatePullRequests } = this.state;
    const isOnline = this.isOnline(connectedAgents, environment)

    if (!environment) {
      return null
    }

    const navigation = [
      { name: 'Details', href: `/env/${environment.name}` },
      { name: 'Components', href: `/env/${environment.name}/components` },
    ]

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-12">
            <div className="sm:flex items-start">
              <div className='relative flex-1'>
                <h1 className="flex text-3xl font-bold capitalize leading-tight text-gray-900">
                  {environment.name}
                  <span title={isOnline ? "Connected" : "Disconnected"}>
                    <svg className={(isOnline ? "text-green-400" : "text-red-400") + " inline fill-current ml-2"} xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 20 20">
                      <path
                        d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
                      />
                    </svg>
                  </span>
                  {environment.infraRepo === "" &&
                    <span title="Gitops automation is not bootstrapped">
                      <svg xmlns="http://www.w3.org/2000/svg" className="inline ml-2 h-6 w-6 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                      </svg>
                    </span>}
                </h1>
                {!isOnline &&
                  <div className="absolute top-0 right-0">
                    <DeleteButton
                      envName={environment.name}
                      deleteFunc={this.delete}
                    />
                  </div>
                }
                <button className="text-gray-500 hover:text-gray-700" onClick={() => this.props.history.push("/environments")}>
                  &laquo; back
                </button>
              </div>
            </div>
            <div className="space-y-4">
              <PullRequests
                title="Gitops manifest updates"
                items={gitopsUpdatePullRequests}
              />
              <PullRequests
                title="Gimlet stack"
                items={environment.pullRequests}
                />
            </div>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 py-8 sm:px-0">
              {environment.infraRepo ?
                <>
                  {isOnline ?
                    <div>
                      <div className="block">
                        <div className="border-b border-gray-200">
                          <nav className="-mb-px flex" aria-label="Tabs">
                            {navigation.map((item) => {
                              const selected = this.props.location.pathname === item.href;
                              return (
                                <button
                                  key={item.name}
                                  href="#"
                                  className={(
                                    selected
                                      ? 'border-indigo-500 text-indigo-600'
                                      : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700') +
                                    ' w-1/2 border-b-2 py-4 px-1 text-center text-sm font-medium'
                                  }
                                  aria-current={selected ? 'page' : undefined}
                                  onClick={() => {
                                    this.props.history.push(item.href);
                                    return true
                                  }}
                                >
                                  {item.name}
                                </button>
                              )
                            })
                            }
                          </nav>
                        </div>
                      </div>
                      <div className="my-8">
                        {
                          navigation[0].href === this.props.location.pathname &&
                          this.configurationTab()
                        }
                        {
                          navigation[1].href === this.props.location.pathname &&
                          this.infrastructureComponentsTab()
                        }
                      </div>
                    </div>
                    :
                    <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-gray-200 px-4 py-5 sm:px-6">
                      <h3 className="text-lg font-medium pl-2 text-gray-900">Connect your cluster</h3>
                      <BootstrapGuide
                        envName={environment.name}
                        token={user.token}
                      />
                    </div>}
                </>
                :
                this.gitopsBootstrapWizard()
              }

            </div>
          </div>
        </main>
      </div>
    )
  }
}

const PullRequests = ({title, items}) => {
  if (!items) {
    return null
  }

  if (items.length === 0) {
    return null
  }

  const prList = [];
  items.forEach(p => {
    prList.push(
      <li key={p.sha}>
        <a href={p.link} target="_blank" rel="noopener noreferrer">
          <span className="font-medium">{`#${p.number}`}</span>: {p.title}
        </a>
      </li>)
  })

  return (
    <div className="rounded-md bg-blue-50 p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
        </div>
        <div className="ml-3 flex-1 text-blue-700 md:flex md:justify-between">
          <div className="text-xs flex flex-col">
            <span className="font-semibold text-sm">{title}:</span>
            <ul className="list-disc list-inside text-xs ml-2">
              {prList}
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}

const findEnv = (envs, envName) => {
  if (envs.length === 0) {
    return null
  }

  return envs.find(env => env.name === envName)
};
