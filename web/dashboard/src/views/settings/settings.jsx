import React, { Component } from 'react';
import Installer from './installer';

export default class Settings extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      application: reduxState.application,
      settings: reduxState.settings
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ application: reduxState.application });
      this.setState({ settings: reduxState.settings });
    });
  }

  render() {
    const { settings, application } = this.state;

    return (
      <div>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 sm:px-0">
              {settings.scmUrl === "https://github.com" &&
                githubAppSettings(application.name, application.appSettingsURL, application.installationURL)}
              {settings.provider === "" &&
                gimletInstaller()}
            </div>
          </div>
        </main >
      </div >
    )
  }
}

function githubAppSettings(appName, appSettingsURL, installationURL) {
  return (
    <div className="bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200 my-4">
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

function gimletInstaller() {
  return (
    <div className="bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200 my-4">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Gimlet Installer
        </h3>
        <p className="mt-1 text-sm text-gray-500">
          Integrate with Source Code Manager
        </p>
      </div>
      <div className="px-4 py-5 sm:p-6">
        <Installer />
      </div>
    </div>
  )
}
