import { Component } from 'react';
import {
  ACTION_TYPE_ENVS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWRESET,
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

  setTimeOutForButtonTriggeredAndPopupWindow() {
    setTimeout(() => {
      this.props.store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
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
                {this.state.settings.provider && this.state.settings.provider !== "" &&
                  <li className="col-span-1 relative flex items-center justify-center w-full border-2 border-gray-300 border-dashed rounded-lg p-6 text-center hover:border-pink-400 cursor-pointer text-gray-500 hover:text-pink-500"
                    onClick={() => this.setState({ isOpen: true })}>
                    <div className="flex items-center justify-center">
                      <p className="text-sm font-bold py-2">Create new environment</p>
                    </div>
                  </li>
                }
              </ul>
              {this.state.isOpen &&
                <div className="bg-white p-4 sm:p-6 lg:p-8 space-y-4 px-4 py-5 sm:px-6 rounded-md shadow-md mt-8">
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
          </div>
        </main>
      </div>
    )
  }
}

export default Environments;
