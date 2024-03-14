import { Component } from "react";

export class GhcrRegistryWidget extends Component {
  constructor(props) {
    super(props);

    this.state = {
      login: "",
      token: "",
    };
  }

  seal() {
    const { email, login, token } = this.state;
    const configjson = {
      "auths": {
        "ghcr.io": {
          "email": email,
          "auth": btoa(`${login}:${token}`)
        }
      }
    }
    const configjsonString = JSON.stringify(configjson)
    return () => {
      this.props.onChange("toSeal: " + configjsonString)
    };
  }

  render() {
    return (
      <>
        <div class="form-group field field-string">
          <label class="control-label" for="root_login">Login</label>
          <input class="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
            value={this.state.login} onChange={e => this.setState({ login: e.target.value })} />
        </div>
        <div class="form-group field field-string">
          <label class="control-label" for="root_token">Token</label>
          <input class="form-control" id="root_token" required="" label="Token" placeholder="" type="text" list="examples_root_token"
            value={this.state.token} onChange={e => this.setState({ token: e.target.value })} />
        </div>
        <button className="m-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded h-12"
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
      email: "",
      login: "",
      token: "",
    };
  }

  seal() {
    const { email, login, token } = this.state;
    const configjson = {
      "auths": {
        "https://index.docker.io/v1/": {
          "email": email,
          "auth": btoa(`${login}:${token}`)
        }
      }
    }
    const configjsonString = JSON.stringify(configjson)
    return () => {
      this.props.onChange("toSeal: " + configjsonString)
    };
  }

  render() {
    return (
      <>
        <div class="form-group field field-string">
          <label class="control-label" for="root_email">Email</label>
          <input class="form-control" id="root_email" required="" label="Email" placeholder="" type="text" list="examples_root_email"
            value={this.state.email} onChange={e => this.setState({ email: e.target.value })} />
        </div>
        <div class="form-group field field-string">
          <label class="control-label" for="root_login">Login</label>
          <input class="form-control" id="root_login" required="" label="Login" placeholder="" type="text" list="examples_root_login"
            value={this.state.login} onChange={e => this.setState({ login: e.target.value })} />
        </div>
        <div class="form-group field field-string">
          <label class="control-label" for="root_token">Token</label>
          <input class="form-control" id="root_token" required="" label="Token" placeholder="" type="text" list="examples_root_token"
            value={this.state.token} onChange={e => this.setState({ token: e.target.value })} />
        </div>
        <button className="m-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded h-12"
          onClick={this.seal()} >
          Seal
        </button>
      </>
    );
  }
}
