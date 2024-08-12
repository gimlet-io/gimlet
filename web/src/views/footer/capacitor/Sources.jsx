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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/Sources.jsx
*/

import React, { useState, useMemo } from 'react';
import { filterResources } from './utils.js';
import { Source } from "./Source.jsx"

export function Sources(props) {
  const { capacitorClient, fluxState, targetReference } = props
  const [filter, setFilter] = useState(false)
  const sortedSources = useMemo(() => {
    const sources = [];
    if (fluxState.ociRepositories) {
      sources.push(...fluxState.ociRepositories)
      sources.push(...fluxState.gitRepositories)
      sources.push(...fluxState.buckets)
    }
    return [...sources].sort((a, b) => a.metadata.name.localeCompare(b.metadata.name));
  }, [fluxState]);

  const filteredSources = filterResources(sortedSources, filter)

  return (
    <div className="space-y-4">
      <button className={(filter ? "text-blue-50 bg-blue-500 dark:bg-blue-800" : "bg-neutral-50 dark:bg-neutral-600 text-neutral-600 dark:text-neutral-200") + " rounded-full px-3"}
        onClick={() => setFilter(!filter)}
      >
        Filter errors
      </button>
      {
        filteredSources?.map(source =>
          <Source
            key={"source-" + source.metadata.namespace + source.metadata.name}
            capacitorClient={capacitorClient}
            source={source}
            targetReference={targetReference}
          />
        )
      }
    </div>
  )
}
