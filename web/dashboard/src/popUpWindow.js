import { Component } from 'react';
import { CheckIcon, XCircleIcon } from '@heroicons/react/solid'
import {
    ACTION_TYPE_POPUPWINDOWRESET
} from "./redux/redux";

export default class PopUpWindow extends Component {
    constructor(props) {
        super(props);
        let reduxState = this.props.store.getState();

        this.state = {
            popupWindow: reduxState.popupWindow
        };
        this.props.store.subscribe(() => {
            let reduxState = this.props.store.getState();

            this.setState({
                popupWindow: reduxState.popupWindow
            });
        });
    }

    inProgress() {
        return (
            <div className="flex">
                <div className="flex-shrink-0">
                    <svg className="animate-spin h-5 w-5 text-black" xmlns="http://www.w3.org/2000/svg" fill="none"
                        viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4}></circle>
                        <path className="opacity-75" fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                </div>
                <div className="ml-3">
                    <h3 className="text-sm font-medium text-gray-800">{this.state.popupWindow.header}</h3>
                </div>
            </div>)
    }

    error() {
        return (
            <div className="flex">
                <div className="flex-shrink-0">
                    <XCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
                </div>
                <div className="ml-3">
                    <h3 className="text-sm font-medium text-red-800">{this.state.popupWindow.header}</h3>
                    <div className="mt-2 text-sm text-red-700">
                        {this.state.popupWindow.errorList ?
                            <ul className="list-disc pl-5 space-y-1">
                                {this.renderErrorList()}
                            </ul> :
                            <p>{this.state.popupWindow.message}</p>
                        }
                    </div>
                </div>
            </div>)
    }

    success() {
        return (
            <div className="flex">
                <div className="flex-shrink-0">
                    <CheckIcon className="h-5 w-5 text-green-400" aria-hidden="true" />
                </div>
                <div className="ml-3">
                    <h3 className="text-sm font-medium text-green-800">{this.state.popupWindow.header}</h3>
                    <div className="mt-2 text-sm text-green-700">
                        <p>{this.state.popupWindow.message}</p>
                    </div>
                </div>
            </div>
        )
    }

    backgroundColor() {
        return !this.state.popupWindow.finished ? "bg-gray-200" : this.state.popupWindow.isError ? "bg-red-50" : "bg-green-50"
    }

    renderErrorList() {
        return Object.keys(this.state.popupWindow.errorList)
            .filter(variable => this.state.popupWindow.errorList[variable] !== null)
            .map(variable => this.state.popupWindow.errorList[variable]
                .map(error => (<li>{`${variable} ${error.message}`}</li>)))
    }

    savingText() {
        if (!this.state.popupWindow.finished) {
            return this.inProgress();
        } else if (this.state.popupWindow.isError) {
            return this.error();
        } else {
            return this.success();
        }
    }

    close() {
        this.props.store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWRESET
        });
    }

    render() {
        return (
            <div
                className={(this.state.popupWindow.visible ? "visible" : "invisible") + " fixed z-50 inset-0 flex px-4 py-6 pointer-events-none sm:p-6 w-full flex-col items-end space-y-4"}>
                <div
                    className={this.backgroundColor() + " max-w-lg w-full text-gray-100 text-sm shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5"}>
                    <div className="flex p-4 justify-between">
                        {this.savingText()}
                        <div className="ml-4 flex-shrink-0 flex items-start">
                            <button className="rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none"
                                onClick={() => { this.close() }} >
                                <span className="sr-only">Close</span>
                                <svg className="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>)
    }
}
