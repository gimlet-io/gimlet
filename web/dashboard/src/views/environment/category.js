import './style.css'
import React, { Component } from 'react'
import { Tile } from './tile';
import HelmUI from "helm-react-ui";
import { InformationCircleIcon } from '@heroicons/react/solid'
import { Remarkable } from "remarkable";

export class Category extends Component {
  constructor(props) {
    super(props);

    this.state = {
      toggleState: {},
    }

    this.toggleComponent = this.toggleComponent.bind(this)
  }

  toggleComponent(category, component) {
    this.setState(prevState => ({
      toggleState: {
        ...prevState.toggleState,
        [category]: prevState.toggleState[category] === component ? undefined : component
      },
    }))
  }

  render() {
    let { toggleState } = this.state

    const {
      category,
      stackDefinition,
      stack,
      genericComponentSaver,
      genericValidationCallback,
      componentsToRender,
      customFields
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
          key={component.variable}
          category={category}
          component={component}
          componentConfig={stack[component.variable]}
          selectedComponentName={selectedComponentName}
          toggleComponentHandler={this.toggleComponent}
        />
      )
    })

    if (selectedComponentName !== undefined) {
      if (typeof selectedComponent.schema !== 'object') {
        selectedComponent.schema = JSON.parse(selectedComponent.schema)
      }

      if (typeof selectedComponent.uiSchema !== 'object') {
        selectedComponent.uiSchema = JSON.parse(selectedComponent.uiSchema)
      }
    }

    const componentConfigPanel = selectedComponentName === undefined ? null : (
      <HelmUI
        schema={selectedComponent.schema}
        config={selectedComponent.uiSchema}
        fields={customFields}
        values={selectedComponentConfig}
        setValues={componentSaver}
        validate={true}
        validationCallback={validationCallback}
      />
    );

    const md = new Remarkable();
    const gettingStartedPanel = selectedComponentName === undefined ? null : (
      <div className="prose max-w-lg" dangerouslySetInnerHTML={{ __html: md.render(selectedComponent.onePager) }} />
    );
    const emptyOnePager = selectedComponentName === undefined ? null : selectedComponent.onePager === "";

    return (
      <div>
        <div className="flex space-x-2 my-2">
          {componentTitles}
        </div>
        {selectedComponentName !== undefined &&
          <div className='flex my-2'>
            <div className="p-4 max-w-lg min-w-[500px] shadow sm:rounded-md sm:overflow-hidden bg-white relative">
              <div className="col-span-6">
                {componentConfigPanel}
              </div>
            </div>
            {!emptyOnePager &&
              <div className='overflow-visible'>
                <div className="py-6 pl-10 sm:p-6 rounded-md bg-blue-50">
                  <h3 className="text-sm font-medium text-blue-800">
                    <InformationCircleIcon className="h-5 w-5 text-blue-400 inline" aria-hidden="true" />
                    <span className='pl-1'>Getting started</span>
                    </h3>
                  <div className="mt-2 text-sm text-blue-700">
                    {gettingStartedPanel}
                  </div>
                </div>
              </div>
            }
          </div>
        }
      </div>
    )
  }
}
