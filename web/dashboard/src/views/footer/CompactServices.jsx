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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/CompactServices.jsx
*/

import React, { memo } from 'react';
import { CompactService } from "./service"

export const CompactServices = memo(function CompactServices(props) {
  const { capacitorClient, store, services } = props

  return (
    <div className="space-y-4">
      {
        services?.map((service) => {
          return (
            <CompactService
              key={`${service.deployment.metadata.namespace}/${service.deployment.metadata.name}`}
              service={service}
              capacitorClient={capacitorClient}
              store={store}
            />
          )
        })
      }
    </div>
  )
})
