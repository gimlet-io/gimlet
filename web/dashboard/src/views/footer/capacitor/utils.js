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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/utils.js
*/

import jp from 'jsonpath'

export function findSource(sources, reconciler) {
  let namespace = reconciler.metadata.namespace
  if (reconciler.spec.sourceRef.namespace) { // namespace is not mandatory
    namespace = reconciler.spec.sourceRef.namespace
  }

  return sources.find((source) => source.kind === reconciler.spec.sourceRef.kind &&
    source.metadata.name === reconciler.spec.sourceRef.name &&
    source.metadata.namespace === namespace)
}

export function filterResources(resources, filterErrors) {
  let filteredResources = resources;
  if (filterErrors) {
    filteredResources = filteredResources.filter(resource => {
      const readyConditions = jp.query(resource.status, '$..conditions[?(@.type=="Ready")]');
      const readyCondition = readyConditions.length === 1 ? readyConditions[0] : undefined
      const ready = readyCondition && readyConditions[0].status === "True"

      const dependencyNotReady = readyCondition && readyCondition.reason === "DependencyNotReady"

      const readyTransitionTime = readyCondition ? readyCondition.lastTransitionTime : undefined
      const parsed = Date.parse(readyTransitionTime, "yyyy-MM-dd'T'HH:mm:ss");
      const fiveMinutesAgo = new Date();
      fiveMinutesAgo.setMinutes(fiveMinutesAgo.getMinutes() - 5);
      const stalled = fiveMinutesAgo > parsed

      const reconcilingConditions = jp.query(resource.status, '$..conditions[?(@.type=="Reconciling")]');
      const reconcilingCondition = reconcilingConditions.length === 1 ? reconcilingConditions[0] : undefined
      const reconciling = reconcilingCondition && reconcilingCondition.status === "True"

      const fetchFailedConditions = jp.query(resource.status, '$..conditions[?(@.type=="FetchFailed")]');
      const fetchFailedCondition = fetchFailedConditions.length === 1 ? fetchFailedConditions[0] : undefined
      const fetchFailed = fetchFailedCondition && fetchFailedCondition.status === "True"

      if (resource.kind === 'GitRepository' || resource.kind === "OCIRepository" || resource.kind === "Bucket") {
        return fetchFailed
      }

      if (ready || ((reconciling || dependencyNotReady) && !stalled)) {
        return false;
      } else {
        return true;
      }
    })
  }

  return filteredResources;
}
