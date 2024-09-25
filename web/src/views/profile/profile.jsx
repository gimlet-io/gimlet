import React, { Component } from 'react';

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
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8 pt-24">
            <div className="px-4 py-8 sm:px-0">
              <div className="card">
                <div className="px-4 py-5 sm:px-6">
                  <h3 className="text-lg leading-6 font-medium">
                    Gimlet CLI access
                  </h3>
                  <p className="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
                    Follow the steps to set up CLI access
                  </p>
                </div>
                <div className="px-4 sm:p-6">
                  <h3 className="text-lg leading-6 font-medium">
                    Installation on Linux / Mac
                  </h3>
                  <code
                    className="block whitespace-pre overflow-x-scroll font-mono text-sm p-2 my-4 bg-neutral-800 dark:bg-neutral-900 text-yellow-100 rounded">
                    {`curl -L "https://github.com/gimlet-io/gimlet/releases/download/cli-v1.0.0-beta.1/gimlet-$(uname)-$(uname -m)" -o gimlet
chmod +x gimlet
sudo mv ./gimlet /usr/local/bin/gimlet
gimlet --version`}
                  </code>
                  <h3 className="pt-4 text-lg leading-6 font-medium">
                    Setting the API Key
                  </h3>
                  <code
                    className="block whitespace-pre overflow-x-scroll font-mono text-sm p-2 my-4 bg-neutral-800 dark:bg-neutral-900 text-yellow-100 rounded">
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
