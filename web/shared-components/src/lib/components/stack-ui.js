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
    let {stack, stackDefinition, setValues, validationCallback} = this.props

    if (stackDefinition === undefined || stack === undefined) {
      return null;
    }

    const categories = stackDefinition.categories.map(category => {
      return <Category
        category={category}
        stackDefinition={stackDefinition}
        stack={stack}
        genericComponentSaver={setValues}
        genericValidationCallback={validationCallback}
      />
    })

    return (
      <div>
        <div>
          <h1 class="text-2xl font-bold mb-4">{stackDefinition.name}
            <span class="font-normal text-lg block">{stackDefinition.description}</span>
          </h1>
          {categories}
        </div>
      </div>
    )
  }
};

export default StackUI;
