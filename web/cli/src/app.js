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
          return {}
        }
        return response.json()
      })
      .then(data => this.setState({ values: data }))
  }

  setValues (values, nonDefaultValues) {
    this.setState({ values: values, nonDefaultValues: nonDefaultValues })
    this.state.client.saveValues(nonDefaultValues)
  }

  render () {
    let { schema, helmUISchema, values } = this.state

    if (schema === undefined || helmUISchema === undefined || values === undefined) {
      return null;
    }

    return (
      <div>
        <StreamingBackend client={this.state.client}/>
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
          <HelmUI
            schema={schema}
            config={helmUISchema}
            values={values}
            setValues={this.setValues}
          />
        </div>
      </div>
    )
  }
}

export default hot(module)(App)
