import React, {Component, Fragment} from 'react';
import './nav.css';
import {Disclosure, Menu, Transition} from '@headlessui/react'
import {MenuIcon, XIcon} from '@heroicons/react/outline'
import logo from './logo.svg';
import {ACTION_TYPE_SEARCH} from "../../redux/redux";

const navigation = [
  {name: 'Services', href: '/services'},
  {name: 'Repositories', href: '/repositories'},
  {name: 'Environments', href: '/environments'},
]
const userNavigation = [
  {name: 'Profile', href: '/profile'},
  {name: 'Sign out', href: '/logout'},
]

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}

export default class Nav extends Component {
  constructor(props) {
    super(props);

    // default state
    let reduxState = this.props.store.getState();
    this.state = {
      user: reduxState.user
    }

    // handling API and streaming state changes
    this.props.store.subscribe(() => {
      let reduxState = this.props.store.getState();

      this.setState({user: reduxState.user});
    });

  }

  render() {
    const {user} = this.state;
    const {store} = this.props;

    const loggedIn = user !== undefined;
    if (!loggedIn) {
      return null;
    }

    user.imageUrl = `https://github.com/${user.login}.png?size=128`

    return (
      <Disclosure as="nav" className="bg-white border-b border-gray-200">
        {({open}) => (
          <>
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <div className="flex justify-between h-16">
                <div className="flex">
                  <div className="flex-shrink-0 flex items-center">
                    <img
                      className="h-8 w-auto"
                      src={logo}
                      alt="Workflow"
                    />
                  </div>
                  <div className="hidden sm:-my-px sm:ml-6 sm:flex sm:space-x-8">
                    {navigation.map((item) => {
                      const selected = this.props.location.pathname === item.href;
                      return (
                        <button
                          key={item.name}
                          href="#"
                          className={classNames(
                            selected
                              ? 'border-indigo-500 text-gray-900'
                              : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                            'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium'
                          )}
                          aria-current={selected ? 'page' : undefined}
                          onClick={() => {
                            this.props.history.push(item.href);
                            return true
                          }}
                        >
                          {item.name}
                        </button>
                      )
                    })
                    }
                  </div>
                </div>
                <div className="flex-1 flex items-center justify-center px-2 sm:ml-6 sm:justify-end">
                  <div className="max-w-lg w-full lg:max-w-xs">
                    <label htmlFor="search" className="sr-only">Search</label>
                    <div className="relative">
                      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <svg className="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"
                             fill="currentColor" aria-hidden="true">
                          <path fillRule="evenodd"
                                d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
                                clipRule="evenodd"/>
                        </svg>
                      </div>
                      <input id="search" name="search"
                             className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                             placeholder="Search" type="search" onChange={(e) => store.dispatch({
                        type: ACTION_TYPE_SEARCH,
                        payload: {filter: e.target.value}
                      })}/>
                    </div>
                  </div>
                </div>
                <div className="hidden sm:-my-px sm:ml-6 sm:flex">
                  <a
                    href="https://gimlet.io/docs"
                    target="_blank"
                    rel="noreferrer"
                    className="text-gray-500 inline-flex items-center px-1 pt-1 text-sm"
                  >
                    Docs
                  </a>
                  <a
                    href="https://discord.com/invite/ZwQDxPkYzE"
                    target="_blank"
                    rel="noreferrer"
                    className="text-gray-500 inline-flex items-center px-1 pt-1 text-sm"
                  >
                    Community
                  </a>
                </div>
                <div className="hidden sm:ml-2 sm:flex sm:items-center">
                  {/*<button*/}
                  {/*  className="bg-white p-1 rounded-full text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">*/}
                  {/*  <span className="sr-only">View notifications</span>*/}
                  {/*  <BellIcon className="h-6 w-6" aria-hidden="true"/>*/}
                  {/*</button>*/}

                  {/* Profile dropdown */}
                  <Menu as="div" className="relative">
                    {({open}) => (
                      <>
                        <div>
                          <Menu.Button
                            className="max-w-xs bg-white flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                            <span className="sr-only">Open user menu</span>
                            <img className="h-8 w-8 rounded-full" src={user.imageUrl} alt=""/>
                          </Menu.Button>
                        </div>
                        <Transition
                          show={open}
                          as={Fragment}
                          enter="transition ease-out duration-200"
                          enterFrom="transform opacity-0 scale-95"
                          enterTo="transform opacity-100 scale-100"
                          leave="transition ease-in duration-75"
                          leaveFrom="transform opacity-100 scale-100"
                          leaveTo="transform opacity-0 scale-95"
                        >
                          <Menu.Items
                            static
                            className="origin-top-right absolute right-0 mt-2 w-48 rounded-md shadow-lg py-1 bg-white ring-1 ring-black ring-opacity-5 focus:outline-none"
                          >
                            {userNavigation.map((item) => (
                              <Menu.Item key={item.name}>
                                {({active}) => (
                                  <button
                                    href="#"
                                    className={classNames(
                                      active ? 'bg-gray-100' : '',
                                      'block px-4 py-2 text-sm text-gray-700 w-full text-left'
                                    )}
                                    onClick={() => {
                                      if (item.href === '/logout') {
                                        window.location.replace("/logout");
                                        return true
                                      }
                                      this.props.history.push(item.href);
                                      return true
                                    }}
                                  >
                                    {item.name}
                                  </button>
                                )}
                              </Menu.Item>
                            ))}
                          </Menu.Items>
                        </Transition>
                      </>
                    )}
                  </Menu>
                </div>
                <div className="-mr-2 flex items-center sm:hidden">
                  {/* Mobile menu button */}
                  <Disclosure.Button
                    className="bg-white inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                    <span className="sr-only">Open main menu</span>
                    {open ? (
                      <XIcon className="block h-6 w-6" aria-hidden="true"/>
                    ) : (
                      <MenuIcon className="block h-6 w-6" aria-hidden="true"/>
                    )}
                  </Disclosure.Button>
                </div>
              </div>
            </div>

            <Disclosure.Panel className="sm:hidden">
              <div className="pt-2 pb-3 space-y-1">
                {navigation.map((item) => {
                  const selected = this.props.location.pathname === item.href;
                  return (
                    <button
                      key={item.name}
                      href="#"
                      className={classNames(
                        selected
                          ? 'bg-indigo-50 border-indigo-500 text-indigo-700'
                          : 'border-transparent text-gray-600 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-800',
                        'block pl-3 pr-4 py-2 border-l-4 text-base font-medium'
                      )}
                      aria-current={selected ? 'page' : undefined}
                      onClick={() => {
                        this.props.history.push(item.href);
                        return true
                      }}
                    >
                      {item.name}
                    </button>
                  )
                })}
              </div>
              <div className="pt-4 pb-3 border-t border-gray-200">
                <div className="flex items-center px-4">
                  <div className="flex-shrink-0">
                    <img className="h-10 w-10 rounded-full" src={user.imageUrl} alt=""/>
                  </div>
                  <div className="ml-3">
                    <div className="text-sm font-medium text-gray-600">{user.login}</div>
                  </div>
                  {/*<button*/}
                  {/*  className="ml-auto bg-white flex-shrink-0 p-1 rounded-full text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">*/}
                  {/*  <span className="sr-only">View notifications</span>*/}
                  {/*  <BellIcon className="h-6 w-6" aria-hidden="true"/>*/}
                  {/*</button>*/}
                </div>
                <div className="mt-3 space-y-1">
                  {userNavigation.map((item) => (
                    <a
                      key={item.name}
                      href={item.href}
                      className="block px-4 py-2 text-base font-medium text-gray-500 hover:text-gray-800 hover:bg-gray-100"
                    >
                      {item.name}
                    </a>
                  ))}
                </div>
              </div>
            </Disclosure.Panel>
          </>
        )}
      </Disclosure>
    )
  }
}
