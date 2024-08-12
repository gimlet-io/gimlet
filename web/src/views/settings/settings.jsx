import React, { Component, useState } from 'react';
import { InformationCircleIcon } from '@heroicons/react/24/solid';
import DefaultProfilePicture from '../profile/defaultProfilePicture.png';
import {
  ACTION_TYPE_USERS,
  ACTION_TYPE_POPUPWINDOWRESET,
  ACTION_TYPE_POPUPWINDOWPROGRESS,
  ACTION_TYPE_POPUPWINDOWSUCCESS,
  ACTION_TYPE_POPUPWINDOWERROR
} from '../../redux/redux';
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

export default class Settings extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      application: reduxState.application,
      settings: reduxState.settings,
      users: reduxState.users,
      login: reduxState.user.login,
      input: "",
      saveButtonTriggered: false,
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({ application: reduxState.application });
      this.setState({ settings: reduxState.settings });
      this.setState({ users: reduxState.users });
      this.setState({ login: reduxState.user.login });
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
    const { settings, application, users, login, input, saveButtonTriggered } = this.state;
    const sortedUsers = this.sortAlphabetically(users);

    return (
      <div>
        <main>
          <div className="max-w-7xl mx-auto pt-32 pb-16 sm:px-6 lg:px-8">
            <div className="px-4 sm:px-0 space-y-6">
              {settings.provider === "github" &&
                githubAppSettings(application)
              }
              <div className="card">
                <div className="px-4 py-5 sm:px-6">
                  <h3 className="text-lg leading-6 font-medium">
                    API Keys
                  </h3>
                </div>
                <Users
                  users={sortedUsers}
                  login={login}
                  scmUrl={settings.scmUrl}
                  deleteUser={this.deleteUser}
                />
                <div className="px-4 py-5 sm:px-6">
                  <h3 className="text-lg leading-6 font-medium">Create an API key</h3>
                  <div>
                    <input
                      onChange={e => this.setState({ input: e.target.value })}
                      className="input my-4"
                      id="environment"
                      type="text"
                      value={input}
                      placeholder="Please enter an API key name" />
                    <div className="p-0 flow-root">
                      <span className="inline-flex rounded-md shadow-sm gap-x-3 float-right">
                        <button
                          disabled={input === "" || saveButtonTriggered}
                          onClick={() => this.save()}
                          className={input === "" || saveButtonTriggered ? "primaryButtonDisabled px-4" : "primaryButton px-4"}>
                          Create
                        </button>
                      </span>
                    </div>
                  </div>
                </div>
              </div>
              {dashboardVersion(application)}
              <Licensed settings={this.state.settings}/>
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
    <div className="card">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium0">
          Github Integration
        </h3>
      </div>
      <div className="px-4 py-5 sm:px-6">
        <div className="inline-grid text-sm">
          <a
            href={application.appSettingsURL}
            rel="noreferrer"
            target="_blank"
            className="externalLink mt-1">
              Github Application
            <ArrowTopRightOnSquareIcon className="externalLinkIcon ml-1" aria-hidden="true" />
          </a>
          <a
            href={application.installationURL}
            rel="noreferrer"
            target="_blank"
            className="externalLink mt-1">
              Application installation
            <ArrowTopRightOnSquareIcon className="externalLinkIcon ml-1" aria-hidden="true" />
          </a>
        </div>
      </div>
    </div>
  )
}

function Users({ users, login, scmUrl, deleteUser }) {
  return (
      <div className="px-4 py-5 sm:px-6 text-sm">
        {users.map(user => (
          <div key={user.login} className="flex justify-between p-2 hover:bg-neutral-100 dark:hover:bg-neutral-600 rounded">
            <div className="inline-flex items-center">
              <img
                className="h-8 w-8 rounded-full text-2xl font-medium text-neutral-900"
                src={`${scmUrl}/${user.login}.png?size=128`}
                onError={(e) => { e.target.src = DefaultProfilePicture }}
                alt={user.login} />
              <div className="ml-4">{user.login}</div>
            </div>
            <TokenInfo 
              token={user.token}
            />
            {user.login !== login &&
              <div className="flex items-center">
                <svg xmlns="http://www.w3.org/2000/svg"
                  onClick={() => {
                    // eslint-disable-next-line no-restricted-globals
                    confirm(`Are you sure you want to delete ${user.login}?`) &&
                      deleteUser(user.login);
                  }}
                  className="items-center cursor-pointer inline text-red-500 dark:text-red-700 hover:text-red-400 dark:hover:text-red-800 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </div>
            }
          </div>
        ))}
      </div>
  )
}

function TokenInfo({ token }) {
  const [isCopied, setIsCopied] = useState(false);

  const handleCopyClick = () => {
    setIsCopied(true);

    setTimeout(() => {
      setIsCopied(false);
    }, 2000);
  };

  if (!token) {
    return null
  }

  return (
    <div className="rounded-md bg-blue-50 p-4 w-5/6">
      <div className="flex">
        <div className="flex-shrink-0">
          <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
        </div>
        <div className="ml-3">
          <h3 className="text-sm font-medium text-blue-800">API key:</h3>
          <div className="mt-2 text-sm text-blue-700">
            <div className="flex items-center">
              <span className="text-xs font-mono bg-blue-100 text-blue-500 font-medium px-1 py-1 rounded break-all">{token}</span>
              <div className="relative ml-3 cursor-pointer" onClick={() => {
                copyToClipboard(token);
                handleCopyClick();
              }}>
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-blue-400 hover:text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
                {isCopied && (
                  <div className="absolute top-6 right-0">
                    <div className="p-2 bg-indigo-600 select-none text-white inline-block rounded">
                      Copied!
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>)
};

export const Licensed = (props) => {
  console.log(props.settings)
  if (!props.settings.licensed) {
    return null
  }

  return (
    <div className="card">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium">
          License
        </h3>
      </div>
      <div className="px-4 py-5 sm:px-6">
        <div className="inline-grid">
          <span
            className="mt-1 text-sm">
            {JSON.stringify(props.settings.licensed)}
          </span>
        </div>
      </div>
    </div>
  )
}

function dashboardVersion(application) {
  if (!application.dashboardVersion) {
    return null
  }

  return (
    <div className="card">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium">
          Gimlet Version
        </h3>
      </div>
      <div className="px-4 py-5 sm:px-6">
        <div className="inline-grid">
          <span
            className="mt-1 text-sm">
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
