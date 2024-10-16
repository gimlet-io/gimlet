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

import React, { useMemo, useState } from "react";
import { TerraformResource } from "./TerraformResource";
import { filterResources } from "./utils";
// import FilterBar from "./FilterBar";
import { ErrorBoundary } from "react-error-boundary";
import { fallbackRender } from "./FallbackRender"

export function TerraformResources(props) {
  const {
    capacitorClient,
    tfResources,
    targetReference,
    handleNavigationSelect,
  } = props;
  const [filters, setFilters] = useState([]);
  const sortedHelmReleases = useMemo(() => {
    if (!tfResources) {
      return null;
    }

    return [...tfResources].sort((a, b) =>
      a.metadata.name.localeCompare(b.metadata.name),
    );
  }, [tfResources]);
  
  // const filteredHelmReleases = filterResources(sortedHelmReleases, filters);

  return (
    <div className="space-y-4">
      {/* <FilterBar
        properties={["Name", "Namespace", "Errors"]}
        filters={filters}
        change={setFilters}
      /> */}
      <ErrorBoundary fallbackRender={fallbackRender}>
      {sortedHelmReleases?.map((resource) => (
        <TerraformResource
          key={"hr-" + resource.metadata.namespace + resource.metadata.name}
          capacitorClient={capacitorClient}
          item={resource}
          handleNavigationSelect={handleNavigationSelect}
          targetReference={targetReference}
        />
      ))}
      </ErrorBoundary>
    </div>
  );
}
