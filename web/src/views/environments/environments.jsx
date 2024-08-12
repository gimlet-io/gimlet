import { Component } from 'react';
import {
  ACTION_TYPE_ENVS,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWERROR,
  ACTION_TYPE_POPUPWINDOWRESET,
} from "../../redux/redux";
import EnvironmentCard from '../../components/environmentCard/environmentCard';
import { SkeletonLoader } from '../../../src/views/repositories/repositories';

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
      environmentsLoading: true,
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

  componentDidMount() {
    this.props.gimletClient.getEnvs()
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_ENVS, payload: data
        });
        this.setState({ environmentsLoading: false });
      }, () => {
        this.setState({ environmentsLoading: false });
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
    if (this.state.envs.some(env => env.name === this.state.input)) {
      this.props.store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
          header: "Error",
          message: "Environment already exists"
        }
      });
      return
    }

    this.setState({ isOpen: false });
    this.props.store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Saving..."
      }
    });
    this.setState({ saveButtonTriggered: true });
    this.props.gimletClient.saveEnvToDB(this.state.input)
      .then(() => {
        this.setState({
          envs: [...this.state.envs, {
            name: this.state.input,
            infraRepo: "",
            appsRepo: "",
            expiry: 0,
          }],
          input: "",
          saveButtonTriggered: false
        });
        this.refreshEnvs();
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWRESET
        });
      }, err => {
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
            header: "Error",
            message: err.statusText
          }
        });
      })
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
    const { connectedAgents, envs, settings } = this.state;
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
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 flex items-center">
            <h1 className="text-3xl leading-tight text-medium flex-grow">Environments</h1>
            <button type="button" className={`${!settings.trial ? 'primaryButton' : 'primaryButtonDisabled'} px-4`}
              onClick={() => !settings.trial && this.setState({ isOpen: true })}
              title={`${!settings.trial ? '' : 'Upgrade Gimlet to create environments'}`}
              >
              Create
            </button>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8 mt-8">
            <div className="px-4 sm:px-0">
              {this.state.isOpen &&
                <div className="card mb-4 p-4 flex space-x-2 items-center">
                  <input
                    onChange={e => this.setState({ input: e.target.value })}
                    className="input" id="environment" type="text" value={this.state.input} placeholder="Staging" />
                  <div className="p-0 flow-root space-x-1">
                    <span className="inline-flex rounded-md shadow-sm gap-x-1 float-right">
                      <button
                        disabled={this.state.input === "" || this.state.saveButtonTriggered}
                        onClick={() => this.save()}
                        className={(this.state.input === "" || this.state.saveButtonTriggered ? "primaryButtonDisabled" : "primaryButton") + " px-4"}>
                        Save
                      </button>
                      <button
                        disabled={this.state.input === "" || this.state.saveButtonTriggered}
                        onClick={() => this.setState({ isOpen: false })}
                        className='border-blue-500 dark:border-blue-700 text-blue-500 dark:text-blue-700 border hover:border-blue-400 dark:hover:border-blue-800 hover:text-blue-400 dark:hover:text-blue-800 cursor-pointer inline-flex items-center px-6 py-2 text-base leading-6 font-medium rounded-md transition ease-in-out duration-150'>
                        Cancel
                      </button>
                    </span>
                  </div>
                </div>
              }
              <ul className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                {
                  this.state.environmentsLoading ?
                    <SkeletonLoader />
                    :
                    <>
                      {
                        sortedEnvs.map(env => (<EnvironmentCard
                          key={env.name}
                          env={env}
                          navigateToEnv={() => this.props.history.push(`/env/${env.name}`)}
                          isOnline={this.isOnline(connectedAgents, env)}
                          trial={settings.trial}
                        />))
                      }
                    </>
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
