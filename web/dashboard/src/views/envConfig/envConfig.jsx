import { Component } from "react";
import HelmUI from "helm-react-ui";
import "./style.css";
import ReactDiffViewer from "react-diff-viewer";
import yaml from "js-yaml";
import { Spinner } from "../repositories/repositories";
import {
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWERRORLIST,
  ACTION_TYPE_SAVE_REPO_PULLREQUEST
} from "../../redux/redux";
import { Menu } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/solid'
import { Switch } from '@headlessui/react'
import posthog from "posthog-js"
import ImageWidget from "./imageWidget";

class EnvConfig extends Component {
  constructor(props) {
    super(props);

    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    let reduxState = this.props.store.getState();

    this.state = {
      timeoutTimer: {},
      deployEvents: ["push", "tag", "pr"],
      popupWindow: reduxState.popupWindow,
      scmUrl: reduxState.settings.scmUrl,
      envs: reduxState.envs,
    };

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        envs: reduxState.envs,
        popupWindow: reduxState.popupWindow,
        scmUrl: reduxState.settings.scmUrl
      });

      this.ensureRepoAssociationExists(repoName);
      this.ensureGitCloneUrlExists();
    });

    this.setValues = this.setValues.bind(this);
    this.resetNotificationStateAfterThreeSeconds = this.resetNotificationStateAfterThreeSeconds.bind(this);
  }

  componentDidMount() {
    const { owner, repo, env, config, action } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    const { gimletClient } = this.props;

    if (action === "new") {
      gimletClient.getDefaultDeploymentTemplates()
      .then(data => {
        const selectedTemplate = this.patchImageWidget(data[0])
        this.setState({
          templates: data,
          selectedTemplate: selectedTemplate,
          configFile: {
            app: config,
            namespace: "default",
            env:       env,
            chart: selectedTemplate.reference,
            values: {
              gitRepository: repoName,
              gitSha:        "{{ .SHA }}",
              image: {
                repository: "127.0.0.1:32447/"+config,
                tag:        "{{ .SHA }}",
              },
              resources: {
                ignoreLimits: true,
              },
            },
          },
          defaultConfigFile: {},
        });
      }, () => {/* Generic error handler deals with it */ });
    } else {
      gimletClient.getDeploymentTemplates(owner, repo, env, config)
      .then(data => {
        this.setState({
          templates: data,
          selectedTemplate: this.patchImageWidget(data[0])
        });
      }, () => {/* Generic error handler deals with it */ });

      this.props.gimletClient.getEnvConfigs(owner, repo)
        .then(envConfigs => {         
          if (envConfigs[env]) {
            const configFileContentFromEnvConfigs = envConfigs[env].find(c => c.app === config)
            let deepCopied = JSON.parse(JSON.stringify(configFileContentFromEnvConfigs))
            this.setState({
              configFile: configFileContentFromEnvConfigs,
              defaultConfigFile: deepCopied,
            });
          }
        }, () => {/* Generic error handler deals with it */
      });
    }
  }

  ensureRepoAssociationExists() {
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    if (this.state.configFile) {
      if (!this.state.configFile.values.gitSha) {
        this.setState(prevState => ({
          configFile: {
            ...prevState.configFile,
            values: {
              ...prevState.configFile.values,
              gitSha: "{{ .SHA }}"
            },
          },
        }));
      }

      if (!this.state.configFile.values.gitRepository) {
        this.setState(prevState => ({
          configFile: {
            ...prevState.configFile,
            values: {
              ...prevState.configFile.values,
              gitRepository: repoName
            },
          },
        }));
      }
    }
  }

  setAppName(appName) {
    this.setState(prevState => ({
      configFile: {
        ...prevState.configFile,
        app: appName,
      },
    }));
  }

  setNamespace(namespace) {
    this.setState(prevState => ({
      configFile: {
        ...prevState.configFile,
        namespace: namespace,
      },
    }));
  }

  setDeployFilter(filter) {
    this.setState(prevState => {
      if (prevState.configFile.deploy.event === "tag") {
        return {
          configFile: {
            ...prevState.configFile,
            deploy: {
              ...prevState.configFile.deploy,
              tag: filter
            },
          },
        }
      }

      return {
        configFile: {
          ...prevState.configFile,
          deploy: {
            ...prevState.configFile.deploy,
            branch: filter
          },
        },
      }
    });
  }

  setDeployEvent(deployEvent) {
    this.setState(prevState => ({
      configFile: {
        ...prevState.configFile,
        deploy: {
          event: deployEvent
        },
      },
    }));
  }

  toggleDeployPolicy() {
    this.setState(prevState => ({
      configFile: {
        ...prevState.configFile,
        deploy: prevState.configFile.deploy ? undefined : {event: "push"},
      },
    }));
  }

  ensureGitCloneUrlExists() {
    const { owner, repo } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    if (this.state.selectedTemplate) {
      if (this.state.selectedTemplate.reference.name === "static-site") {
        this.setState(prevState => {
          if (prevState.configFile.values.gitCloneUrl) {
            return prevState
          }
          return {
            configFile: {
              ...prevState.configFile,
              values: {
                ...prevState.configFile.values,
                gitCloneUrl: `${this.state.scmUrl}/${repoName}.git`
              },
            },
          }
        });
      }
    }
  }

  patchImageWidget(chart) {
    if (chart.reference.name !== "onechart") {
      return chart  
    }

    if (!chart.uiSchema[0].uiSchema["#/properties/image"]) {
      chart.uiSchema[0].uiSchema = {
        ...chart.uiSchema[0].uiSchema,
        "#/properties/image": {
          "ui:field": "imageWidget"
        },
      }
    }
   
    return chart
  }

  setDeployPolicy(deploy) {
    if (deploy) {
      this.setState({
        useDeployPolicy: true,
        deployFilterInput: deploy.branch ?? deploy.tag,
        selectedDeployEvent: deploy.event,
        defaultUseDeployPolicy: true,
        defaultDeployFilterInput: deploy.branch ?? deploy.tag,
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
      this.setState({ errors: errors });
      this.displayErrors(errors);
    } else {
      this.setState({ errors: undefined });

      this.props.store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }
  }

  displayErrors(errors) {
    this.props.store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWERRORLIST, payload: {
        header: "Error",
        errorList: errors
      }
    });
  }

  setValues(values) {
    this.setState(prevState => ({
      configFile: {
        ...prevState.configFile,
        values: values
      }
    }));
  }

  resetNotificationStateAfterThreeSeconds() {
    setTimeout(() => {
      this.props.store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  }

  startApiCallTimeOutHandler() {
    const timeoutTimer = setTimeout(() => {
      if (this.state.popupWindow.visible) {
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: "Saving failed: The process has timed out."
          }
        });
        this.resetNotificationStateAfterThreeSeconds()
      }
    }, 60000);

    this.setState({
      timeoutTimer: timeoutTimer
    })
  }

  save() {
    if (this.state.errors) {
      this.displayErrors(this.state.errors);
      return
    }

    const { owner, repo, env, config } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    this.props.store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Saving..."
      }
    });
    this.startApiCallTimeOutHandler();

    this.props.gimletClient.saveEnvConfig(owner, repo, env, encodeURIComponent(config), this.state.configFile.values, this.state.configFile.namespace, this.state.configFile.chart, this.state.configFile.app, this.state.configFile.deploy ? true : false, this.state.configFile.deploy?.branch, this.state.configFile.deploy?.tag, this.state.configFile.deploy?.event)
      .then((data) => {
        if (!this.state.popupWindow.visible) {
          // if no saving is in progress, practically it timed out
          return
        }

        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Pull request was created",
            link: data.createdPr.link
          }
        });
        this.props.store.dispatch({
          type: ACTION_TYPE_SAVE_REPO_PULLREQUEST,
          payload: {
            repoName: repoName,
            envName: env,
            createdPr: data.createdPr
          }
        });

        clearTimeout(this.state.timeoutTimer);
        this.props.history.push(`/repo/${repoName}`);
        window.scrollTo({ top: 0, left: 0 });
      }, err => {
        clearTimeout(this.state.timeoutTimer);
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.data?.message ?? err.statusText
          }
        });
        this.resetNotificationStateAfterThreeSeconds();
      })
  }

  delete() {
    const { owner, repo, config, env } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    this.props.store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Deleting..."
      }
    });
    this.startApiCallTimeOutHandler();

    this.props.gimletClient.deleteEnvConfig(owner, repo, env, config)
      .then((data) => {
        if (!this.state.popupWindow.visible) {
          // if no deleting is in progress, practically it timed out
          return
        }

        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "Pull request was created",
            link: data.createdPr.link
          }
        });
        this.props.store.dispatch({
          type: ACTION_TYPE_SAVE_REPO_PULLREQUEST,
          payload: {
            repoName: repoName,
            envName: env,
            createdPr: data.createdPr
          }
        });

        clearTimeout(this.state.timeoutTimer);
        this.props.history.replace(`/repo/${repoName}`);
        window.scrollTo({ top: 0, left: 0 });
      }, err => {
        clearTimeout(this.state.timeoutTimer);
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.data?.message ?? err.statusText
          }
        });
        this.resetNotificationStateAfterThreeSeconds();
      })
  }

  setDeploymentTemplate(template) {
    const selectedTemplate = this.patchImageWidget(template)
    this.setState(prevState => {
      let copiedConfigFile = Object.assign({}, prevState.configFile)
      delete copiedConfigFile.deploy
      copiedConfigFile.values = {}
      copiedConfigFile.chart = selectedTemplate.reference

      return {
        configFile: copiedConfigFile,
        selectedTemplate: selectedTemplate,
      }
    }, () => {
      this.ensureRepoAssociationExists();
      this.ensureGitCloneUrlExists();
    })
  }

  renderTemplateFromConfig() {
    let title = "Web application template"
    let description = "To deploy any web application. Multiple image build options available."
    if (this.state.selectedTemplate.reference.name === "static-site") {
      title = "Static site template"
      description = "If your build generates static files only, let us host it in an Nginx container."
    }

    return (
      <div className="mb-6 grid grid-cols-1 gap-y-6 sm:grid-cols-3 sm:gap-x-4">
        <div className="relative flex rounded-lg p-4 focus:outline-none bg-white text-gray-500">
          <span className="flex flex-1">
            <span className="flex flex-col">
              <span id="project-type-0-label" className="block text-sm font-medium text-gray-900 select-none">{title}</span>
              <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm select-none">{description}</span>
            </span>
          </span>
        </div>
      </div>
    )
  }

  render() {
    const { owner, repo, env, config, action } = this.props.match.params;
    const repoName = `${owner}/${repo}`

    const hasChange = JSON.stringify(this.state.configFile) !== JSON.stringify(this.state.defaultConfigFile)

    const customFields = {
      imageWidget: ImageWidget,
    }

    if (!this.state.configFile) {
      return <Spinner />;
    }

    if (!this.state.selectedTemplate) {
      return <Spinner />;
    }

    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-bold leading-tight text-gray-900">Editing {config} config for {env}</h1>
        <h2 className="text-xl leading-tight text-gray-900">{repoName}
          <a href={`${this.state.scmUrl}/${repoName}`} target="_blank" rel="noopener noreferrer">
            <svg xmlns="http://www.w3.org/2000/svg"
              className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
              viewBox="0 0 24 24">
              <path d="M0 0h24v24H0z" fill="none" />
              <path
                d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
            </svg>
          </a>
        </h2>
        <button className="text-gray-500 hover:text-gray-700" onClick={() => this.props.history.push(`/repo/${repoName}`)}>
          &laquo; back
        </button>

        <div className="mt-16 mb-16">
          {action === "new" ?
            <div className="mb-16 items-center">
              <div className="mt-4 grid grid-cols-1 gap-y-6 sm:grid-cols-3 sm:gap-x-4">
                {this.state.templates.map((template) => {
                  let title = "Web application template"
                  let description = "To deploy any web application. Multiple image build options available."
                  if (template.reference.name === "static-site") {
                    title = "Static site template"
                    description = "If your build generates static files only, let us host it in an Nginx container."
                  }
                  return (
                    <div
                      key={template.reference.name + template.reference.repository + template.reference.version}
                      className={`relative flex cursor-pointer rounded-lg bg-white p-4 shadow-lg focus:outline-none text-gray-500 ${this.state.selectedTemplate.reference.name  === template.reference.name ? "border border-blue-500" : "bg-gray-300 opacity-50 text-gray-600"}`}
                      onClick={() => this.setDeploymentTemplate(template)}
                    >
                      <span className="flex flex-1">
                        <span className="flex flex-col">
                          <span id="project-type-0-label" className="block text-sm font-medium text-gray-900 select-none">{title}</span>
                          <span id="project-type-0-description-0" className="mt-1 flex items-center text-sm select-none">{description}</span>
                        </span>
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
            :
            this.renderTemplateFromConfig()
          }
        <div className="mb-4 items-center">
          <label htmlFor="appName" className={`${!this.state.configFile.app ? "text-red-600" : "text-gray-700"} mr-4 block text-sm font-medium`}>
            App name*
          </label>
          <input
            type="text"
            name="appName"
            id="appName"
            disabled={action !== "new"}
            value={this.state.configFile.app}
            onChange={e => this.setAppName(e.target.value)}
            className={action !== "new" ? "border-0 bg-gray-100" : "mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"}
          />
        </div>
        <div className="mb-4 items-center">
          <label htmlFor="namespace" className={`${!this.state.configFile.namespace ? "text-red-600" : "text-gray-700"} mr-4 block text-sm font-medium`}>
            Namespace*
          </label>
          <input
            type="text"
            name="namespace"
            id="namespace"
            value={this.state.configFile.namespace}
            onChange={e => this.setNamespace(e.target.value)}
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
              key={this.state.configFile.deploy}
              checked={this.state.configFile.deploy !== undefined}
              onChange={e => this.toggleDeployPolicy()}
              className={(
                this.state.configFile.deploy !== undefined ? "bg-indigo-600" : "bg-gray-200") +
                " relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200"
              }
            >
              <span className="sr-only">Use setting</span>
              <span
                aria-hidden="true"
                className={(
                  this.state.configFile.deploy !== undefined ? "translate-x-5" : "translate-x-0") +
                  " pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                }
              />
            </Switch>
          </div>
        </div>

        {this.state.configFile.deploy &&
          <div className="ml-8 mb-8">
            <div className="mb-4 items-center">
              <label htmlFor="deployEvent" className="text-gray-700 mr-4 block text-sm font-medium">
                Deploy event*
              </label>
              <Menu as="span" className="mt-2 relative inline-flex shadow-sm rounded-md align-middle">
                <Menu.Button
                  className="relative cursor-pointer inline-flex items-center px-4 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-700"
                >

                  {this.state.configFile.deploy.event}
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
                                this.setDeployEvent(deployEvent)
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
            <div className="mb-4 items-center">
              <label htmlFor="deployFilterInput" className="text-gray-700 mr-4 block text-sm font-medium">
                {`${this.state.configFile.deploy.event === "tag" ? "Tag" : "Branch"} filter`}
              </label>
              <input
                key={this.state.configFile.deploy.event}
                type="text"
                name="deployFilterInput"
                id="deployFilterInput"
                value={this.state.configFile.deploy.event === "tag" ? this.state.configFile.deploy.tag : this.state.configFile.deploy.branch}
                onChange={e => { this.setDeployFilter(e.target.value)}}
                className="mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"
              />
              <ul className="list-none text-sm text-gray-500 mt-2">
                {this.state.configFile.deploy.event === "tag" ?
                  <>
                    <li>
                      Filter tags to deploy based on tag name patterns.
                    </li>
                    <li>
                      Use glob patterns like <code>`v1.*`</code> or negated conditions like <code>`!v2.*`</code>.
                    </li>
                  </>
                  :
                  <>
                    <li>
                      Filter branches to deploy based on branch name patterns.
                    </li>
                    <li>
                      Use glob patterns like <code>`feature/*`</code> or negated conditions like <code>`!main`</code>.
                    </li>
                  </>}
              </ul>
            </div>
          </div>
        }
        </div>
        <div className="container mx-auto m-8">
          <HelmUI
            key={this.state.selectedTemplate.reference.name + this.state.selectedTemplate.reference.repository + this.state.selectedTemplate.reference.version}
            schema={this.state.selectedTemplate.schema}
            config={this.state.selectedTemplate.uiSchema}
            fields={customFields}
            values={this.state.configFile.values}
            setValues={this.setValues}
            validate={true}
            validationCallback={this.validationCallback}
          />
          <div className="w-full mt-16">
            <ReactDiffViewer
              oldValue={yaml.dump(this.state.defaultConfigFile)}
              newValue={yaml.dump(this.state.configFile)}
              splitView={false}
              showDiffOnly={false}
              styles={{
                diffContainer: {
                  overflowX: "auto",
                  display: "block",
                  "& pre": { whiteSpace: "pre" }
                },
                emptyLine: { background: "#fff" }
              }} />
          </div>
        </div>
        <div className="p-0 flow-root my-16">
          {action !== "new" &&
            <span className="inline-flex gap-x-3 float-left">
              <button
                type="button"
                className="bg-red-600 hover:bg-red-500 focus:outline-none focus:border-red-700 focus:shadow-outline-indigo active:bg-red-700 inline-flex items-center px-6 py-2 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150"
                onClick={() => {
                  // eslint-disable-next-line no-restricted-globals
                  confirm(`Are you sure you want to delete the ${this.state.appName} deployment configuration? (deployed app instances of this configuration will remain deployed, and you can delete them later)`) &&
                    this.delete()
                }}
              >
                Delete
              </button>
            </span>}
          <span className="inline-flex gap-x-3 float-right">
            { action !== "new" &&
            <button
              type="button"
              disabled={!hasChange || this.state.popupWindow.visible}
              className={(hasChange && !this.state.popupWindow.visible ? `cursor-pointer bg-blue-600 hover:bg-blue-500 focus:border-yellow-700 focus:shadow-outline-indigo active:bg-blue-700` : `bg-gray-600 cursor-default`) + ` inline-flex items-center px-6 py-2 border border-transparent text-base leading-6 font-medium rounded-md text-white focus:outline-none transition ease-in-out duration-150`}
              onClick={() => {
                let deepCopied = JSON.parse(JSON.stringify(this.state.defaultConfigFile))
                this.setState({ configFile: deepCopied });
                this.setState({
                  selectedTemplate: this.patchImageWidget(this.state.templates[0])
                });
              }}
            >
              Reset
            </button>
            }
            <button
              type="button"
              disabled={!hasChange || this.state.popupWindow.visible || !this.state.configFile.namespace || !this.state.configFile.app}
              className={(hasChange && !this.state.popupWindow.visible && this.state.configFile.namespace && this.state.configFile.app ? 'bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700' : `bg-gray-600 cursor-default`) + ` inline-flex items-center px-6 py-2 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150`}
              onClick={() => {
                posthog?.capture('Env config save pushed')
                this.save()
              }}
            >
              Save
            </button>
          </span>
        </div>
      </div>
    );
  }
}

export default EnvConfig;
