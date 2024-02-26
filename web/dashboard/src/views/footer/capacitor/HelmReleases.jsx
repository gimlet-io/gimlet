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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/HelmReleases.jsx
*/

import React, { useMemo, useState } from 'react';
import { HelmRelease } from "./HelmRelease"
import { filterResources } from './utils.js';

export function HelmReleases(props) {
  const { capacitorClient, helmReleases, targetReference, handleNavigationSelect } = props
  const [filter, setFilter] = useState(false)
  const sortedHelmReleases = useMemo(() => {
    if (!helmReleases) {
      return null;
    }

    return [...helmReleases].sort((a, b) => a.metadata.name.localeCompare(b.metadata.name));
  }, [helmReleases]);

  const filteredHelmReleases = filterResources(sortedHelmReleases, filter)

  return (
    <div className="space-y-4">
      <button className={(filter ? "text-blue-50 bg-blue-600" : "bg-gray-50 text-gray-600") + " rounded-full px-3"}
        onClick={() => setFilter(!filter)}
      >
        Filter errors
      </button>
      {
        filteredHelmReleases?.map(helmRelease =>
          <HelmRelease
            key={"hr-" + helmRelease.metadata.namespace + helmRelease.metadata.name}
            capacitorClient={capacitorClient}
            item={helmRelease}
            handleNavigationSelect={handleNavigationSelect}
            targetReference={targetReference}
          />
        )}
    </div>
  )
}