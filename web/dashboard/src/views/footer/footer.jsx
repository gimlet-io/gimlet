/*
Copyright 2023 The Capacitor Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/Footer.js
*/

import { ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/solid';
import React, { memo, Component } from 'react';
import { GitRepositories, Kustomizations, HelmReleases, CompactServices, Summary } from './fluxState';

const Footer = memo(class DeployPanel extends Component {
  constructor(props) {
    super(props);
    let reduxState = this.props.store.getState();

    this.state = {
      fluxStates: reduxState.fluxState,
      expanded: false,
      selectedTab: "Kustomizations",
      targetReference: "",
    };
    this.props.store.subscribe(() => this.setState({ fluxStates: reduxState.fluxState }));

    this.handleToggle = this.handleToggle.bind(this)
    this.handleNavigationSelect = this.handleNavigationSelect.bind(this)
  }

  handleToggle() {
    this.setState(prevState => ({
      expanded: !prevState.expanded
    }));
  }

  handleNavigationSelect(selectedNav, ref) {
    this.setState({
      selectedTab: selectedNav,
      targetReference: ref
    })
  }

  render() {
    const { gimletClient, store } = this.props;
    const { fluxStates, expanded, selectedTab, targetReference } = this.state;

    if (!fluxStates || Object.keys(fluxStates).length === 0) {
      return null
    }

    // TODO poorly implementation
    // TODO footer re-rendering misbehavior
    const fluxState = fluxStates["preview"];

    return (
      <div aria-labelledby="slide-over-title" role="dialog" aria-modal="true" className={`fixed inset-x-0 bottom-0 bg-neutral-200 z-40 border-t border-neutral-300 ${expanded ? 'h-4/5' : ''}`}>
        <div className={`flex justify-between w-full ${expanded ? '' : 'h-full'}`}>
          <div
            className='h-auto w-full cursor-pointer px-16 py-4 flex gap-x-12'
            onClick={this.handleToggle} >
            {!expanded &&
              <div className="grid grid-cols-3">
                {Object.keys(fluxStates).slice(0, 3).map(env => {
                  const fluxState = fluxStates[env];

                  return (
                    <div className="w-full truncate" key={env}>
                      <p className="font-semibold text-neutral-700">{`${env.toUpperCase()}`}</p>
                      <div className="ml-2">
                        <Summary resources={fluxState.gitRepositories} label="SOURCES" />
                        <Summary resources={fluxState.kustomizations} label="KUSTOMIZATIONS" />
                        <Summary resources={fluxState.helmReleases} label="HELM-RELEASES" />
                      </div>
                    </div>
                  )
                })}
              </div>
            }
          </div>
          <div className='px-4 py-2'>
            <button
              onClick={this.handleToggle}
              type="button" className="ml-1 rounded-md hover:bg-white hover:text-black text-neutral-700 p-1">
              <span className="sr-only">{expanded ? 'Close panel' : 'Open panel'}</span>
              {expanded ? <ArrowDownIcon className="h-5 w-5" aria-hidden="true" /> : <ArrowUpIcon className="h-5 w-5" aria-hidden="true" />}
            </button>
          </div>
        </div>
        {expanded &&
          <div>
            {/* <nav className="flex space-x-8 px-6 pt-4" aria-label="Tabs">
              {Object.keys(fluxStates).map((env) => (
                <span
                  key={env}
                  onClick={() => { setSelectedEnv(env); return false }}
                  className={classNames(
                    env === selectedEnv
                      ? 'border-indigo-500 text-gray-900'
                      : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                    'whitespace-nowrap border-b-2 pb-2 px-1 text-sm font-medium cursor-pointer'
                  )}
                  aria-current={env === selectedEnv ? 'page' : undefined}
                >
                  {env.toUpperCase()}
                </span>
              ))}
            </nav> */}
            <div className="flex w-full h-full">
              <div>
                <div className="w-56 px-4 border-r border-neutral-300">
                  <SideBar
                    navigation={[
                      { name: 'Sources', href: '#', count: fluxState.gitRepositories.length },
                      { name: 'Kustomizations', href: '#', count: fluxState.kustomizations.length },
                      { name: 'Helm Releases', href: '#', count: fluxState.helmReleases.length },
                      { name: 'Flux Runtime', href: '#', count: fluxState.fluxServices.length },
                    ]}
                    selectedMenu={this.handleNavigationSelect}
                    selected={selectedTab}
                  />
                </div>
              </div>

              <div className="w-full px-4 overflow-x-hidden overflow-y-scroll">
                <div className="w-full max-w-7xl mx-auto flex-col h-full">
                  <div className="pb-24 pt-2">
                    {selectedTab === "Kustomizations" &&
                      <Kustomizations gimletClient={gimletClient} fluxState={fluxState} handleNavigationSelect={this.handleNavigationSelect} />
                    }
                    {selectedTab === "Helm Releases" &&
                      <HelmReleases gimletClient={gimletClient} helmReleases={fluxState.helmReleases} />
                    }
                    {selectedTab === "Sources" &&
                      <GitRepositories gimletClient={gimletClient} gitRepositories={fluxState.gitRepositories} targetReference={targetReference} />
                    }
                    {selectedTab === "Flux Runtime" &&
                      <CompactServices gimletClient={gimletClient} store={store} services={fluxState.fluxServices} />
                    }
                  </div>
                </div>
              </div>
            </div>
          </div>
        }
      </div>
    )
  }
})

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}

function SideBar(props) {
  const { navigation, selectedMenu, selected } = props;

  return (
    <nav className="flex flex-1 flex-col" aria-label="Sidebar">
      <ul className="space-y-1">
        {navigation.map((item) => (
          <li key={item.name}>
            <a
              href={item.href}
              className={classNames(item.name === selected ? 'bg-white text-black' : 'text-neutral-700 hover:bg-white hover:text-black',
                'group flex gap-x-3 p-2 pl-3 text-sm leading-6 rounded-md')}
              onClick={() => selectedMenu(item.name)}
            >
              {item.name}
              {item.count ? (
                <span
                  className="ml-auto w-6 min-w-max whitespace-nowrap rounded-full bg-white px-2.5 py-0.5 text-center text-xs font-medium leading-5 text-neutral-700 ring-1 ring-inset ring-neutral-200"
                  aria-hidden="true"
                >
                  {item.count}
                </span>
              ) : null}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
};

export default Footer;
