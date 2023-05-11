import React, { Component } from 'react';
import Installer from './installer';
import { InformationCircleIcon } from '@heroicons/react/solid';
import DefaultProfilePicture from '../profile/defaultProfilePicture.png';
import {
  ACTION_TYPE_USERS,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWERROR
} from '../../redux/redux';

export default class Settings extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      application: reduxState.application,
      settings: reduxState.settings,
      users: reduxState.users,
      input: "",
      saveButtonTriggered: false,
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ application: reduxState.application });
      this.setState({ settings: reduxState.settings });
      this.setState({ users: reduxState.users });
    });
  }

  save() {
    this.props.store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Saving..."
      }
    });

    this.setState({ saveButtonTriggered: true });
    if (!this.state.users.some(user => user.login === this.state.input)) {
      this.props.gimletClient.saveUser(this.state.input)
        .then(saveUserResponse => {
          this.setTimeOutForButtonTriggeredAndPopupWindow();
          this.setState({ input: "" });
          this.props.store.dispatch({
            type: ACTION_TYPE_USERS,
            payload: [...this.state.users, {
              login: saveUserResponse.login,
              token: saveUserResponse.token,
              admin: false
            }]
          });
          this.props.store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
              header: "Success",
              message: "User saved"
            }
          });
        }, err => {
          this.setTimeOutForButtonTriggeredAndPopupWindow();
          this.props.store.dispatch({
            type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
              header: "Error",
              message: err.statusText
            }
          });
        })
    } else {
      this.setTimeOutForButtonTriggeredAndPopupWindow();
      this.props.store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWERROR, payload: {
          header: "Error",
          message: "User already exists"
        }
      });
    }
  }

  setTimeOutForButtonTriggeredAndPopupWindow() {
    setTimeout(() => {
      this.setState({ saveButtonTriggered: false })
      this.props.store.dispatch({
        type: ACTION_TYPE_POPUPWINDOWRESET
      });
    }, 3000);
  }

  sortAlphabetically(users) {
    return users.sort((a, b) => a.login.localeCompare(b.login));
  }

  render() {
    const { settings, application, users, input, saveButtonTriggered } = this.state;
    const sortedUsers = this.sortAlphabetically(users);

    return (
      <div>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 sm:px-0">
              {settings.scmUrl === "https://github.com" &&
                githubAppSettings(application.name, application.appSettingsURL, application.installationURL)}
              {(!settings.provider || settings.provider === "") &&
                gimletInstaller()}
              {users &&
                userList(sortedUsers, DefaultProfilePicture, settings.scmUrl)
              }
              <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
                <div className="px-4 py-5 sm:px-6">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Create new user</h3>
                </div>
                <div className="px-4 py-5 sm:px-6">
                  <input
                    onChange={e => this.setState({ input: e.target.value })}
                    className="shadow appearance-none border rounded w-full my-4 py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                    id="environment"
                    type="text"
                    value={input}
                    placeholder="Please enter a username" />
                  <div className="p-0 flow-root">
                    <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
                      <button
                        disabled={input === "" || saveButtonTriggered}
                        onClick={() => this.save()}
                        className={(input === "" || saveButtonTriggered ? "bg-gray-600 cursor-not-allowed" : "bg-green-600 hover:bg-green-500 focus:outline-none focus:border-green-700 focus:shadow-outline-indigo active:bg-green-700") + " inline-flex items-center px-6 py-3 border border-transparent text-base leading-6 font-medium rounded-md text-white transition ease-in-out duration-150"}>
                        Create
                      </button>
                    </span>
                  </div>
                </div>
              </div>
              {dashboardVersion(application)}
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

function userList(sortedUsers, defaultProfilePicture, scmUrl) {
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
                src={`${scmUrl}/${user.login}.png?size=128`}
                onError={(e) => { e.target.src = defaultProfilePicture }}
                alt={user.login} />
              <div className="ml-4">{user.login}</div>
            </div>
            {user.token &&
              <div className="rounded-md bg-blue-50 p-4 w-5/6">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-blue-800">User token:</h3>
                    <div className="mt-2 text-sm text-blue-700">
                      <div className="flex items-center">
                        <span className="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded break-all">{user.token}</span>
                        <div className="ml-3 cursor-pointer" onClick={() => { copyToClipboard(user.token) }}>
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

function dashboardVersion(application) {
  if (!application.dashboardVersion) {
    return null
  }

  return (
    <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Dashboard version
        </h3>
      </div>
      <div className="px-4 py-5 sm:px-6">
        <div className="inline-grid">
          <span
            className="mt-1 text-sm text-gray-500">
            {application.dashboardVersion}
          </span>
        </div>
      </div>
    </div>
  )
}

export function copyToClipboard(copyText) {
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(copyText);
  } else {
    unsecuredCopyToClipboard(copyText);
  }
}

function unsecuredCopyToClipboard(text) {
  const textArea = document.createElement("textarea");
  textArea.value = text;
  textArea.style.position = "fixed";
  textArea.style.opacity = "0";
  document.body.appendChild(textArea);
  textArea.select();
  try {
    document.execCommand('copy');
  } catch (err) {
    console.error('Unable to copy to clipboard', err);
  }
  document.body.removeChild(textArea);
}
