import { Component } from "react";
import {
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWERROR
} from "../../redux/redux";

export class GhcrRegistryWidget extends Component {
  constructor(props) {
    super(props);

    this.state = {
      login: "",
      token: "",
      value: props.formData ?? "",
      sealed: props.formData ? true : false,
    };
  }

  componentDidUpdate(prevProps) {
    if (prevProps.formData !== this.props.formData) {
      this.setState({
        value: this.props.formData,
        sealed: true,
      });
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
    const { email, login, token } = this.state;
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
          this.props.onChange(data);
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

  render() {
    const { login, token, sealed, value } = this.state;

    return (
      <>
        <label class="control-label" for="root_login">Login</label>
        <input class="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
          value={login} onChange={e => this.setState({ login: e.target.value })} />
        <label class="control-label" for="root_token">Token</label>
        <input class="form-control" id="root_token" required="" label="Token" placeholder="" type="text" list="examples_root_token"
          value={token} onChange={e => this.setState({ token: e.target.value })} />
        <button className="my-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded h-12"
          onClick={this.seal()} >
          Seal
        </button>
        {
          sealed &&
          <div class="rounded-md bg-blue-50 p-4">
            <h3 class="text-sm font-medium text-blue-800">Sealed value:</h3>
            <div class="mt-2 text-sm text-blue-700">
              <div class="flex items-center">
                <span class="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded break-all">
                  {value}
                </span>
              </div>
            </div>
          </div>
        }
      </>
    );
  }
}

export class DockerhubRegistryWidget extends Component {
  constructor(props) {
    super(props);

    this.state = {
      email: "",
      login: "",
      token: "",
      value: props.formData ?? "",
      sealed: props.formData ? true : false,
    };
  }

  componentDidUpdate(prevProps) {
    if (prevProps.formData !== this.props.formData) {
      this.setState({
        value: this.props.formData,
        sealed: true,
      });
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
    const { email, login, token } = this.state;
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
          this.props.onChange(data);
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

  render() {
    const { email, login, token, sealed, value } = this.state;

    return (
      <>
        <label class="control-label" for="root_email">Email</label>
        <input class="form-control" id="root_email" required="" label="Email" placeholder="" type="text" list="examples_root_email"
          value={email} onChange={e => this.setState({ email: e.target.value })} />
        <label class="control-label" for="root_login">Login</label>
        <input class="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
          value={login} onChange={e => this.setState({ login: e.target.value })} />
        <label class="control-label" for="root_token">Token</label>
        <input class="form-control" id="root_token" required="" label="Token" placeholder="" type="text" list="examples_root_token"
          value={token} onChange={e => this.setState({ token: e.target.value })} />
        <button className="my-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded h-12"
          onClick={this.seal()} >
          Seal
        </button>
        {
          sealed &&
          <div class="rounded-md bg-blue-50 p-4">
            <h3 class="text-sm font-medium text-blue-800">Sealed value:</h3>
            <div class="mt-2 text-sm text-blue-700">
              <div class="flex items-center">
                <span class="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded break-all">
                  {value}
                </span>
              </div>
            </div>
          </div>
        }
      </>
    );
  }
}
