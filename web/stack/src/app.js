import React, {Component} from 'react'
import {hot} from 'react-hot-loader'
import * as stackDefinitionFixture from '../fixtures/stack-definition.json'
import './style.css'
import StreamingBackend from './streamingBackend'
import GimletCLIClient from './client'

import {Category} from "./components/category";

class App extends Component {
  constructor(props) {
    super(props)

    const client = new GimletCLIClient()
    client.onError = (response) => {
      console.log(response)
      console.log(`${response.status}: ${response.statusText} on ${response.path}`)
    }

    this.state = {
      client: client,
      stack: {},
      stackNonDefaultValues: {},
      errors: {},
      showErrors: false
    }
    this.setValues = this.setValues.bind(this)
    this.validationCallback = this.validationCallback.bind(this)
  }

  componentDidMount() {
    fetch('/stack-definition.json')
      .then(response => {
        if (!response.ok && window !== undefined) {
          console.log("Using fixture")
          return stackDefinitionFixture.default
        }

        return response.json()
      })
      .then(data => this.setState({stackDefinition: data}))

    fetch('/stack.json')
      .then(response => {
        if (!response.ok && window !== undefined) {
          console.log("Using fixture")
          return {}
        }
        return response.json()
      })
      .then(data => this.setState({stack: data}))
  }

  setValues(variable, values, nonDefaultValues) {
    const updatedNonDefaultValues = {
      ...this.state.stackNonDefaultValues,
      [variable]: nonDefaultValues
    }

    this.setState(prevState => ({
      stack: {
        ...prevState.stack,
        [variable]: values
      },
      stackNonDefaultValues: {
        ...prevState.stackNonDefaultValues,
        [variable]: nonDefaultValues
      }
    }))
  }

  validationCallback(variable, errors) {
    if (errors === null) {
      this.setState(prevState => {
        delete prevState.errors[variable];

        if (JSON.stringify(prevState.errors) === "{}") {
          return {
            errors: {},
            showErrors: false
          }
        }

        return {errors: prevState.errors}
      })
      return
    }

    errors = errors.filter(error => error.keyword !== 'oneOf');
    errors = errors.filter(error => error.dataPath !== '.enabled');

    this.setState(prevState => ({
      errors: {
        ...prevState.errors,
        [variable]: errors
      }
    }))
  }

  render() {
    let {stackDefinition, stack} = this.state

    if (stackDefinition === undefined || stack === undefined) {
      return null;
    }

    const categories = stackDefinition.categories.map(category => {
      return <Category
        category={category}
        stackDefinition={stackDefinition}
        stack={stack}
        genericComponentSaver={this.setValues}
        genericValidationCallback={this.validationCallback}
      />
    })

    return (
      <div>
        <StreamingBackend client={this.state.client}/>
        <div
          className={this.state.showErrors ? 'block fixed bottom-0 right-0 mb-48 mr-8 bg-red-300 rounded-md shadow py-4 px-8' : 'hidden'}>
          <ul className="list-disc list-inside">
            {
              Object.keys(this.state.errors).map(variable => {
                return (
                  <div>
                    <p className='capitalize font-bold'>{variable}</p>
                    {this.state.errors[variable].map(e => {
                      return (
                        <li>{e.message}</li>
                      )
                    })}
                  </div>
                )
              })
            }
          </ul>
        </div>
        <div className="fixed bottom-0 right-0">
          <span className="inline-flex rounded-md shadow-sm m-8">
            <button
              type="button"
              className="inline-flex items-center px-12 py-6 border border-transparent text-base font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              onClick={() => {
                console.log(this.state.stack)
                console.log(this.state.stackNonDefaultValues)

                if (JSON.stringify(this.state.errors) !== "{}") {
                  this.setState(() => ({
                    showErrors: true
                  }))
                  return false
                }

                this.state.client.saveValues(this.state.stackNonDefaultValues)
                  .then(() => {
                    close()
                  });
              }}
            >
              Close tab & <br/>
              Write config
            </button>
          </span>
        </div>
        <div className="container mx-auto m-8 max-w-4xl">
          <h1 className="text-2xl font-bold my-16">{stackDefinition.name}
            <span className="font-normal text-lg block">{stackDefinition.description}</span>
          </h1>
          {categories}
        </div>
      </div>
    )
  }
};

export default hot(module)(App)
