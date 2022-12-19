import { Component } from 'react';
import EnvironmentCard from './environmentCard.jsx';
import {
    ACTION_TYPE_ENVS,
    ACTION_TYPE_POPUPWINDOWRESET,
    ACTION_TYPE_POPUPWINDOWPROGRESS,
    ACTION_TYPE_POPUPWINDOWSUCCESS,
    ACTION_TYPE_POPUPWINDOWERROR,
    ACTION_TYPE_ENV_PULLREQUESTS
} from "../../redux/redux";

class Environments extends Component {
    constructor(props) {
        super(props);
        let reduxState = this.props.store.getState();

        this.state = {
            connectedAgents: reduxState.connectedAgents,
            envs: reduxState.envs,
            input: "",
            saveButtonTriggered: false,
            user: reduxState.user,
            popupWindow: reduxState.popupWindow,
            releaseStatuses: reduxState.releaseStatuses,
            scmUrl: reduxState.settings.scmUrl
        };
        this.props.store.subscribe(() => {
            let reduxState = this.props.store.getState();

            this.setState({
                connectedAgents: reduxState.connectedAgents,
                envs: reduxState.envs,
                user: reduxState.user,
                popupWindow: reduxState.popupWindow,
                releaseStatuses: reduxState.releaseStatuses,
                scmUrl: reduxState.settings.scmUrl
            });
        });

        this.refreshEnvs = this.refreshEnvs.bind(this)
    }

    refreshEnvs() {
        this.props.gimletClient.getEnvs()
            .then(data => {
                this.props.store.dispatch({
                    type: ACTION_TYPE_ENVS,
                    payload: data
                });
            }, () => {/* Generic error handler deals with it */
            });
    }

    componentDidMount() {
        this.props.gimletClient.getPullRequestsFromInfraRepo()
            .then(data => {
                this.props.store.dispatch({
                    type: ACTION_TYPE_ENV_PULLREQUESTS,
                    payload: data
                  });
            }, () => {/* Generic error handler deals with it */
            });
    }

    getEnvironmentCards() {
        console.log("generating env cards")
        const { connectedAgents, envs, releaseStatuses, popupWindow, scmUrl } = this.state;
        const { environment, tab } = this.props.match.params;
        const sortedEnvs = this.sortingByName(envs);

        return (
            sortedEnvs.map(env => (<EnvironmentCard
                key={env.name}
                store={this.props.store}
                env={env}
                deleteEnv={() => this.delete(env.name)}
                isOnline={this.isOnline(connectedAgents, env)}
                gimletClient={this.props.gimletClient}
                refreshEnvs={this.refreshEnvs}
                tab={tab}
                envFromParams={environment}
                releaseStatuses={releaseStatuses[env.name]}
                popupWindow={popupWindow}
                pullRequests={env.pullRequests}
                scmUrl={scmUrl}
            />))
        )
    }

    sortingByName(envs) {
        const envsCopy = [...envs]
        return envsCopy.sort((a, b) => a.name.localeCompare(b.name));
    }

    isOnline(onlineEnvs, singleEnv) {
        return Object.keys(onlineEnvs)
            .map(env => onlineEnvs[env])
            .some(onlineEnv => {
                return onlineEnv.name === singleEnv.name
            })
    };

    setTimeOutForButtonTriggeredAndPopupWindow() {
        setTimeout(() => {
            this.setState({ saveButtonTriggered: false })
            this.props.store.dispatch({
                type: ACTION_TYPE_POPUPWINDOWRESET
            });
        }, 3000);
    }

    save() {
        this.props.store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
                header: "Saving..."
            }
        });
        
        this.setState({ saveButtonTriggered: true });
        if (!this.state.envs.some(env => env.name === this.state.input)) {
            this.props.gimletClient.saveEnvToDB(this.state.input)
                .then(() => {
                    this.props.store.dispatch({
                        type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
                            header: "Success",
                            message: "Environment saved"
                        }
                    });
                    this.setState({
                        envs: [...this.state.envs, {
                            name: this.state.input,
                            infraRepo: "",
                            appsRepo: ""
                        }],
                        input: "",
                        saveButtonTriggered: false
                    });
                    this.refreshEnvs();
                    this.setTimeOutForButtonTriggeredAndPopupWindow();
                }, err => {
                    this.props.store.dispatch({
                        type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
                            header: "Error",
                            message: err.statusText
                        }
                    });
                    this.setTimeOutForButtonTriggeredAndPopupWindow();
                })
        } else {
            this.props.store.dispatch({
                type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
                    header: "Error",
                    message: "Environment already exists"
                }
            });
            this.setTimeOutForButtonTriggeredAndPopupWindow();
        }
    }

    delete(envName) {
        this.props.store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
                header: "Deleting..."
            }
        });

        this.props.gimletClient.deleteEnvFromDB(envName)
            .then(() => {
                this.props.store.dispatch({
                    type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
                        header: "Success",
                        message: "Environment deleted"
                    }
                });
                this.setState({ envs: this.state.envs.filter(env => env.name !== envName) });
                this.refreshEnvs();
                this.setTimeOutForButtonTriggeredAndPopupWindow();
            }, err => {
                this.props.store.dispatch({
                    type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
                        header: "Error",
                        message: err.statusText
                    }
                });
                this.setTimeOutForButtonTriggeredAndPopupWindow();
            });
    }

    render() {
        if (!this.state.envs) {
            return null;
        }

        if (!this.state.connectedAgents) {
            return null;
        }

        return (
            <>
                <header>
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <h1 className="text-3xl font-bold leading-tight text-gray-900">Environments</h1>
                    </div>
                </header>
                <main>
                    <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
                        <div className="px-4 py-8 sm:px-0">
                            {this.getEnvironmentCards()}
                            <div className="mt-12 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
                                <div className="px-4 py-5 sm:px-6">
                                    <h3 className="text-lg leading-6 font-medium text-gray-900">Create new environment</h3>
                                </div>
                                <div className="px-4 py-5 sm:px-6">
                                    <input
                                        onChange={e => this.setState({ input: e.target.value })}
                                        className="shadow appearance-none border rounded w-full my-4 py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="environment" type="text" value={this.state.input} placeholder="Please enter an environment name" />
                                    <div className="p-0 flow-root">
                                        <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
                                            <button
                                                disabled={this.state.input === "" || this.state.saveButtonTriggered}
                                                onClick={() => this.save()}
                                                className={(this.state.input === "" || this.state.saveButtonTriggered ? "bg-gray-600 cursor-not-allowed" : "bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700") + " inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150"}>
                                                Create
                                            </button>
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </main>
            </>
        )
    }
}

export default Environments;
