import { Component } from 'react';
import {
  ACTION_TYPE_ENVS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWERROR,
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
      scmUrl: reduxState.settings.scmUrl,
      settings: reduxState.settings,
    };
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({
        connectedAgents: reduxState.connectedAgents,
        envs: reduxState.envs,
        user: reduxState.user,
        popupWindow: reduxState.popupWindow,
        releaseStatuses: reduxState.releaseStatuses,
        scmUrl: reduxState.settings.scmUrl,
        settings: reduxState.settings
      });
    });
  }

  getEnvironmentCards() {
    const { connectedAgents, envs } = this.state;
    const sortedEnvs = this.sortingByName(envs);

    return (
      sortedEnvs.map(env => (<EnvironmentCard
        key={env.name}
        name={env.name}
        builtIn={env.builtIn}
        navigateToEnv={() => this.props.history.push(`/env/${env.name}`)}
        isOnline={this.isOnline(connectedAgents, env)}
      />))
    )
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
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-12">
            <h1 className="text-3xl font-bold leading-tight text-gray-900">Environments</h1>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            {this.state.settings.provider && this.state.settings.provider !== "" &&
              <div className="my-8 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
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
            }
            <div className="px-4 pt-12 sm:px-0">
              <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                {this.getEnvironmentCards()}
              </ul>
            </div>
          </div>
        </main>
      </div>
    )
  }
}

export default Environments;

function EnvironmentCard({ name, builtIn, navigateToEnv, isOnline }) {
  return (
    <li className="col-span-1 bg-white rounded-lg shadow divide-y divide-gray-200">
      <div className="w-full flex items-center justify-between p-6 space-x-6 cursor-pointer"
        onClick={() => navigateToEnv(name)}>
        <div className="flex-1 truncate">
          <div className="flex justify-between">
            <div className="flex">
              <p className="text-sm font-bold capitalize">{name}</p>
              {builtIn &&
                <span
                  className="flex-shrink-0 inline-block px-2 py-0.5 mx-1 text-gray-800 text-xs font-medium bg-gray-100 rounded-full">
                  built-in
                </span>
              }
            </div>
            <svg className={(isOnline ? "text-green-400" : "text-red-400") + " inline fill-current ml-1"} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 20 20">
              <path
                d="M0 14v1.498c0 .277.225.502.502.502h.997A.502.502 0 0 0 2 15.498V14c0-.959.801-2.273 2-2.779V9.116C1.684 9.652 0 11.97 0 14zm12.065-9.299l-2.53 1.898c-.347.26-.769.401-1.203.401H6.005C5.45 7 5 7.45 5 8.005v3.991C5 12.55 5.45 13 6.005 13h2.327c.434 0 .856.141 1.203.401l2.531 1.898a3.502 3.502 0 0 0 2.102.701H16V4h-1.832c-.758 0-1.496.246-2.103.701zM17 6v2h3V6h-3zm0 8h3v-2h-3v2z"
              />
            </svg>
          </div>
          <div className="p-2">
          </div>
        </div>
      </div>
    </li>
  )
}
