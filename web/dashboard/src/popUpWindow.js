import { Component } from 'react';
import { CheckIcon, XCircleIcon } from '@heroicons/react/solid'

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
                        <h3 className="text-sm font-medium text-gray-800">Saving...</h3>
                    </div>
                </div>
            )
        } else if (this.state.popupWindow.isError) {
            return (
                <div className="flex">
                    <div className="flex-shrink-0">
                        <XCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
                    </div>
                    <div className="ml-3">
                        <h3 className="text-sm font-medium text-red-800">There were error with your submission</h3>
                        <div className="mt-2 text-sm text-red-700">
                            {this.state.popupWindow.errorList ?
                                <ul className="list-disc pl-5 space-y-1">
                                    {this.renderErrorList()}
                                </ul> :
                                <p>{this.state.popupWindow.message}</p>
                            }
                        </div>
                    </div>
                </div>
            )
        } else {
            return (
                <div className="flex">
                    <div className="flex-shrink-0">
                        <CheckIcon className="h-5 w-5 text-green-400" aria-hidden="true" />
                    </div>
                    <div className="ml-3">
                        <h3 className="text-sm font-medium text-green-800">Operation success</h3>
                        <div className="mt-2 text-sm text-green-700">
                            <p>{this.state.popupWindow.message}</p>
                        </div>
                    </div>
                </div>
            )
        }
    }

    render() {
        return (
            <div
                className={(this.state.popupWindow.visible ? "visible" : "invisible") + " fixed z-50 inset-0 flex px-4 py-6 pointer-events-none sm:p-6 w-full flex-col items-end space-y-4"}>
                <div
                    className={this.backgroundColor() + " max-w-lg w-full text-gray-100 text-sm shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5"}>
                    <div className="flex p-4">
                        {this.savingText()}
                    </div>
                </div>
            </div>)
    }
}
