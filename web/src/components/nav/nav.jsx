import React, {Fragment} from 'react';
import { useState } from 'react';
import './nav.css';
import {Disclosure, Menu, Transition} from '@headlessui/react'
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import logo from "./logo.svg";
import DefaultProfilePicture from '../../../src/views/profile/defaultProfilePicture.png';
import { ThemeSelector } from '../../views/repositories/themeSelector';
import { useMatch, useLocation, useParams, useNavigate } from 'react-router-dom'

const navigation = [
  {name: 'Repositories', href: '/repositories'},
  {name: 'Environments', href: '/environments'},
]

const repoNavigation = [
  {name: 'Deployments', href: '/deployments'},
  {name: 'Previews', href: '/previews'},
  {name: 'Commits', href: '/commits'},
  {name: 'Settings', href: '/settings'},
]

const userNavigation = [
  {name: 'CLI', href: '/cli'},
  {name: 'Cloud', href: '/accounts'},
  {name: 'Settings', href: '/settings'},
  {name: 'Theme', href: '#'},
  {name: 'Log out', href: '/logout'},
]

export default function Nav (props) {
  const { store } = props;
  const reduxState = store.getState();
  const [user, setUser] = useState(reduxState.user)
  const [settings, setSettings] = useState(reduxState.settings)
  const [connectedAgents, setConnectedAgents] = useState(reduxState.connectedAgents)
  const [envs, setEnvs] = useState(reduxState.envs)
  const { owner, repo, env, config, action } = useParams()

  store.subscribe(() => {
    const reduxState = store.getState();
    setUser(reduxState.user);
    setSettings(reduxState.settings);
    setConnectedAgents(reduxState.connectedAgents);
    setEnvs(reduxState.envs);
  })

  user.imageUrl = `${settings.scmUrl}/${user.login}.png?size=128`

  const configScreen = useMatch('/repo/:owner/:repo/envs/:env/config/:config/:action?/:nav?') && (action === 'new' || action === 'edit')
  const previewConfigScreen = useMatch('/repo/:owner/:repo/envs/:env/config/:config/:action?/:nav?') && (action === 'new-preview' || action === 'edit-preview')
  const deployment = useMatch('/repo/:owner/:repo/:environment?/:deployment?')
  const commits = useMatch('/repo/:owner/:repo/commits')
  const previewDeployments = useMatch('/repo/:owner/:repo/previews/:deployment?')
  const repoSettings = useMatch('/repo/:owner/:repo/settings/:nav?')
  const repoScreen = deployment || commits || previewDeployments || repoSettings
  const environmentScreen = useMatch('/env/:env/:tab?')
  const deployWizzardScreen = useMatch('/repo/:owner/:repo/envs/:env/deploy')
  const repoWizzardScreen = useMatch('/import-repositories')
  const navigate = useNavigate()

  const loggedIn = user !== undefined;
  if (!loggedIn) {
    return null;
  }

  let menu = <MainMenu items={navigation} />
  let submenu = null
  if (repoScreen) {
    const repoLink = <a href={`${settings.scmUrl}/${owner}/${repo}`} target="_blank" rel="noreferrer" className='externalLink'>{owner}/{repo} <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
    menu = <Crumbs crumb={repoLink} label='Repositories' href="/repositories" />
    submenu = <MainMenu items={repoNavigation} submenu={true} />
  } else if (environmentScreen) {
    menu = <Crumbs crumb={env} label='Environments' href="/environments" />
  } else if (configScreen) {
    menu = <ConfigCrumbs owner={owner} repo={repo} config={config} env={env} />
  } else if (previewConfigScreen) {
    menu = <PreviewConfigCrumbs owner={owner} repo={repo} />
  } else if (deployWizzardScreen) {
    menu = <DeployWizzardCrumbs owner={owner} repo={repo} env={env} />
  } else if (repoWizzardScreen) {
    menu = <ImportRepoCrumbs items={navigation} />
  }

  return (
    <Disclosure as="nav" className={`fixed w-full z-40 bg-white dark:bg-neutral-800 border-b border-neutral-200 dark:border-neutral-700 z-1`}>
      {({open}) => (
        <>
          <div className="">
            <div className="flex justify-between">
              <div className="flex">
                <div className="flex-shrink-0 flex items-center bg-neutral-800 dark:bg-black py-2 px-4">
                  <img
                    className="h-8 w-auto cursor-pointer"
                    src={logo}
                    alt=""
                    onClick={() => {
                      navigate("/");
                      return true
                    }}
                  />
                </div>
                <div className="ml-4 flex">
                  {menu}
                </div>
              </div>
              <div className="flex-1 flex items-center justify-center px-2 sm:ml-6">
                {!settings.licensed &&
                <div className='rounded-lg bg-yellow-100 text-neutral-900 text-sm px-4'>
                  <a href="https://gimlet.io/pricing" className='underline' target="_blank" rel="noopener noreferrer">Licensed for individual and non-profit use.</a>
                </div>
                }
                { settings.trial && <Connecting connectedAgents={connectedAgents} envs={envs} /> }
              </div>
              <div className="hidden sm:flex mr-10 font-sans font-light space-x-4 text-sm text-neutral-500 dark:text-neutral-400">
                {/* <a
                  href="https://gimlet.io/changelog"
                  target="_blank"
                  rel="noreferrer"
                  className="hover:text-neutral-800 dark:hover:text-neutral-300 inline-flex items-center"
                >
                  Changelog
                </a> */}
                <a
                  href="https://gimlet.io/docs"
                  target="_blank"
                  rel="noreferrer"
                  className="hover:text-neutral-800 dark:hover:text-neutral-300 inline-flex items-center"
                >
                  Docs
                </a>
                <a
                  href="https://discord.com/invite/ZwQDxPkYzE"
                  target="_blank"
                  rel="noreferrer"
                  className="hover:text-neutral-800 dark:hover:text-neutral-300 inline-flex items-center"
                >
                  Community
                </a>
              </div>
              <div className="mr-6 ml-2 flex items-center">
                {/* Profile dropdown */}
                <Menu as="div" className="relative">
                  {({open}) => (
                    <>
                      <div>
                        <Menu.Button
                          className="max-w-xs bg-white flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                          <span className="sr-only">Open user menu</span>
                          <img className="h-8 w-8 rounded-full" src={user.imageUrl} alt={user.login} onError={(e) => { e.target.src = DefaultProfilePicture }}/>
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
                          className="origin-top-right absolute right-0 z-10 mt-2 w-48 py-1 card focus:outline-none"
                        >
                          {userNavigation.map((item) => {
                            if (item.name === "Theme") {
                              return (
                                <button
                                  key={item.name}
                                  className="flex items-center justify-between cursor-default px-4 py-2 text-sm font-sans font-light text-neutral-500 dark:text-neutral-200 w-full text-left duration-150"
                                >
                                  {item.name}
                                  <ThemeSelector className="relative z-10 items-end" />
                                </button>
                              )
                            }

                            return (
                              <Menu.Item key={item.name}>
                                {({ active }) => (
                                  <button
                                    href="https://gimlet.io"
                                    className={(
                                      active ? 'bg-neutral-200 text-neutral-700 dark:bg-neutral-500 dark:text-neutral-100' : '') +
                                      ' block px-4 py-2 text-sm font-sans font-light text-neutral-500 dark:text-neutral-200 w-full text-left duration-150'
                                    }
                                    onClick={() => {
                                      if (item.href === '/logout' || item.href === '/accounts') {
                                        window.location.replace(item.href);
                                        return true
                                      }
                                      navigate(item.href);
                                      return true
                                    }}
                                  >
                                    {item.name}
                                  </button>
                                )}
                              </Menu.Item>
                            )})}
                        </Menu.Items>
                      </Transition>
                    </>
                  )}
                </Menu>
              </div>
            </div>
          </div>
          <div className="px-4 sm:px-4">
            {submenu}
          </div>
        </>
      )}
    </Disclosure>
  )
}

function MainMenu(props) {
  const { items, submenu } = props
  const location = useLocation()
  const navigate = useNavigate()

  let current = items.find(i => location.pathname.endsWith(i.href))
  if (!current) {
    current = items[0]
  }

  return (
    items.map((item) => {
      const selected = current.href === item.href;
      return (
        <div
          key={item.href}
          className={(
            selected
              ? 'border-teal-400'
              : 'border-transparent') +
            ' inline-flex items-center border-b-2 py-1'
          }
        >
          <button
            key={item.name}
            className={(
              selected
                ? 'text-neutral-900 dark:text-neutral-200'
                : 'navUnselected') +
              ' inline-flex items-center px-3 py-2 transition-colors duration-150 ease-in-out hover:bg-neutral-200 dark:hover:bg-neutral-600 rounded-md text-sm font-light font-sans'
            }
            aria-current={selected ? 'page' : undefined}
            onClick={() => {
              if (submenu) {
                navigate(location.pathname.replace(current.href, "") + item.href);
              } else {
                navigate(item.href);
              }
              return true
            }}
          >
            {item.name}
          </button>
        </div>
      )
    })
  )
}

function Crumbs(props) {
  const { label, crumb, href } = props
  const navigate = useNavigate()

  return (
    <span className='inline-flex items-center text-sm text-neutral-500 dark:text-neutral-300 font-light py-1 border-b-2 border-transparent'>
    <button
      href="https://gimlet.io"
      className='navUnselected pl-3 py-2'
      onClick={() => {
        navigate(href);
        return true
      }}
    >{label}</button>
    <span className='px-4'>/</span>
    <span className='text-black dark:text-neutral-200 font-normal cursor-default'>{crumb}</span>
    </span>
  )
}

function ImportRepoCrumbs(props) {
  const { items } = props
  const navigate = useNavigate()

  return (
    items.map((item) => {
      return (
        <div
          key={item.href}
          className='border-transparent inline-flex items-center border-b-2 py-1'
        >
          <button
            key={item.name}
            href="https://gimlet.io"
            className='navUnselected inline-flex items-center px-3 py-2 transition-colors duration-150 ease-in-out hover:bg-neutral-200 dark:hover:bg-neutral-600 rounded-md text-sm font-light font-sans'
            onClick={() => {
              navigate(item.href);
              return true
            }}
          >
            {item.name}
          </button>
        </div>
      )
    })
  )
}

function ConfigCrumbs(props) {
  const { owner, repo, config, env } = props
  const navigate = useNavigate()

  return (
    <span className='inline-flex items-center text-sm text-neutral-500 dark:text-neutral-300 font-light py-1 border-b-2 border-transparent'>
    <button
      href="https://gimlet.io"
      className='navUnselected pl-3 py-2'
      onClick={() => {
        navigate('/repositories');
        return true
      }}
    >Repositories</button>
    <span className='px-4'>/</span>
    <button
      href="https://gimlet.io"
      className='navUnselected py-2'
      onClick={() => {
        navigate(`/repo/${owner}/${repo}`);
        return true
      }}
    >{owner}/{repo}</button>
    <span className='px-5'>/</span>
    <span className='text-black dark:text-neutral-100 font-normal cursor-default'>{env}/{config}</span>
    </span>
  )
}

function DeployWizzardCrumbs(props) {
  const { owner, repo, env } = props
  const navigate = useNavigate()

  return (
    <span className='inline-flex items-center text-sm text-neutral-500 font-light py-1 border-b-2 border-transparent'>
    <button
      href="https://gimlet.io"
      className='navUnselected pl-3 py-2'
      onClick={() => {
        navigate('/repositories');
        return true
      }}
    >Repositories</button>
    <span className='px-4'>/</span>
    <button
      href="https://gimlet.io"
      className='navUnselected py-2'
      onClick={() => {
        navigate(`/repo/${owner}/${repo}`);
        return true
      }}
    >{owner}/{repo}</button>
    <span className='px-5'>/</span>
    <span className='text-black dark:text-neutral-100 font-normal cursor-default'>{env}</span>
    </span>
  )
}

function PreviewConfigCrumbs(props) {
  const { owner, repo } = props
  const navigate = useNavigate()

  return (
    <span className='inline-flex items-center text-sm text-neutral-500 font-light py-1 border-b-2 border-transparent'>
    <button
      href="https://gimlet.io"
      className='navUnselected pl-3 py-2'
      onClick={() => {
        navigate('/repositories');
        return true
      }}
    >Repositories</button>
    <span className='px-4'>/</span>
    <button
      href="https://gimlet.io"
      className='navUnselected py-2'
      onClick={() => {
        navigate(`/repo/${owner}/${repo}`);
        return true
      }}
    >{owner}/{repo}</button>
    <span className='px-4'>/</span>
    <span className='text-black dark:text-neutral-100 font-normal cursor-default'>Preview Config</span>
    </span>
  )
}

function Connecting(props) {
  const { connectedAgents, envs } = props

  if (envs.length !== 1) {
    return null
  }

  const env = envs[0]
  const isOnline = Object.keys(connectedAgents).includes(env.name)

  if (isOnline) {
    return null
  }

  const expiringAt = new Date(env.expiry * 1000);
  const expired = expiringAt < new Date()
  if (expired) {
    return (
      <div className='rounded-lg bg-red-100 text-red-900 text-sm px-4 mx-1'>
        <a href={"/env/"+env.name} className='underline' rel="noopener noreferrer">
          <span>Ephemeral cluster expired</span>
        </a>
      </div>
    )
  }

  const age = new Date() - expiringAt
  if (age > 5*60*1000) {
    return (
      <div className='rounded-lg bg-red-100 text-red-900 text-sm px-4 mx-1'>
        <a href="https://gimlet.io/docs/learn-more/contact-us" className='underline' target="_blank" rel="noopener noreferrer">
          <span>Ephemeral cluster stuck - contact support</span>
        </a>
      </div>
    )
  }

  return (
    <div className='rounded-lg bg-blue-100 text-blue-900 text-sm px-4'>
      <a href={"/env/"+env.name} className='underline' rel="noopener noreferrer">
        <span>Ephemeral cluster starting up</span>
        <svg className="animate-spin h-3 w-3 text-black inline ml-1" xmlns="http://www.w3.org/2000/svg" fill="none"
            viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4}></circle>
            <path className="opacity-75" fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </a>
    </div>
  )
}
