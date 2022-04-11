import React, { Component } from 'react';
import { ACTION_TYPE_USERS } from '../../redux/redux';
import DefaultProfilePicture from './defaultProfilePicture.png';
import { InformationCircleIcon } from '@heroicons/react/solid';

export default class Profile extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      user: reduxState.user,
      gimletd: reduxState.gimletd,
      application: reduxState.application,
      users: reduxState.users,
      input: "",
      saveButtonTriggered: false,
      hasSameUsernameError: false,
      hasRequestError: false,
      latestUser: "",
      tokenOfLatestUser: ""
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ user: reduxState.user });
      this.setState({ users: reduxState.users });
      this.setState({ gimletd: reduxState.gimletd });
      this.setState({ application: reduxState.application });
    });
  }

  setTimeOutForButtonTriggered() {
    setTimeout(() => {
      this.setState({
        saveButtonTriggered: false,
        hasSameUsernameError: false
      })
    }, 3000);
  }

  save() {
    this.setState({ saveButtonTriggered: true });
    if (!this.state.users.some(user => user.login === this.state.input)) {
      this.props.gimletClient.saveUser(this.state.input)
        .then(saveUserResponse => {
          this.setState({
            users: [...this.state.users, {
              login: this.state.input,
              token: '',
              admin: false,
            }],
            saveButtonTriggered: false,
            input: "",
            tokenOfLatestUser: saveUserResponse.token,
            latestUser: saveUserResponse.login
          });
          this.refreshUsers()
        }, () => {
          this.setState({ hasRequestError: true });
          this.setTimeOutForButtonTriggered();
        })
    } else {
      this.setState({ hasSameUsernameError: true });
      this.setTimeOutForButtonTriggered();
    }
  }

  refreshUsers() {
    this.props.gimletClient.getUsers()
      .then(data => {
        this.props.store.dispatch({
          type: ACTION_TYPE_USERS,
          payload: data
        });
      }, () => {/* Generic error handler deals with it */
      });
  }

  sortAlphabetically(users) {
    const copiedUsers = users;
    return copiedUsers.sort((a, b) => a.login.localeCompare(b.login));
  }

  render() {
    const { user, users, gimletd } = this.state

    const sortedUsers = this.sortAlphabetically(users);

    const loggedIn = user !== undefined;
    if (!loggedIn) {
      return null;
    }

    user.imageUrl = `https://github.com/${user.login}.png?size=128`;

    const gimletdIntegrationEnabled = gimletd !== undefined;

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center space-x-5">
              <div className="flex-shrink-0">
                <div className="relative">
                  <img className="h-16 w-16 rounded-full"
                    src={user.imageUrl}
                    alt="" />
                  <span className="absolute inset-0 shadow-inner rounded-full" aria-hidden="true"></span>
                </div>
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{user.name}</h1>
                <p className="text-sm font-medium text-gray-500">{user.login}</p>
              </div>
            </div>
          </div>
        </header>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 py-8 sm:px-0">
              {githubAppSettings(this.state.application.name, this.state.application.appSettingsURL, this.state.application.installationURL)}
              {gimletdIntegrationEnabled &&
                <div className="bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200 my-4">
                  <div className="px-4 py-5 sm:px-6">
                    <h3 className="text-lg leading-6 font-medium text-gray-900">
                      Gimlet CLI access
                    </h3>
                    <p className="mt-1 text-sm text-gray-500">
                      Follow the steps to set up CLI access
                    </p>
                  </div>
                  <div className="px-4 py-5 sm:p-6">
                    <h3 className="text-lg leading-6 font-medium text-gray-900">
                      Installation on Linux / Mac
                    </h3>
                    <code
                      className="block whitespace-pre overflow-x-scroll font-mono text-sm p-2 my-4 bg-gray-800 text-yellow-100 rounded">
                      {`curl -L https://github.com/gimlet-io/gimlet-cli/releases/download/v0.9.6/gimlet-$(uname)-$(uname -m) -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version`}
                    </code>
                    <h3 className="text-lg leading-6 font-medium text-gray-900">
                      Setting the API Key
                    </h3>
                    <code
                      className="block whitespace-pre overflow-x-scroll font-mono text-sm p-2 my-4 bg-gray-800 text-yellow-100 rounded">
                      {`mkdir -p ~/.gimlet

cat << EOF > ~/.gimlet/config
export GIMLET_SERVER=${gimletd.url}
export GIMLET_TOKEN=${gimletd.user.token}
EOF
                      
source ~/.gimlet/config`}
                    </code>
                    <h3 className="text-lg leading-6 font-medium text-gray-900">
                      Getting the latest releases
                    </h3>
                    <code
                      className="block whitespace-pre overflow-x-scroll font-mono text-sm p-2 my-4 bg-gray-800 text-yellow-100 rounded">
                      {`gimlet release status --env staging`}
                    </code>
                  </div>
                </div>
              }
              {this.so}
              {users &&
                userList(sortedUsers, this.state.latestUser, this.state.tokenOfLatestUser, DefaultProfilePicture)
              }
              <div className="mt-12 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
                <div className="px-4 py-5 sm:px-6">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Create new user</h3>
                </div>
                <div className="px-4 py-5 sm:px-6">
                  <input
                    onChange={e => this.setState({ input: e.target.value })}
                    className="shadow appearance-none border rounded w-full my-4 py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="environment" type="text" value={this.state.input} placeholder="Please enter a username" />
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
          {
            !gimletdIntegrationEnabled &&
            <div className="max-w-7xl mx-auto sm:px-6 lg:px-8 my-8">
              {integrateGimletD()}
            </div>
          }
        </main >
      </div >
    )
  }
}

function integrateGimletD() {
  return (
    // eslint-disable-next-line
    <a
      href="https://gimlet.io/docs/configuring-gimletd/"
      target="_blank"
      className="relative block w-full border-2 border-gray-300
              border-dashed rounded-lg p-12 text-center hover:border-gray-400
              focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="mx-auto h-12 w-12 text-gray-400"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>

      <span className="mt-2 block text-lg leading-6 font-medium text-gray-900">Integrate GimletD</span>
    </a>
  )
}

function githubAppSettings(appName, appSettingsURL, installationURL) {
  return (
    <div className="bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Github Application
        </h3>
      </div>
      <div className="px-4 py-5 sm:px-6">
        <div className="inline-grid">
          <span
            onClick={() => window.open(appSettingsURL)}
            className="mt-1 text-sm text-gray-500 hover:text-gray-600 cursor-pointer">
            Settings for {appName}
          </span>
          <span>
            <a
              href={installationURL}
              rel="noreferrer"
              target="_blank"
              className="mt-1 text-sm text-gray-500 hover:text-gray-600">
              Application installation
            </a>
          </span>
        </div>
      </div>
    </div>
  )
}

function userList(sortedUsers, latestUser, tokenOfLatestUser, defaultProfilePicture) {
  return (
    <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Users
        </h3>
      </div>
      <div className="px-4 py-5 sm:px-6">
        {sortedUsers.map(user => (
          <div className="flex my-4 justify-between">
            <div className="inline-flex items-center">
              <img
                className="h-8 w-8 rounded-full text-2xl font-medium text-gray-900"
                src={`https://github.com/${user.login}.png?size=128`}
                onError={(e) => { e.target.src = defaultProfilePicture }}
                alt="" />
              <div className="ml-4">{user.login}</div>
            </div>
            {user.login === latestUser &&
              <div className="rounded-md bg-blue-50 p-4 w-5/6">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-blue-800">User token:</h3>
                    <div className="mt-2 text-sm text-blue-700">
                      <div className="flex items-center">
                        <span className="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded break-all">{tokenOfLatestUser}</span>
                        <div className="ml-3 cursor-pointer" onClick={() => { navigator.clipboard.writeText(tokenOfLatestUser) }}>
                          <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-blue-400 hover:text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                          </svg>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            }
          </div>
        ))}
      </div>
    </div>
  )
}
