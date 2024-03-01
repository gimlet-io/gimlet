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
*/

import React from 'react';
import { Logs } from '../logs'
import { Describe } from './Describe'
import { Pod, podContainers } from './Service'

export function CompactService(props) {
  const { service, capacitorClient, store } = props;
  const deployment = service.deployment;

  return (
    <div className="w-full flex items-center justify-between space-x-6 bg-white pb-6 rounded-lg border border-neutral-300 shadow-lg">
      <div className="flex-1">
        <h3 className="flex text-lg font-bold rounded p-4">
          <span className="cursor-pointer">{deployment.metadata.name}</span>
          <div className="flex items-center ml-auto space-x-2">
            {deployment &&
              <>
                <Logs
                  capacitorClient={capacitorClient}
                  store={store}
                  namespace={deployment.metadata.namespace}
                  deployment={deployment.metadata.name}
                  containers={podContainers(service.pods)}
                />
                <Describe
                  capacitorClient={capacitorClient}
                  namespace={deployment.metadata.namespace}
                  deployment={deployment.metadata.name}
                  pods={service.pods}
                  store={store}
                />
              </>
            }
          </div>
        </h3>
        <div>
          <div className="grid grid-cols-12 px-4">
            <div className="col-span-5 space-y-4">
              <div>
                <p className="text-base text-neutral-600">Pods</p>
                {
                  service.pods.map((pod) => (
                    <Pod key={pod.metadata.name} pod={pod} />
                  ))
                }
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
