import React, { Component } from "react";
import HelmUI from "helm-react-ui";
import "./style.css";
import PopUpWindow from "./popUpWindow";
import ReactDiffViewer from "react-diff-viewer";
import YAML from "json-to-pretty-yaml";
import CopiableCodeSnippet from "./copiableCodeSnippet";
import {
  ACTION_TYPE_CHARTSCHEMA,
  ACTION_TYPE_ENVCONFIGS,
  ACTION_TYPE_REPO_METAS,
  ACTION_TYPE_ADD_ENVCONFIG,
} from "../../redux/redux";
import { Menu } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/solid'
import EnvVarsTable from "./envVarsTable";
import { Switch } from '@headlessui/react'

class EnvConfig extends Component {
  constructor(props) {
    super(props);

    const { owner, repo, env, config } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    let reduxState = this.props.store.getState();

    this.state = {
      chartSchema: reduxState.chartSchema,
      chartUISchema: reduxState.chartUISchema,
      fileInfos: reduxState.fileInfos,

      saveButtonTriggered: false,
      hasAPIResponded: false,
      isError: false,
      errorMessage: "",
      isTimedOut: false,
      timeoutTimer: {},
      hasFormValidationError: false,
      environmentVariablesExpanded: false,
      codeSnippetExpanded: false,
      deployEvents: ["push", "tag", "pr"],
      selectedDeployEvent: "push",
      useDeployPolicy: false,

      envs: reduxState.envs,
      repoMetas: reduxState.repoMetas,
    };

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        chartSchema: reduxState.chartSchema,
        chartUISchema: reduxState.chartUISchema,
        fileInfos: reduxState.fileInfos,
        envs: reduxState.envs,
        repoMetas: reduxState.repoMetas,
      });

      this.ensureRepoAssociationExists(repoName, reduxState.repoMetas);

      if (!this.state.values) {
        this.setLocalEnvConfigState(reduxState, repoName, env, config);
      }
    });

    this.setValues = this.setValues.bind(this);
    this.resetNotificationStateAfterThreeSeconds = this.resetNotificationStateAfterThreeSeconds.bind(this);
  }

  setLocalEnvConfigState(reduxState, repoName, env, config) {
    let configFileContent = configFileContentFromEnvConfigs(reduxState.envConfigs, repoName, env, config);
    if (configFileContent) { // if data not loaded yet, store.subscribe will take care of this
      let envConfig = configFileContent.values;

      this.setState({
        configFile: configFileContent,

        appName: configFileContent.app,
        namespace: configFileContent.namespace,
        defaultAppName: configFileContent.app,
        defaultNamespace: configFileContent.namespace,

        values: Object.assign({}, envConfig),
        nonDefaultValues: Object.assign({}, envConfig),
        defaultState: Object.assign({}, envConfig),
      });

      this.setDeployPolicy(configFileContent.deploy);
    }
  }

  componentDidUpdate(prevProps) {
    const { owner, repo, env, config } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    if (prevProps.match.params.config !== config) {
      let reduxState = this.props.store.getState();
      this.setLocalEnvConfigState(reduxState, repoName, env, config);
    }
  }

  componentDidMount() {
    const { owner, repo, env, config } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    const { gimletClient, store } = this.props;

    gimletClient.getChartSchema(owner, repo, env)
      .then(data => {
        store.dispatch({
          type: ACTION_TYPE_CHARTSCHEMA, payload: data
        });
      }, () => {/* Generic error handler deals with it */
      });

    this.props.gimletClient.getRepoMetas(owner, repo)
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_REPO_METAS, payload: {
            repoMetas: data,
          }
        });
      }, () => {/* Generic error handler deals with it */
      });

    if (!this.state.values) { // envConfigs not loaded when we directly navigate to edit
      loadEnvConfig(gimletClient, store, owner, repo)
    }

    let reduxState = this.props.store.getState();
    this.setLocalEnvConfigState(reduxState, repoName, env, config);
  }

  ensureRepoAssociationExists(repoName, repoMetas) {
    if (this.state.defaultState && repoMetas) {
      if (!this.state.defaultState.gitSha) {
        if (repoMetas.githubActions) {
          this.setGitSha("{{ .GITHUB_SHA }}");
        }

        if (repoMetas.circleCi) {
          this.setGitSha("{{ .CIRCLE_SHA1 }}");
        }
      }

      if (!this.state.defaultState.gitRepository) {
        this.setGitRepository(repoName);
      }
    }
  }

  setGitSha(gitSha) {
    this.setState(prevState => ({
      values: {
        ...prevState.values,
        gitSha: gitSha
      },
      nonDefaultValues: {
        ...prevState.nonDefaultValues,
        gitSha: gitSha
      },
    }));
  }

  setGitRepository(repoName) {
    this.setState(prevState => ({
      values: {
        ...prevState.values,
        gitRepository: repoName
      },
      nonDefaultValues: {
        ...prevState.nonDefaultValues,
        gitRepository: repoName
      },
    }))
  }

  setDeployPolicy(deploy){
    if (deploy) {
      this.setState({
        useDeployPolicy: true,
        deployBranch: deploy.branch,
        deployTag: deploy.tag,
        selectedDeployEvent: deploy.event,
        defaultUseDeployPolicy: true,
        defaultDeployBranch: deploy.branch,
        defaultDeployTag: deploy.tag,
        defaultSelectedDeployEvent: deploy.event,
      });
      return
    }

    this.setState({
      defaultUseDeployPolicy: false,
      defaultSelectedDeployEvent: this.state.selectedDeployEvent,
    });
  }

  validationCallback = (errors) => {
    if (errors) {
      console.log(errors);
      this.setState({ hasFormValidationError: true })
    } else {
      this.setState({ hasFormValidationError: false })
    }
  }

  setValues(values, nonDefaultValues) {
    this.setState({ values: values, nonDefaultValues: nonDefaultValues });
  }

  resetNotificationStateAfterThreeSeconds() {
    setTimeout(() => {
      this.setState({
        saveButtonTriggered: false,
        hasAPIResponded: false,
        errorMessage: "",
        isError: false,
        isTimedOut: false
      });
    }, 3000);
  }

  startApiCallTimeOutHandler() {
    const timeoutTimer = setTimeout(() => {
      if (this.state.saveButtonTriggered) {
        this.setState({ isTimedOut: true, hasAPIResponded: true });
        this.resetNotificationStateAfterThreeSeconds()
      }
    }, 15000);

    this.setState({
      timeoutTimer: timeoutTimer
    })
  }

  save() {
    const { owner, repo, env, config, action } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    this.setState({ saveButtonTriggered: true });
    this.startApiCallTimeOutHandler();

    const appNameToSave = action === "new" ? this.state.appName : this.state.defaultAppName;

    this.props.gimletClient.saveEnvConfig(owner, repo, env, config, this.state.nonDefaultValues, this.state.namespace, appNameToSave, this.state.useDeployPolicy, this.state.deployBranch, this.state.deployTag, this.state.deployEvents.indexOf(this.state.selectedDeployEvent))
      .then((data) => {
        if (!this.state.saveButtonTriggered) {
          // if no saving is in progress, practically it timed out
          return
        }

        this.setDeployPolicy(data.deploy);

        clearTimeout(this.state.timeoutTimer);
        this.props.history.replace(`/repo/${owner}/${repo}/envs/${env}/config/${appNameToSave}`);
        this.setState({
          hasAPIResponded: true,

          configFile: data,

          appName: data.app,
          namespace: data.namespace,
          defaultAppName: data.app,
          defaultNamespace: data.namespace,

          values: Object.assign({}, data.values),
          nonDefaultValues: Object.assign({}, data.values),
          defaultState: Object.assign({}, data.values),
        });
        if (action === "new") {
          this.props.history.replace(`/repo/${repoName}/envs/${env}/config/${appNameToSave}`);
        }
        this.resetNotificationStateAfterThreeSeconds();
      }, err => {
        clearTimeout(this.state.timeoutTimer);
        this.setState({
          hasAPIResponded: true,
          isError: true,
          errorMessage: err.data?.message ?? err.statusText
        });
        this.resetNotificationStateAfterThreeSeconds();
      })
  }

  findFileName(envName, appName) {
    if (this.state.fileInfos.find(fileInfo => fileInfo.envName === envName && fileInfo.appName === appName)) {
      return this.state.fileInfos.find(fileInfo => fileInfo.envName === envName && fileInfo.appName === appName).fileName
    }
  }

  render() {
    const { owner, repo, env, config, action } = this.props.match.params;
    const repoName = `${owner}/${repo}`
    const configFileCopy = Object.assign({}, this.state.configFile)
    configFileCopy.values = this.state.nonDefaultValues;

    const fileName = this.findFileName(env, config)
    const nonDefaultValuesString = JSON.stringify(this.state.nonDefaultValues);
    const hasChange = (nonDefaultValuesString !== '{ }' &&
      nonDefaultValuesString !== JSON.stringify(this.state.defaultState)) ||
      this.state.namespace !== this.state.defaultNamespace || this.state.deployBranch !== this.state.defaultDeployBranch || this.state.deployTag !== this.state.defaultDeployTag || this.state.selectedDeployEvent !== this.state.defaultSelectedDeployEvent || this.state.useDeployPolicy !== this.state.defaultUseDeployPolicy || action === "new";

    if (!this.state.chartSchema) {
      return null;
    }

    if (!this.state.chartUISchema) {
      return null;
    }

    if (!this.state.values) {
      return null;
    }

    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-bold leading-tight text-gray-900">Editing {config} config for {env}
          {fileName &&
            <>
              <a href={`https://github.com/${repoName}/blob/main/.gimlet/${fileName}`} target="_blank" rel="noopener noreferrer">
                <svg xmlns="http://www.w3.org/2000/svg"
                  className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="16" height="16"
                  viewBox="0 0 24 24">
                  <path d="M0 0h24v24H0z" fill="none" />
                  <path
                    d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
                </svg>
              </a>
            </>}
        </h1>
        <h2 className="text-xl leading-tight text-gray-900">{repoName}
          <a href={`https://github.com/${repoName}`} target="_blank" rel="noopener noreferrer">
            <svg xmlns="http://www.w3.org/2000/svg"
              className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
              viewBox="0 0 24 24">
              <path d="M0 0h24v24H0z" fill="none" />
              <path
                d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
            </svg>
          </a>
        </h2>
        <button className="text-gray-500 hover:text-gray-700" onClick={() => window.location.href.indexOf(`${env}#`) > -1 ? this.props.history.go(-2) : this.props.history.go(-1)}>
          &laquo; back
        </button>
        <div className="mt-8 mb-4 items-center">
          <label htmlFor="appName" className={`${!this.state.appName ? "text-red-600" : "text-gray-700"} mr-4 block text-sm font-medium`}>
            App name*
          </label>
          <input
            type="text"
            name="appName"
            id="appName"
            disabled={action !== "new"}
            value={this.state.appName}
            onChange={e => { this.setState({ appName: e.target.value }) }}
            className={action !== "new" ? "border-0 bg-gray-100" : "mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"}
          />
        </div>
        <div className="mb-4 items-center">
          <label htmlFor="namespace" className={`${!this.state.namespace ? "text-red-600" : "text-gray-700"} mr-4 block text-sm font-medium`}>
            Namespace*
          </label>
          <input
            type="text"
            name="namespace"
            id="namespace"
            value={this.state.namespace}
            onChange={e => { this.setState({ namespace: e.target.value }) }}
            className="mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"
          />
        </div>
        <div className="mb-4 items-center">
          <div className="text-gray-700 block text-sm font-medium">Policy based releases</div>
          <div className="text-sm mb-4 text-gray-500 leading-loose">
            You can automate releases to your staging or production environment.
          </div>
          <div className="max-w-lg flex rounded-md">
            <Switch
              checked={this.state.useDeployPolicy}
              onChange={() => this.setState({ useDeployPolicy: !this.state.useDeployPolicy })}
              className={(
                this.state.useDeployPolicy ? "bg-indigo-600" : "bg-gray-200") +
                " relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200"
              }
            >
              <span className="sr-only">Use setting</span>
              <span
                aria-hidden="true"
                className={(
                  this.state.useDeployPolicy ? "translate-x-5" : "translate-x-0") +
                  " pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                }
              />
            </Switch>
          </div>
        </div>

        {this.state.useDeployPolicy &&
          <div className="ml-8">
            <div className="mb-4 items-center">
              <label htmlFor="deployEvent" className="text-gray-700 mr-4 block text-sm font-medium">
                Deploy event*
              </label>
              <Menu as="span" className="mt-2 relative inline-flex shadow-sm rounded-md align-middle">
                <Menu.Button
                  className="relative cursor-pointer inline-flex items-center px-4 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-700"
                >

                  {this.state.selectedDeployEvent}
                </Menu.Button>
                <span className="-ml-px relative block">
                  <Menu.Button
                    className="relative z-0 inline-flex items-center px-2 py-3 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500">
                    <span className="sr-only">Open options</span>
                    <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
                  </Menu.Button>
                  <Menu.Items
                    className="origin-top-right absolute z-50 left-0 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
                    <div className="py-1">
                      {this.state.deployEvents.map((deployEvent) => (
                        <Menu.Item key={`${deployEvent}`}>
                          {({ active }) => (
                            <button
                              onClick={() => {
                                this.setState({
                                  selectedDeployEvent: deployEvent,
                                  deployBranch: "",
                                  deployTag: "",
                                })
                              }}
                              className={(
                                active ? 'bg-gray-100 text-gray-900' : 'text-gray-700') +
                                ' block px-4 py-2 text-sm w-full text-left'
                              }
                            >
                              {deployEvent}
                            </button>
                          )}
                        </Menu.Item>
                      ))}
                    </div>
                  </Menu.Items>
                </span>
              </Menu>
            </div>
            {this.state.selectedDeployEvent === "tag" ?
              <div className="mb-8 items-center">
                <label htmlFor="deployTag" className="text-gray-700 mr-4 block text-sm font-medium">
                  Deploy tag*
                </label>
                <input
                  type="text"
                  name="deployTag"
                  id="deployTag"
                  value={this.state.deployTag}
                  onChange={e => { this.setState({ deployTag: e.target.value }) }}
                  className="mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"
                />
              </div>
              :
              <div className="mb-8 items-center">
                <label htmlFor="deployBranch" className="text-gray-700 mr-4 block text-sm font-medium">
                  Deploy branch*
                </label>
                <input
                  type="text"
                  name="deployBranch"
                  id="deployBranch"
                  value={this.state.deployBranch}
                  onChange={e => { this.setState({ deployBranch: e.target.value }) }}
                  className="mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"
                />
              </div>}
          </div>}
        <div className="container mx-auto m-8">
          <HelmUI
            schema={this.state.chartSchema}
            config={this.state.chartUISchema}
            values={this.state.values}
            setValues={this.setValues}
            validate={true}
            validationCallback={this.validationCallback}
          />
          <div className="w-full mt-16">
            <ReactDiffViewer
              oldValue={YAML.stringify(this.state.defaultState)}
              newValue={YAML.stringify(this.state.nonDefaultValues)}
              splitView={false}
              showDiffOnly={false}
              styles={{
                diffContainer: {
                  overflowX: "auto",
                  display: "block",
                  "& pre": { whiteSpace: "pre" }
                }
              }} />
          </div>
          {!this.state.environmentVariablesExpanded ?
            <Button
              text={"Check the list of environment variables you can use in the Gimlet manifest"}
              action={() => this.setState({ environmentVariablesExpanded: true })}
            />
            :
            <div className="w-full my-16">
              <EnvVarsTable />
              <LinkToDefaultVariables
                repoMetas={this.state.repoMetas}
              />
            </div>
          }
          {JSON.stringify(this.state.envConfig) !== "{}" &&
            <>
              {!this.state.codeSnippetExpanded ?
                <Button
                  text={"Want to render the manifest locally? Click to see the Gimlet CLI command!"}
                  action={() => this.setState({ codeSnippetExpanded: true })}
                />
                :
                <div className="my-8">
                  <h3 className="text-baseline leading-6 text-gray-500">
                    Copy the code snippet to check the generated Kubernetes manifest on the command line:
                  </h3>
                  <div className="w-full mb-16">
                    <CopiableCodeSnippet
                      copiable={true}
                      code={
                        `cat << EOF > manifest.yaml
${YAML.stringify(configFileCopy)}EOF

gimlet manifest template -f manifest.yaml`}
                    />
                  </div>
                </div>
              }
            </>}
        </div>
        <div className="p-0 flow-root">
          <span className="inline-flex gap-x-3 float-right">
            <Menu as="span" className="ml-2 relative inline-flex shadow-sm rounded-md align-middle">
              <Menu.Button
                className="relative cursor-pointer inline-flex items-center px-4 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-700"
              >
                Replicate to..
              </Menu.Button>
              <span className="-ml-px relative block">
                <Menu.Button
                  className="relative z-0 inline-flex items-center px-2 py-3 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500">
                  <span className="sr-only">Open options</span>
                  <ChevronDownIcon className="h-5 w-5" aria-hidden="true" />
                </Menu.Button>
                <Menu.Items
                  className="origin-top-right absolute z-50 right-0 mt-2 -mr-1 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
                  <div className="py-1">
                    {this.state.envs.map((env) => (
                      <Menu.Item key={`${env.name}`}>
                        {({ active }) => (
                          <button
                            onClick={() => {
                              this.props.history.push(`/repo/${repoName}/envs/${env.name}/config/${config}-copy/new`);
                              this.props.store.dispatch({
                                type: ACTION_TYPE_ADD_ENVCONFIG, payload: {
                                  repo: repoName,
                                  env: env.name,
                                  envConfig: {
                                    ...this.state.configFile,
                                    app: `${this.state.configFile.app}-copy`,
                                    env: env.name
                                  },
                                }
                              });
                            }}
                            className={(
                              active ? 'bg-gray-100 text-gray-900' : 'text-gray-700') +
                              ' block px-4 py-2 text-sm w-full text-left'
                            }
                          >
                            {env.name}
                          </button>
                        )}
                      </Menu.Item>
                    ))}
                  </div>
                </Menu.Items>
              </span>
            </Menu>
            <button
              type="button"
              disabled={!hasChange || this.state.saveButtonTriggered}
              className={(hasChange && !this.state.saveButtonTriggered ? `cursor-pointer bg-red-600 hover:bg-red-500 focus:border-red-700 focus:shadow-outline-indigo active:bg-red-700` : `bg-gray-600 cursor-default`) + ` inline-flex items-center px-6 py-2 border border-transparent text-base leading-6 font-medium rounded-md text-white focus:outline-none transition ease-in-out duration-150`}
              onClick={() => {
                this.setState({ values: Object.assign({}, this.state.defaultState) });
                this.setState({ nonDefaultValues: Object.assign({}, this.state.defaultState) });
                this.setState({ namespace: this.state.defaultNamespace })
                this.setState({ useDeployPolicy: this.state.defaultUseDeployPolicy })
                this.setState({ deployBranch: this.state.defaultDeployBranch })
                this.setState({ deployTag: this.state.defaultDeployTag })
                this.setState({ selectedDeployEvent: this.state.defaultSelectedDeployEvent })
              }}
            >
              Reset
            </button>
            <button
              type="button"
              disabled={!hasChange || !this.state.namespace || !this.state.appName || this.state.saveButtonTriggered || this.state.hasFormValidationError}
              className={(hasChange && this.state.namespace && this.state.appName && !this.state.saveButtonTriggered && !this.state.hasFormValidationError ? 'bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700' : `bg-gray-600 cursor-default`) + ` inline-flex items-center px-6 py-2 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150`}
              onClick={() => this.save()}
            >
              Save
            </button>
            {this.state.saveButtonTriggered &&
              <PopUpWindow
                hasAPIResponded={this.state.hasAPIResponded}
                errorMessage={this.state.errorMessage}
                isError={this.state.isError}
                isTimedOut={this.state.isTimedOut}
              />
            }
          </span>
        </div>
      </div>
    );
  }
}

function Button({ text, action }) {
  return (
    <div>
      <button className="cursor-pointer text-xs leading-6 text-blue-500 hover:text-blue-700"
        onClick={action}
      >
        {text}
      </button>
    </div>)
}

function LinkToDefaultVariables({ repoMetas }) {
  if (!repoMetas.githubActions && !repoMetas.circleCi) {
    return null
  }

  let defaultVariablesUrl = "";

  if (repoMetas.githubActions) {
    defaultVariablesUrl = "https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables"
  } else if (repoMetas.circleCi) {
    defaultVariablesUrl = "https://circleci.com/docs/env-vars?section=pipelines&utm_source=google&utm_medium=sem&utm_campaign=sem-google-dg--emea-en-dsa-maxConv-auth-brand&utm_term=g_-_c__dsa_&utm_content=&gclid=Cj0KCQjwz96WBhC8ARIsAATR251pCKLp8uHHmudeI2J3nRulg38fcPRscyjM0KdiomXQsvsFEMJ-NsIaAgFkEALw_wcB#built-in-environment-variables"
  }

  return (
    <div className="mt-2">
      <a
        href={defaultVariablesUrl}
        target="_blank"
        rel="noreferrer"
        className="text-gray-500 hover:text-gray-700 text-xs"
      >
        Additionally you can use all built-in environment variables from CI
      </a>
    </div>
  )
}

function configFileContentFromEnvConfigs(envConfigs, repoName, env, config) {
  if (envConfigs[repoName]) {
    if (envConfigs[repoName][env]) {
      const configFileContentFromEnvConfigs = envConfigs[repoName][env].filter(c => c.app === config)
      if (configFileContentFromEnvConfigs.length > 0) {
        return configFileContentFromEnvConfigs[0]
      } else {
        // "envConfigs loaded, we have data for env, but we don't have config for app"
        return {}
      }
    } else {
      // "envConfigs loaded, but we don't have data for env"
      return {}
    }
  } else {
    // envConfigs not loaded, we shall wait for it to be loaded
    return undefined
  }
}

function loadEnvConfig(gimletClient, store, owner, repo) {
  gimletClient.getEnvConfigs(owner, repo)
    .then(envConfigs => {
      store.dispatch({
        type: ACTION_TYPE_ENVCONFIGS, payload: {
          owner: owner,
          repo: repo,
          envConfigs: envConfigs
        }
      });
    }, () => {/* Generic error handler deals with it */
    });
}

export default EnvConfig;
