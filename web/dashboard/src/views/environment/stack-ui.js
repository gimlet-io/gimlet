import React, { Component } from 'react'
import './style.css'
import { Category } from './category';

class StackUI extends Component {
  constructor(props) {
    super(props)

    this.state = {
      showErrors: false,
      selected: this.props.stackDefinition.categories[0]
    }
  }

  select(name) {
    this.setState({
      selected: name
    })
  }

  render() {
    let { stack, stackDefinition, setValues, validationCallback, categoriesToRender, componentsToRender, hideTitle } = this.props

    if (stackDefinition === undefined || stack === undefined) {
      return null;
    }

    const { categories, name, description } = stackDefinition;
    const sidebar = categories.length > 1
    let gridClass = 'lg:grid lg:grid-cols-12 lg:gap-x-16 pt-8';
    let gridSpan = 'lg:col-span-10'
    if (!sidebar) {
      gridClass = '';
      gridSpan = '';
    }

    return (
      <div>
        <h1 className={hideTitle ? "hidden" : "text-2xl font-bold mb-4"}>{name}
          <span className="font-normal text-lg block">{description}</span>
        </h1>
        <div className={gridClass}>
          {sidebar &&
            <aside className="py-6 px-2 sm:px-6 lg:py-0 lg:px-0 lg:col-span-2">
              <nav className="flex flex-1 flex-col" aria-label="Sidebar">
                <ul className="-mx-2 space-y-1">
                  {categories.map((category) => {
                    if (categoriesToRender) {
                      const toRender = categoriesToRender.find(c => category.id === c);
                      if (!toRender) {
                        return null;
                      }
                    }

                    const selected = this.state.selected.id === category.id

                    return (
                      <li key={category.name}>
                        {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                        <a
                          className={
                            (selected ? 'bg-gray-50 text-indigo-600' : 'text-gray-700 hover:text-indigo-600 hover:bg-gray-50') +
                            ' group flex gap-x-3 rounded-md p-2 pl-3 text-sm leading-6 font-semibold cursor-pointer'
                          }
                          aria-current="page"
                          onClick={() => this.select(category)}
                        >
                          {category.name}
                        </a>
                      </li>
                    )
                  })}
                </ul>
              </nav>
            </aside>
          }
          <div className={gridSpan}>
            <Category
              category={this.state.selected}
              stackDefinition={stackDefinition}
              stack={stack}
              genericComponentSaver={setValues}
              genericValidationCallback={validationCallback}
              componentsToRender={componentsToRender}
            />
          </div>
        </div>
      </div>
    )
  }
};

export default StackUI;
