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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/ExpandedFooter.jsx
*/

import { Kustomizations } from './Kustomizations';
import { HelmReleases } from './HelmReleases';
import FluxEvents from './FluxEvents';
import { Sources } from './Sources';
import { TerraformResources } from "./TerraformResources";
import { CompactServices } from './CompactServices';
import { ErrorBoundary } from "react-error-boundary";
import { fallbackRender } from "./FallbackRender"

export function ExpandedFooter(props) {
  const { client, fluxState, fluxEvents, sources, handleNavigationSelect, targetReference, selected, store } = props;

  return (
    <div className="flex w-full h-full overscroll-contain">
      <div>
        <div className="w-56 px-4 border-r border-neutral-300 dark:border-neutral-600">
          <SideBar
            navigation={[
              { name: 'Sources', href: '#', count: sources.length },
              { name: 'Kustomizations', href: '#', count: fluxState.kustomizations?.length },
              { name: 'Helm Releases', href: '#', count: fluxState.helmReleases?.length },
              {
                name: "Terraform",
                href: "#",
                count: fluxState.tfResources?.length,
              },
              { name: 'Flux Runtime', href: '#', count: undefined },
              { name: 'Flux Events', href: '#', count: undefined },
            ]}
            selectedMenu={handleNavigationSelect}
            selected={selected}
          />
        </div>
      </div>

      <div className="w-full px-4 overflow-x-hidden overflow-y-scroll">
        <div className="w-full max-w-7xl mx-auto flex-col h-full">
          <div className="pb-24 pt-2">
            {selected === "Kustomizations" &&
              <Kustomizations capacitorClient={client} fluxState={fluxState} targetReference={targetReference} handleNavigationSelect={handleNavigationSelect} />
            }
            {selected === "Helm Releases" &&
              <HelmReleases capacitorClient={client} helmReleases={fluxState.helmReleases} targetReference={targetReference} handleNavigationSelect={handleNavigationSelect} />
            }
            {selected === "Terraform" && (
              <ErrorBoundary fallbackRender={fallbackRender}>
              <TerraformResources
                capacitorClient={client}
                store={store}
                tfResources={fluxState.tfResources}
                targetReference={targetReference}
                handleNavigationSelect={handleNavigationSelect}
              />
              </ErrorBoundary>
            )}
            {selected === "Sources" &&
              <Sources capacitorClient={client} fluxState={fluxState} targetReference={targetReference} />
            }
            {selected === "Flux Runtime" &&
              <CompactServices capacitorClient={client} store={store} services={fluxState.fluxServices} />
            }
            {selected === "Flux Events" &&
              <FluxEvents events={fluxEvents} handleNavigationSelect={handleNavigationSelect} />
            }
          </div>
        </div>
      </div>
    </div>
  )
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
              className={classNames(item.name === selected ? 'bg-white dark:bg-neutral-600 text-black dark:text-neutral-200' : 'text-neutral-700 dark:text-neutral-300 hover:bg-white hover:dark:bg-neutral-600 hover:text-black hover:dark:text-neutral-200',
                  'group flex gap-x-3 p-2 pl-3 text-sm leading-6 rounded-md')}
              onClick={() => selectedMenu(item.name)}
            >
              {item.name}
              {item.count ? (
                <span
                  className="ml-auto w-6 min-w-max whitespace-nowrap rounded-full bg-white dark:bg-neutral-600 px-2.5 py-0.5 text-center text-xs font-medium leading-5 text-neutral-700 dark:text-neutral-200 ring-1 ring-inset ring-neutral-200 dark:ring-neutral-700"
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

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}
