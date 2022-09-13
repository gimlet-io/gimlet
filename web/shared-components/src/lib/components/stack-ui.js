import React, {Component} from 'react'
import './style.css'
import {Category} from "./category";

class StackUI extends Component {
  constructor(props) {
    super(props)

    this.state = {
      showErrors: false
    }
  }

  render() {
    let {stack, stackDefinition, setValues, validationCallback, categoriesToRender, componentsToRender, hideTitle} = this.props

    if (stackDefinition === undefined || stack === undefined) {
      return null;
    }

    const categories = stackDefinition.categories.map(category => {
      if (categoriesToRender) {
        const toRender = categoriesToRender.find(c => category.id === c);
        if (!toRender) {
          return null;
        }
      }

      return <Category
        category={category}
        stackDefinition={stackDefinition}
        stack={stack}
        genericComponentSaver={setValues}
        genericValidationCallback={validationCallback}
        componentsToRender={componentsToRender}
      />
    })

    return (
      <div>
        <div>
          <h1 className={hideTitle ? "hidden" : "text-2xl font-bold mb-4"}>{stackDefinition.name}
            <span className="font-normal text-lg block">{stackDefinition.description}</span>
          </h1>
          {categories}
        </div>
      </div>
    )
  }
};

export default StackUI;
