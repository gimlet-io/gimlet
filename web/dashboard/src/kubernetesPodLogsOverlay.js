import { XIcon } from '@heroicons/react/solid'
import { Component } from 'react';
import { ACTION_TYPE_OVERLAYRESET } from './redux/redux';


export default class KubernetesPodLogsOverlay extends Component {
    constructor(props) {
        super(props);
        let reduxState = this.props.store.getState();

        this.state = {
            podLogs: reduxState.podLogs,
            overlay: reduxState.overlay,
        };
        this.props.store.subscribe(() => {
            let reduxState = this.props.store.getState();

            this.setState({
                podLogs: reduxState.podLogs,
                overlay: reduxState.overlay,
            });
        });
    }

    render() {
        return (
            <div className={(this.state.overlay.visible ? "visible" : "invisible") + " fixed inset-0 z-10 overflow-y-auto"}>
                <div className="bg-slate-600 bg-opacity-70 top-0 left-0 w-full h-full outline-none overflow-x-hidden overflow-y-auto show flex min-h-full items-end justify-center text-center sm:items-center p-4">
                    <div className="w-full transform overflow-hidden rounded-lg bg-white text-left shadow-xl">
                        <div className="flex p-4">
                            <div className="w-0 h-96 flex-1 justify-between bg-gray-800 rounded-md overflow-auto text-left p-4">
                                {this.state.podLogs?.split('\n').map(line => <p key={line} className='font-mono text-xs text-yellow-200'>{line}</p>)}
                            </div>
                            <div className="ml-4 flex-shrink-0 flex items-start">
                                <button
                                    className="rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none"
                                    onClick={() => {
                                        this.props.store.dispatch({
                                            type: ACTION_TYPE_OVERLAYRESET, payload: {}
                                        });
                                    }}
                                >
                                    <span className="sr-only">Close</span>
                                    <XIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}
