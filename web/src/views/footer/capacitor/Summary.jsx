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

import jp from 'jsonpath'

export function Summary(props) {
  const { resources, label, simple } = props;

  if (!resources) {
    return null;
  }

  const totalCount = resources.length
  const readyCount = resources.filter(resourece => {
    const readyConditions = jp.query(resourece.status, '$..conditions[?(@.type=="Ready")]');
    const ready = readyConditions.length === 1 && readyConditions[0].status === "True"
    return ready
  }).length
  const dependencyNotReadyCount = resources.filter(resourece => {
    const readyConditions = jp.query(resourece.status, '$..conditions[?(@.type=="Ready")]');
    const dependencyNotReady = readyConditions.length === 1 && readyConditions[0].reason === "DependencyNotReady"
    return dependencyNotReady
  }).length
  const reconcilingCount = resources.filter(resourece => {
    const readyConditions = jp.query(resourece.status, '$..conditions[?(@.type=="Reconciling")]');
    const ready = readyConditions.length === 1 && readyConditions[0].status === "True"
    return ready
  }).length
  const stalledCount = resources.filter(resourece => {
    const readyConditions = jp.query(resourece.status, '$..conditions[?(@.type=="Ready")]');
    const ready = readyConditions.length === 1 && readyConditions[0].status === "True"
    if (ready) {
      return false
    }

    const readyTransitionTime = readyConditions.length === 1 ? readyConditions[0].lastTransitionTime : undefined
    const parsed = Date.parse(readyTransitionTime, "yyyy-MM-dd'T'HH:mm:ss");

    const fiveMinutesAgo = new Date();
    fiveMinutesAgo.setMinutes(fiveMinutesAgo.getMinutes() - 5);
    const stalled = fiveMinutesAgo > parsed

    return stalled
  }).length

  const ready = readyCount === totalCount
  const reconciling = reconcilingCount > 0 || dependencyNotReadyCount > 0
  const stalled = stalledCount > 0
  const readyLabel = ready ? "Ready" : reconciling && !stalled ? "Reconciling" : "Error"
  const color = ready ? "bg-teal-400 dark:bg-teal-600" : reconciling && !stalled ? "bg-blue-400 animate-pulse" : "bg-orange-400 animate-pulse"

  if (simple) {
    return (
      <span className='relative ml-4'>
        <span
          className={`absolute -left-4 top-1 rounded-full h-3 w-3 ${color} inline-block`}
          title={`${label}: ${readyLabel}(${readyCount}/${totalCount})`}
        />
      </span>
    )
  } 

  return (
    <>
      <div>
        <span className="font-bold text-neutral-700">{label}:</span>
        <span className='relative text-neutral-700 ml-5'>
          <span className={`absolute -left-4 top-1 rounded-full h-3 w-3 ${color} inline-block`}></span>
          <span>{readyLabel}</span>
        </span>
        <span>({readyCount}/{totalCount})</span>
      </div>
    </>
  )
}
