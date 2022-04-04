import React, { Component } from "react";
import HelmUI from "helm-react-ui";
import "./style.css";
import PopUpWindow from "./popUpWindow";
import ReactDiffViewer from "react-diff-viewer";
import YAML from "json-to-pretty-yaml";
import {
  ACTION_TYPE_ENVCONFIGS,
} from "../../redux/redux";

class EnvConfig extends Component {
  constructor(props) {
    super(props);

    const { owner, repo, env, config } = this.props.match.params;
    const repoName = `${owner}/${repo}`;

    let reduxState = this.props.store.getState();

    let envConfig = configFromEnvConfigs(reduxState.envConfigs, repoName, env, config);
    let defaultNamespace = namespaceFromEnvConfigs(reduxState.envConfigs, repoName, env, config);

    this.state = {
      chartSchema: reduxState.chartSchema,
      chartUISchema: reduxState.chartUISchema,
      envConfigs: reduxState.envConfigs,

      saveButtonTriggered: false,
      hasAPIResponded: false,
      isError: false,
      errorMessage: "",
      isTimedOut: false,
      timeoutTimer: {},
      defaultNamespace: defaultNamespace,
      namespace: "",

      values: envConfig ? Object.assign({}, envConfig) : undefined,
      nonDefaultValues: envConfig ? Object.assign({}, envConfig) : undefined,
      defaultState: envConfig ? Object.assign({}, envConfig) : undefined,
    };

    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      let envConfig = configFromEnvConfigs(reduxState.envConfigs, repoName, env, config);
      let defaultNamespace = namespaceFromEnvConfigs(reduxState.envConfigs, repoName, env, config);

      this.setState({
        chartSchema: reduxState.chartSchema,
        chartUISchema: reduxState.chartUISchema,
        envConfigs: reduxState.envConfigs
      });

      if (!this.state.values) {
        this.setState({
          values: envConfig ? Object.assign({}, envConfig) : undefined,
          nonDefaultValues: envConfig ? Object.assign({}, envConfig) : undefined,
          defaultState: envConfig ? Object.assign({}, envConfig) : undefined,
        });
      }

      if (!this.state.namespace) {
        this.setState({ namespace: defaultNamespace })
      }

      if (!this.state.defaultNamspace) {
        this.setState({ defaultNamespace: defaultNamespace })
      }
    });

    this.setValues = this.setValues.bind(this);
    this.resetNotificationStateAfterThreeSeconds = this.resetNotificationStateAfterThreeSeconds.bind(this);
  }

  componentDidMount() {
    const { owner, repo } = this.props.match.params;
    const { gimletClient, store } = this.props;

    if (!this.state.values) { // envConfigs not loaded when we directly navigate to edit
      loadEnvConfig(gimletClient, store, owner, repo)
    }
  }

  validationCallback(errors) {
    if (errors) {
      console.log(errors);
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
    const { owner, repo, env, config } = this.props.match.params;

    this.setState({ saveButtonTriggered: true });
    this.startApiCallTimeOutHandler();

    this.props.gimletClient.saveEnvConfig(owner, repo, env, config, this.state.nonDefaultValues, this.state.namespace)
      .then(data => {
        if (!this.state.saveButtonTriggered) {
          // if no saving is in progress, practically it timed out
          return
        }

        clearTimeout(this.state.timeoutTimer);
        this.setState({
          hasAPIResponded: true,
          defaultState: Object.assign({}, this.state.nonDefaultValues),
          defaultNamespace: this.state.namespace
        });
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

  render() {
    const { owner, repo, env, config } = this.props.match.params;
    const repoName = `${owner}/${repo}`

    const nonDefaultValuesString = JSON.stringify(this.state.nonDefaultValues);
    const hasChange = (nonDefaultValuesString !== '{ }' &&
      nonDefaultValuesString !== JSON.stringify(this.state.defaultState)) ||
      this.state.namespace !== this.state.defaultNamespace;

    if (!this.state.chartSchema) {
      return null;
    }

    if (!this.state.chartUISchema) {
      return null;
    }

    if (!this.state.values) {
      return null;
    }

    // console.log("DEFAULT STATE")
    // console.log(this.state.defaultState)
    // console.log("NON DEFAULT VALUES")
    // console.log(this.state.nonDefaultValues)
    // console.log("HAS CHANGE STATUS")
    // console.log(hasChange)
    // console.log("NAMESAPCE")
    // console.log(this.state.namespace)
    // console.log("VALUES")
    // console.log(this.state.values)
    // console.log("NON DEFAULT VALUES")
    // console.log(this.state.nonDefaultValues)
    // console.log("DEFAULT STATE")
    // console.log(this.state.defaultState)
    console.log("ENV CONFIG NAMESPACE FROM GIT")
    console.log(defaultNamespace)
    console.log("NAMESPACE STATE")
    console.log(this.state.namespace)

    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-bold leading-tight text-gray-900">Editing {config} config for {env}</h1>
        <h2 className="text-xl leading-tight text-gray-900">{repoName}
          <a href={`https://github.com/${owner}/${repo}`} target="_blank" rel="noopener noreferrer">
            <svg xmlns="http://www.w3.org/2000/svg"
              className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
              viewBox="0 0 24 24">
              <path d="M0 0h24v24H0z" fill="none" />
              <path
                d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
            </svg>
          </a>
        </h2>
        <button className="text-gray-500 hover:text-gray-700 mt-8" onClick={() => window.location.href.indexOf(`${env}#`) > -1 ? this.props.history.go(-2) : this.props.history.go(-1)}>
          &laquo; back
        </button>
        <div className="flex mt-4 mb-10 items-center">
          <label htmlFor="namespace" className={`${!this.state.namespace ? "text-red-600" : "text-gray-700"} mr-4 block text-sm font-medium`}>
            Namespace*
          </label>
          <input
            type="text"
            name="namespace"
            id="namespace"
            value={this.state.namespace}
            onChange={e => { this.setState({ namespace: e.target.value }) }}
            className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"
          />
        </div>
        <div className="container mx-auto m-8">
          <HelmUI
            schema={this.state.chartSchema}
            config={this.state.chartUISchema}
            values={this.state.values}
            setValues={this.setValues}
            validate={true}
            validationCallback={this.validationCallback}
          />
          <div className="w-full my-16">
            <ReactDiffViewer
              oldValue={YAML.stringify(this.state.defaultState)}
              newValue={YAML.stringify(this.state.nonDefaultValues)}
              splitView={false}
              showDiffOnly={false} />
          </div>
        </div>
        <div className="p-0 flow-root">
          <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
            <button
              type="button"
              disabled={!hasChange || this.state.saveButtonTriggered}
              className={(hasChange && !this.state.saveButtonTriggered ? `cursor-pointer bg-red-600 hover:bg-red-500 focus:border-red-700 focus:shadow-outline-indigo active:bg-red-700` : `bg-gray-600 cursor-default`) + ` inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white focus:outline-none transition ease-in-out duration-150`}
              onClick={() => {
                this.setState({ values: Object.assign({}, this.state.defaultState) });
                this.setState({ nonDefaultValues: Object.assign({}, this.state.defaultState) });
                this.setState({ namespace: this.state.defaultNamespace })
              }}
            >
              Reset
            </button>
            <button
              type="button"
              disabled={!hasChange || !this.state.namespace || this.state.saveButtonTriggered}
              className={(hasChange && this.state.namespace && !this.state.saveButtonTriggered ? 'bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700' : `bg-gray-600 cursor-default`) + ` inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150`}
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

function configFromEnvConfigs(envConfigs, repoName, env, config) {
  if (envConfigs[repoName]) { // envConfigs are loaded
    if (envConfigs[repoName][env]) { // we have env data
      const configFromEnvConfigs = envConfigs[repoName][env].filter(c => c.app === config)
      if (configFromEnvConfigs.length > 0) {
        // "envConfigs loaded, we have data for env, we have config for app"
        return configFromEnvConfigs[0].values
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

function namespaceFromEnvConfigs(envConfigs, repoName, env, config) {
  if (envConfigs[repoName]) {
    if (envConfigs[repoName][env]) {
      const namespaceFromEnvConfigs = envConfigs[repoName][env].filter(c => c.app === config)
      if (namespaceFromEnvConfigs.length > 0) {
        return namespaceFromEnvConfigs[0].namespace
      } else {
        return ""
      }
    } else {
      return ""
    }
  } else {
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
