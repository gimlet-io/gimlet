import { Component } from 'react';
import {
  ACTION_TYPE_ENVS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWERROR,
} from "../../redux/redux";
import EnvironmentCard from '../../components/environmentCard/environmentCard';

class Environments extends Component {
  constructor(props) {
    super(props);
    let reduxState = this.props.store.getState();

    this.state = {
      connectedAgents: reduxState.connectedAgents,
      envs: reduxState.envs,
      input: "",
      saveButtonTriggered: false,
      settings: reduxState.settings,
    };
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        connectedAgents: reduxState.connectedAgents,
        envs: reduxState.envs,
        settings: reduxState.settings
      });
    });
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

  save() {
    this.setState({ isOpen: false });
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
    const { connectedAgents, envs } = this.state;
    if (!envs) {
      return null;
    }

    if (!connectedAgents) {
      return null;
    }

    const sortedEnvs = this.sortingByName(envs);

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-12">
            <h1 className="text-3xl font-bold leading-tight text-gray-900">Environments</h1>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            {this.state.settings.provider && this.state.settings.provider !== "" &&
              <div className="my-8 bg-white overflow-hidden rounded-lg divide-y divide-gray-200">
                <div className="bg-white p-4 sm:p-6 lg:p-8 space-y-4">
                  <div
                    target="_blank"
                    rel="noreferrer"
                    onClick={() => {
                      this.setState({ isOpen: true });
                    }}
                    className="relative block w-full border-2 border-gray-300 border-dashed rounded-lg p-6 text-center hover:border-pink-400 cursor-pointer text-gray-500 hover:text-pink-500"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="mx-auto w-12 h-12">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" />
                    </svg>

                    <div className="mt-2 block text-sm font-bold">
                      Create new environment
                    </div>
                  </div>
                </div>
                {this.state.isOpen &&
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
                }
              </div>
            }
            <div className="px-4 sm:px-0 mt-8">
              <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                {
                  sortedEnvs.map(env => (<EnvironmentCard
                    key={env.name}
                    name={env.name}
                    builtIn={env.builtIn}
                    navigateToEnv={() => this.props.history.push(`/env/${env.name}`)}
                    isOnline={this.isOnline(connectedAgents, env)}
                  />))
                }
              </ul>
            </div>
          </div>
        </main>
      </div>
    )
  }
}

export default Environments;
