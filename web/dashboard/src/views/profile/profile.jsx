import React, { Component } from 'react';

export default class Profile extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      user: reduxState.user,
      gimletd: reduxState.gimletd,
      application: reduxState.application
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ user: reduxState.user });
      this.setState({ gimletd: reduxState.gimletd });
      this.setState({ application: reduxState.application });
    });
  }

  render() {
    const { user, gimletd } = this.state;

    const loggedIn = user !== undefined;
    if (!loggedIn) {
      return null;
    }

    user.imageUrl = `https://github.com/${user.login}.png?size=128`

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
                <div className="bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
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
            </div>
          </div>
          {!gimletdIntegrationEnabled &&
            <div className="max-w-7xl mx-auto sm:px-6 lg:px-8 my-8">
              {integrateGimletD()}
            </div>
          }
        </main>
      </div>
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
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
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
