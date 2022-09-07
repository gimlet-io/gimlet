import './style.css'
import React, {Component} from 'react'
import {Tile} from "./tile";
import HelmUI from "helm-react-ui";
import {XIcon} from "@heroicons/react/outline";
import {Remarkable} from "remarkable";

export class Category extends Component {
  constructor(props) {
    super(props);

    this.state = {
      toggleState: {},
      tabState: {}
    }

    this.toggleComponent = this.toggleComponent.bind(this)
  }

  toggleComponent(category, component) {
    this.setState(prevState => ({
      toggleState: {
        ...prevState.toggleState,
        [category]: prevState.toggleState[category] == component ? undefined : component
      },
      tabState: {
        ...prevState.tabState,
        [component]: prevState.tabState[component] === undefined ? 'getting-started' : prevState.tabState[component]
      }
    }))
  }

  switchTab(component, tab) {
    this.setState(prevState => ({
      tabState: {
        ...prevState.tabState,
        [component]: tab
      }
    }))
  }

  render() {
    let {toggleState} = this.state

    const {
      category,
      stackDefinition,
      stack,
      genericComponentSaver,
      genericValidationCallback,
      componentsToRender
    } = this.props;

    let selectedComponent = undefined;
    let selectedComponentConfig = undefined;
    let componentSaver = undefined;
    let validationCallback = undefined;
    const selectedComponentName = toggleState[category.id];

    if (selectedComponentName !== undefined) {
      const selectedComponentArray = stackDefinition.components.filter(component => component.variable === toggleState[category.id]);
      selectedComponent = selectedComponentArray[0];
      selectedComponentConfig = stack[selectedComponent.variable];
      if (selectedComponentConfig === undefined) {
        selectedComponentConfig = {}
      }
      componentSaver = function (values, nonDefaultValues) {
        genericComponentSaver(selectedComponent.variable, values, nonDefaultValues)
      };
      validationCallback = function (errors) {
        genericValidationCallback(selectedComponent.variable, errors)
      };
    }

    const componentsForCategory = stackDefinition.components.filter(component => component.category === category.id);
    const componentTitles = componentsForCategory.map(component => {
      if (componentsToRender) {
        const toRender = componentsToRender.find(c => component.variable === c);
        if (!toRender) {
          return null;
        }
      }

      return (
        <Tile
          category={category}
          component={component}
          componentConfig={stack[component.variable]}
          selectedComponentName={selectedComponentName}
          toggleComponentHandler={this.toggleComponent}
        />
      )
    })

    if (selectedComponentName !== undefined){
      if (typeof selectedComponent.schema !== 'object') {
        selectedComponent.schema = JSON.parse(selectedComponent.schema)
      }

      if (typeof selectedComponent.uiSchema !== 'object') {
        selectedComponent.uiSchema = JSON.parse(selectedComponent.uiSchema)
      }
    }

    const componentConfigPanel = selectedComponentName === undefined ? null : (
      <div className="py-6 px-4 space-y-6 sm:p-6">
        <HelmUI
          schema={selectedComponent.schema}
          config={selectedComponent.uiSchema}
          values={selectedComponentConfig}
          setValues={componentSaver}
          validate={true}
          validationCallback={validationCallback}
        />
      </div>
    );

    const md = new Remarkable();
    const gettingStartedPanel = selectedComponentName === undefined ? null : (
      <div className="py-6 px-4 space-y-6 sm:p-6">
        <div className="prose" dangerouslySetInnerHTML={{__html: md.render(selectedComponent.onePager)}}/>
      </div>
    );

    const notSelectedTabStyle = "cursor-defaultborder-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm";
    const selectedTabStyle = "cursor-default border-indigo-500 text-indigo-600 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm";
    const tabState = this.state.tabState

    return (
      <div className="my-8">
        <h2 className="text-lg">{category.name}</h2>
        <div className="flex space-x-2 my-2">
          {componentTitles}
        </div>
        <div className="my-2">
          {selectedComponentName !== undefined &&
          <div className="px-8 py-4 shadow sm:rounded-md sm:overflow-hidden bg-white relative">
            <div className="hidden sm:block absolute top-0 right-0 pt-4 pr-4">
              <span
                className="bg-white rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                onClick={() => this.toggleComponent(category.id, selectedComponent.variable)}
              >
                <span className="sr-only">Close</span>
                <XIcon className="h-6 w-6" aria-hidden="true"/>
              </span>
            </div>
            <div>
              <div className="hidden sm:block">
                <div className="border-b border-gray-200">
                  <nav className="-mb-px flex space-x-8" aria-label="Tabs">
                    <span
                       className={tabState[selectedComponentName] == 'getting-started' ? selectedTabStyle : notSelectedTabStyle}
                       aria-current={tabState[selectedComponentName] == 'getting-started' ? 'page' : undefined}
                       onClick={() => this.switchTab(selectedComponentName, 'getting-started')}
                    >
                      Getting Started
                    </span>
                    <span
                       className={tabState[selectedComponentName] == 'config' ? selectedTabStyle : notSelectedTabStyle}
                       aria-current={tabState[selectedComponentName] == 'config' ? 'page' : undefined}
                       onClick={() => this.switchTab(selectedComponentName, 'config')}
                    >
                      Config
                    </span>
                  </nav>
                </div>
              </div>
            </div>
            {tabState[selectedComponentName] == 'getting-started' &&
              gettingStartedPanel
            }
            {tabState[selectedComponentName] == 'config' &&
              componentConfigPanel
            }
          </div>
          }
        </div>
      </div>
    )
  }
}
