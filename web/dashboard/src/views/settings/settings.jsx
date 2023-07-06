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

    this.deleteUser = this.deleteUser.bind(this)
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

  deleteUser(login) {
    this.props.store.dispatch({
      type: ACTION_TYPE_POPUPWINDOWPROGRESS, payload: {
        header: "Deleting..."
      }
    });

    this.props.gimletClient.deleteUser(login)
      .then(() => {
        this.props.store.dispatch({
          type: ACTION_TYPE_POPUPWINDOWSUCCESS, payload: {
            header: "Success",
            message: "User deleted"
          }
        });

        this.props.gimletClient.getUsers()
          .then(data => {
            this.props.store.dispatch({
              type: ACTION_TYPE_USERS,
              payload: data
            });
          }, () => {/* Generic error handler deals with it */
          });
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
    const { settings, application, users, input, saveButtonTriggered } = this.state;
    const sortedUsers = this.sortAlphabetically(users);

    return (
      <div>
        <main>
          <div className="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <div className="px-4 sm:px-0">
              <div>
                {settings.provider === "github" &&
                  githubAppSettings(application)
                }
                {(!settings.provider || settings.provider === "") &&
                  gimletInstaller()
                }
                <div className="my-4 bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
                  <div className="px-4 py-5 sm:px-6">
                    <h3 className="text-lg leading-6 font-medium text-gray-900">
                      Users and API Keys
                    </h3>
                  </div>
                  <Users
                    users={sortedUsers}
                    scmUrl={settings.scmUrl}
                    deleteUser={this.deleteUser}
                  />
                  <div className="px-4 py-5 sm:px-6">
                    <h3 className="text-lg leading-6 font-medium text-gray-900">Create an API key</h3>
                    <div>
                      <input
                        onChange={e => this.setState({ input: e.target.value })}
                        className="shadow appearance-none border rounded w-full my-4 py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                        id="environment"
                        type="text"
                        value={input}
                        placeholder="Please enter an API key name" />
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
                </div>
                {dashboardVersion(application)}
              </div>
            </div>
          </div>
        </main >
      </div >
    )
  }
}

function githubAppSettings(application) {
  if (application.name === "") {
    return null
  }

  return (
    <div className="bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200 my-4">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Github Integration
        </h3>
      </div>
      <div className="px-4 py-5 sm:px-6">
        <div className="inline-grid">
          <span
            onClick={() => window.open(application.appSettingsURL)}
            className="mt-1 text-sm text-gray-500 hover:text-gray-600 cursor-pointer">
            Github Application
            <svg xmlns="http://www.w3.org/2000/svg"
              className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
              viewBox="0 0 24 24">
              <path d="M0 0h24v24H0z" fill="none" />
              <path
                d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
            </svg>
          </span>
          <span>
            <a
              href={application.installationURL}
              rel="noreferrer"
              target="_blank"
              className="mt-1 text-sm text-gray-500 hover:text-gray-600">
              Application installation
              <svg xmlns="http://www.w3.org/2000/svg"
                className="inline fill-current text-gray-500 hover:text-gray-700 ml-1" width="12" height="12"
                viewBox="0 0 24 24">
                <path d="M0 0h24v24H0z" fill="none" />
                <path
                  d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z" />
              </svg>
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
          Github Integration
        </h3>
      </div>
      <div className="px-4 py-5 sm:p-6">
        <Installer />
      </div>
    </div>
  )
}

function Users({ users, scmUrl, deleteUser }) {
  return (

      <div className="px-4 py-5 sm:px-6">
        {users.map(user => (
          <div key={user.login} className="flex justify-between p-2 hover:bg-gray-100 rounded">
            <div className="inline-flex items-center">
              <img
                className="h-8 w-8 rounded-full text-2xl font-medium text-gray-900"
                src={`${scmUrl}/${user.login}.png?size=128`}
                onError={(e) => { e.target.src = DefaultProfilePicture }}
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
                    <h3 className="text-sm font-medium text-blue-800">API key:</h3>
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
            <div className="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg"
                onClick={() => {
                  // eslint-disable-next-line no-restricted-globals
                  confirm(`Are you sure you want to delete ${user.login}?`) &&
                    deleteUser(user.login);
                }}
                className="items-center cursor-pointer inline text-red-400 hover:text-red-600 opacity-70 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </div>
          </div>
        ))}
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
