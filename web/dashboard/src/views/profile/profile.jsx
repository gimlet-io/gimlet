import React, { Component } from 'react';
import DefaultProfilePicture from './defaultProfilePicture.png';

export default class Profile extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      user: reduxState.user,
      settings: reduxState.settings
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ user: reduxState.user });
      this.setState({ settings: reduxState.settings });
    });
  }

  render() {
    const { user, settings } = this.state

    const loggedIn = user !== undefined;
    if (!loggedIn) {
      return null;
    }

    user.imageUrl = `${settings.scmUrl}/${user.login}.png?size=128`;

    return (
      <div>
        <header>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center space-x-5">
              <div className="flex-shrink-0">
                <div className="relative">
                  <img className="h-16 w-16 rounded-full"
                    src={user.imageUrl}
                    alt={user.login}
                    onError={(e) => { e.target.src = DefaultProfilePicture }} />
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
                    {`curl -L "https://github.com/gimlet-io/gimlet/releases/download/cli-v0.26.0/gimlet-$(uname)-$(uname -m)" -o gimlet
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
export GIMLET_SERVER=${settings.host}
export GIMLET_TOKEN=${user.token}
EOF

source ~/.gimlet/config`}
                  </code>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    )
  }
}
