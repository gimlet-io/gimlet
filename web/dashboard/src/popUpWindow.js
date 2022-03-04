import { Component } from 'react';
import { CheckIcon, ExclamationCircleIcon } from '@heroicons/react/solid'

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
        return this.state.popupWindow.isError ? "bg-red-600" : "bg-green-600";
    }

    savingText() {
        if (this.state.popupWindow.isError) {
            return (
                <>
                    <ExclamationCircleIcon className="mr-1 h-5 w-5" />
                    <div>Something went wrong: {this.state.popupWindow.errorMessage}.</div>
                </>
            )
        } else {
            return (
                <>
                    <CheckIcon className="mr-1 h-5 w-5" />
                    <div>Config saved succesfully!</div>
                </>
            )
        }
    }

    render() {
        return (<div
            className={(this.state.popupWindow.visible ? "visible" : "invisible") + " fixed inset-0 flex px-4 py-6 pointer-events-none sm:p-6 w-full flex-col items-end space-y-4"}>
            <div
                className={this.backgroundColor() + " max-w-lg w-full text-gray-100 text-sm shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5 overflow-hidden"}>
                <div className="flex p-4">
                    {this.savingText()}
                </div>
            </div>
        </div>)
    }
}
