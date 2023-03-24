import React, { Component } from 'react'
import { hot } from 'react-hot-loader'
import * as schemaFixture from '../fixtures/onechart/values.schema.json'
import * as helmUIConfigFixture from '../fixtures/onechart/helm-ui.json'
import HelmUI from 'helm-react-ui'
import './style.css'
import StreamingBackend from './streamingBackend'
import GimletCLIClient from './client'

class App extends Component {
  constructor (props) {
    super(props)

    const client = new GimletCLIClient()
    client.onError = (response) => {
      console.log(response)
      console.log(`${response.status}: ${response.statusText} on ${response.path}`)
    }

    this.state = {
      client: client,
      values: {},
      nonDefaultValues: {},
      defaultApp: "",
      app: "",
      defaultEnv: "",
      env: "",
      namespace: "",
    }
    this.setValues = this.setValues.bind(this)
  }

  componentDidMount () {
    fetch('/helm-ui.json')
      .then(response => {
        if (!response.ok && window !== undefined) {
          console.log("Using fixture")
          return helmUIConfigFixture.default
        }
        return response.json()
      })
      .then(data => this.setState({ helmUISchema: data }))

    fetch('/values.schema.json')
      .then(response => {
        if (!response.ok && window !== undefined) {
          console.log("Using fixture")
          return schemaFixture.default
        }
        return response.json()
      })
      .then(data => this.setState({ schema: data }))

    fetch('/values.json')
      .then(response => {
        if (!response.ok && window !== undefined) {
          console.log("Using fixture")
          return {
            defaultApp: "",
            app: "",
            defaultEnv: "",
            env: "",
            namespace: "default",
            values: {},
          }
        }
        return response.json()
      })
      .then(data => {
        this.setState({ defaultApp: data.app });
        this.setState({ app: data.app });
        this.setState({ defaultEnv: data.env });
        this.setState({ env: data.env });
        this.setState({ namespace: data.namespace || "default" });
        this.setState({ values: data.values ?? {} });
      })
  }

  componentDidUpdate() {
    this.state.client.saveValues({
      app: this.state.app,
      env: this.state.env,
      namespace: this.state.namespace,
      values: this.state.nonDefaultValues,
    });
  }

  setValues (values, nonDefaultValues) {
    this.setState({ values: values, nonDefaultValues: nonDefaultValues })
  }

  validationCallback (errors) {
    if (errors !== null) {
      console.log(errors)
    }
  };

  render () {
    let { schema, helmUISchema, values } = this.state

    if (schema === undefined || helmUISchema === undefined || values === undefined) {
      return null;
    }

    return (
      <div>
        <StreamingBackend client={this.state.client} />
        {(this.state.app === "" || this.state.env === "" || this.state.namespace === "") &&
          <div className="fixed top-0 right-0">
            <span className="inline-flex rounded-md shadow-sm m-8">
              <div
                type="button"
                className="cursor-default inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white bg-red-600 transition ease-in-out duration-150"
              >
                Validation error!
              </div>
            </span>
          </div>}
        <div className="fixed bottom-0 right-0">
          <span className="inline-flex rounded-md shadow-sm m-8">
            <button
              type="button"
              className="cursor-default inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white bg-gray-600 transition ease-in-out duration-150"
              onClick={() => {
                console.log(this.state.values)
                console.log(this.state.nonDefaultValues)
              }}
            >
              Close the browser when you are done, the values will be printed on the console
            </button>
          </span>
        </div>
        <div className="container mx-auto m-8">
          <div className="y-6 px-2 sm:px-6 lg:py-0 lg:px-0">
            <div className="mt-8 mb-4 items-center">
              <label htmlFor="appName" className={`${!this.state.app ? "text-red-600" : "text-gray-700"} mr-4 block text-sm font-medium`}>
                App name*
              </label>
              <input
                type="text"
                name="appName"
                id="appName"
                disabled={this.state.defaultApp !== ""}
                value={this.state.app}
                onChange={e => { this.setState({ app: e.target.value }) }}
                className={this.state.defaultApp !== "" ? "border-0 bg-gray-100" : "mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"}
              />
            </div>
            <div className="mt-4 mb-4 items-center">
              <label htmlFor="appName" className={`${!this.state.env ? "text-red-600" : "text-gray-700"} mr-4 block text-sm font-medium`}>
                Env name*
              </label>
              <input
                type="text"
                name="appName"
                id="appName"
                disabled={this.state.defaultEnv !== ""}
                value={this.state.env}
                onChange={e => { this.setState({ env: e.target.value }) }}
                className={this.state.defaultEnv !== "" ? "border-0 bg-gray-100" : "mt-2 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 border-gray-300 rounded-md w-4/12"}
              />
            </div>
            <div className="mt-4 mb-8 items-center">
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
          </div>
          <HelmUI
            schema={schema}
            config={helmUISchema}
            values={values}
            setValues={this.setValues}
            validate={true}
            validationCallback={this.validationCallback}
          />
        </div>
      </div>
    )
  }
}

export default hot(module)(App)
