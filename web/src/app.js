import React, { Component } from 'react'
import { hot } from 'react-hot-loader'
import * as schema from '../fixtures/onechart/values.schema.json'
import * as helmUIConfig from '../fixtures/onechart/helm-ui.json'
import HelmUI from 'helm-react-ui'
import './style.css'

class App extends Component {
  constructor (props) {
    super(props)

    this.state = {
      values: {},
      nonDefaultValues: {}
    }
    this.setValues = this.setValues.bind(this);
  }

  setValues (values, nonDefaultValues) {
    this.setState({ values: values, nonDefaultValues: nonDefaultValues })
  }

  render () {
    return (
      <div>
        <div className="fixed bottom-0 right-0">
          <span className="inline-flex rounded-md shadow-sm m-8">
            <button
              type="button"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white bg-red-600 hover:bg-red-500 focus:outline-none focus:border-red-700 focus:shadow-outline-indigo active:bg-red-700 transition ease-in-out duration-150"
              onClick={() => {console.log(this.state.values); console.log(this.state.nonDefaultValues)}}
            >
              Log the YAML
            </button>
          </span>
        </div>
        <div className="container mx-auto m-8">
          <HelmUI
            schema={schema.default}
            config={helmUIConfig.default}
            values={this.state.values}
            setValues={this.setValues}
          />
        </div>
      </div>
    )
  }
}

export default hot(module)(App)
