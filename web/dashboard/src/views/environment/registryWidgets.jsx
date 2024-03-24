import { Component } from "react";
import {
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWERROR
} from "../../redux/redux";

export class GhcrRegistryWidget extends Component {
  constructor(props) {
    super(props);

    this.state = {
      enabled: props.formData.enabled,
      login: props.formData.login ?? "",
      token: "",
      sealed: props.formData.encryptedDockerconfigjson !== "" ? true : false,
    };
  }

  componentDidUpdate(prevProps, prevState) {
    if (prevProps.formData.encryptedDockerconfigjson !== this.props.formData.encryptedDockerconfigjson) {
      this.setState({
        sealed: this.props.formData.encryptedDockerconfigjson !== "" ? true : false
      });
    }

    if (prevState.enabled !== this.state.enabled) {
      this.props.onChange({ "enabled": this.state.enabled, "encryptedDockerconfigjson": this.props.formData.encryptedDockerconfigjson, "login": this.props.formData.login })
    }
  }

  resetPopupWindowAfterThreeSeconds() {
    const { store } = this.props;
    setTimeout(() => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  };

  seal() {
    const { enabled, email, login, token } = this.state;
    const { gimletClient, store, env } = this.props;

    const configjson = {
      "auths": {
        "ghcr.io": {
          "email": email,
          "auth": btoa(`${login}:${token}`)
        }
      }
    }

    return () => {
      gimletClient.seal(env, JSON.stringify(configjson))
        .then(data => {
          this.props.onChange({ "enabled": enabled, "encryptedDockerconfigjson": data, "login": login })
        }, () => {
          store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
              header: "Error",
              message: "Failed to seal."
            }
          });
          this.resetPopupWindowAfterThreeSeconds()
        });
    };
  }

  reset() {
    return () => {
      this.props.onChange({ "enabled": this.state.enabled, "encryptedDockerconfigjson": "", "login": this.state.login })
    };
  }

  render() {
    const { login, token, sealed } = this.state;
    const disabled = login === "" || token === "";

    if (sealed) {
      return (
        <>
          <span
            role="checkbox"
            tabindex="0"
            aria-checked={this.state.enabled}
            className={(this.state.enabled ? "bg-indigo-600" : "bg-gray-200") + " mt-1 relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:shadow-outline"}
            onClick={() => this.setState(prevState => ({
              enabled: !prevState.enabled
            }))}
          >
            <span aria-hidden="true" className={(this.state.enabled ? "translate-x-5" : "translate-x-0") + " translate-x-0 inline-block h-5 w-5 rounded-full bg-white shadow transform transition ease-in-out duration-200"}></span>
          </span>
          <ConfiguredPanel />
          <button className="my-2 bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded h-12"
            onClick={this.reset()} >
            Reset
          </button>
        </>
      )
    }

    return (
      <>
        <span
          role="checkbox"
          tabindex="0"
          aria-checked={this.state.enabled}
          className={(this.state.enabled ? "bg-indigo-600" : "bg-gray-200") + " mt-1 relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:shadow-outline"}
          onClick={() => this.setState(prevState => ({
            enabled: !prevState.enabled
          }))}
        >
          <span aria-hidden="true" className={(this.state.enabled ? "translate-x-5" : "translate-x-0") + " translate-x-0 inline-block h-5 w-5 rounded-full bg-white shadow transform transition ease-in-out duration-200"}></span>
        </span>
        <label class="control-label" for="root_login">Login</label>
        <input class="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
          value={login} onChange={e => this.setState({ login: e.target.value })} />
        <label class="control-label" for="root_token">Token</label>
        <input class="form-control" id="root_token" required="" label="Token" placeholder="" type="text" list="examples_root_token"
          value={token} onChange={e => this.setState({ token: e.target.value })} />
        <button disabled={disabled} className={(disabled ? "bg-gray-500" : "bg-blue-500 hover:bg-blue-700") + " my-2 text-white font-bold py-2 px-4 rounded h-12"}
          onClick={this.seal()} >
          Seal
        </button>
      </>
    );
  }
}

export class DockerhubRegistryWidget extends Component {
  constructor(props) {
    super(props);

    this.state = {
      enabled: props.formData.enabled,
      email: "",
      login: props.formData.login ?? "",
      token: "",
      sealed: props.formData.encryptedDockerconfigjson ? true : false,
    };
  }

  componentDidUpdate(prevProps, prevState) {
    if (prevProps.formData !== this.props.formData) {
      this.setState({
        sealed: this.props.formData.encryptedDockerconfigjson !== "" ? true : false
      });
    }

    if (prevState.enabled !== this.state.enabled) {
      this.props.onChange({ "enabled": this.state.enabled, "encryptedDockerconfigjson": this.props.formData.encryptedDockerconfigjson, "login": this.props.formData.login })
    }
  }

  resetPopupWindowAfterThreeSeconds() {
    const { store } = this.props;
    setTimeout(() => {
      store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  };

  seal() {
    const { enabled, email, login, token } = this.state;
    const { gimletClient, store, env } = this.props;

    const configjson = {
      "auths": {
        "https://index.docker.io/v1/": {
          "email": email,
          "auth": btoa(`${login}:${token}`)
        }
      }
    }

    return () => {
      gimletClient.seal(env, JSON.stringify(configjson))
        .then(data => {
          this.props.onChange({ "enabled": enabled, "encryptedDockerconfigjson": data, "login": login })
        }, () => {
          store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
              header: "Error",
              message: "Failed to seal."
            }
          });
          this.resetPopupWindowAfterThreeSeconds()
        });
    };
  }

  reset() {
    return () => {
      this.props.onChange({ "enabled": this.state.enabled, "encryptedDockerconfigjson": "", "login": "" })
    };
  }

  render() {
    const { email, login, token, sealed } = this.state;
    const disabled = email === "" || login === "" || token === "";

    if (sealed) {
      return (
        <>
          <span
            role="checkbox"
            tabindex="0"
            aria-checked={this.state.enabled}
            className={(this.state.enabled ? "bg-indigo-600" : "bg-gray-200") + " mt-1 relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:shadow-outline"}
            onClick={() => this.setState(prevState => ({
              enabled: !prevState.enabled
            }))}
          >
            <span aria-hidden="true" className={(this.state.enabled ? "translate-x-5" : "translate-x-0") + " translate-x-0 inline-block h-5 w-5 rounded-full bg-white shadow transform transition ease-in-out duration-200"}></span>
          </span>
          <ConfiguredPanel />
          <button className="my-2 bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded h-12"
            onClick={this.reset()} >
            Reset
          </button>
        </>
      )
    }

    return (
      <>
        <span
          role="checkbox"
          tabindex="0"
          aria-checked={this.state.enabled}
          className={(this.state.enabled ? "bg-indigo-600" : "bg-gray-200") + " mt-1 relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:shadow-outline"}
          onClick={() => this.setState(prevState => ({
            enabled: !prevState.enabled
          }))}
        >
          <span aria-hidden="true" className={(this.state.enabled ? "translate-x-5" : "translate-x-0") + " translate-x-0 inline-block h-5 w-5 rounded-full bg-white shadow transform transition ease-in-out duration-200"}></span>
        </span>
        <label class="control-label" for="root_email">Email</label>
        <input class="form-control" id="root_email" required="" label="Email" placeholder="" type="text" list="examples_root_email"
          value={email} onChange={e => this.setState({ email: e.target.value })} />
        <label class="control-label" for="root_login">Login</label>
        <input class="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
          value={login} onChange={e => this.setState({ login: e.target.value })} />
        <label class="control-label" for="root_token">Token</label>
        <input class="form-control" id="root_token" required="" label="Token" placeholder="" type="text" list="examples_root_token"
          value={token} onChange={e => this.setState({ token: e.target.value })} />
        <button disabled={disabled} className={(disabled ? "bg-gray-500" : "bg-blue-500 hover:bg-blue-700") + " my-2 text-white font-bold py-2 px-4 rounded h-12"}
          onClick={this.seal()} >
          Seal
        </button>
      </>
    );
  }
}

const ConfiguredPanel = () => {
  return (
    <div className="rounded-md bg-green-50 p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd"></path>
          </svg>
        </div>
        <div className="ml-3">
          <h3 className="text-sm font-medium text-green-800">Configured</h3>
        </div>
      </div>
    </div>
  )
}
