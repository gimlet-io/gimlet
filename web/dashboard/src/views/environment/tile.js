import './style.css'
import React, { Component } from 'react'

export class Tile extends Component {
    render() {
        const { category, component, componentConfig, selectedComponentName, toggleComponentHandler } = this.props;

        const selected = component.variable === selectedComponentName;
        const enabled = componentConfig !== undefined ? componentConfig.enabled : false;
        return (
            <div onClick={() => toggleComponentHandler(category.id, component.variable)}>
                <div className="w-32 h-32 overflow-hidden cursor-pointer">
                    <div className={enabled ? !selected ? 'bg-green-100 hover:bg-green-300' : 'bg-green-300' : !selected ? 'bg-gray-200 hover:bg-gray-300 filter grayscale hover:grayscale-0' : 'bg-gray-300'}>
                        <img className="h-20 mx-auto pt-4" src={component.logo} alt={component.name} />
                        <div className="font-bold text-sm py-2 text-center">{component.name}</div>
                    </div>
                </div>
            </div>
        )
    }
}
