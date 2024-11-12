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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/ReadyWidget.jsx
*/

import jp from 'jsonpath'
import { format } from "date-fns";
import { TimeLabel } from './TimeLabel'

export function ReadyWidget(props) {
  const { resource, displayMessage, label } = props

  const readyConditions = jp.query(resource.status, '$..conditions[?(@.type=="Ready")]');
  const readyCondition = readyConditions.length === 1 ? readyConditions[0] : undefined
  const ready = readyCondition?.status === "True"

  const dependencyNotReady = readyCondition && readyCondition.reason === "DependencyNotReady"

  const readyTransitionTime = readyCondition ? readyCondition.lastTransitionTime : undefined
  const parsed = Date.parse(readyTransitionTime, "yyyy-MM-dd'T'HH:mm:ss");
  const exactDate = format(parsed, 'MMMM do yyyy, h:mm:ss a O')
  const fiveMinutesAgo = new Date();
  fiveMinutesAgo.setMinutes(fiveMinutesAgo.getMinutes() - 5);
  const stalled = fiveMinutesAgo > parsed

  const reconcilingConditions = jp.query(resource.status, '$..conditions[?(@.type=="Reconciling")]');
  const reconcilingCondition = reconcilingConditions.length === 1 ? reconcilingConditions[0] : undefined
  let reconciling = reconcilingCondition?.status === "True"
  if (resource.kind === 'Terraform') {
    const planConditions = jp.query(resource.status, '$..conditions[?(@.type=="Plan")]');
    const planCondition = planConditions.length === 1 ? planConditions[0] : undefined
    const planning = planCondition?.status === "False"

    const applyConditions = jp.query(resource.status, '$..conditions[?(@.type=="Apply")]');
    const applyCondition = applyConditions.length === 1 ? applyConditions[0] : undefined
    const applying = applyCondition?.status === "False"

    reconciling = !ready && (planning || applying || (!planCondition && !applyCondition))
  }

  const fetchFailedConditions = jp.query(resource.status, '$..conditions[?(@.type=="FetchFailed")]');
  const fetchFailedCondition = fetchFailedConditions.length === 1 ? fetchFailedConditions[0] : undefined
  const fetchFailed = fetchFailedCondition && fetchFailedCondition.status === "True"

  var [color,statusLabel,messageColor] = ['','','']
  const readyLabel = label ? label : "Ready"
  if (resource.kind === 'GitRepository' || resource.kind === "OCIRepository" || resource.kind === "Bucket") {
    color = fetchFailed ? "bg-orange-400 animate-pulse" : reconciling ? "bg-blue-400 animate-pulse" : ready ? "bg-green-300 dark:bg-teal-600" : "bg-orange-400 animate-pulse"
    statusLabel = fetchFailed ? "Error" : reconciling ?  "Reconciling" : ready ? readyLabel : "Error"
    messageColor = fetchFailed ? "bg-orange-400" : reconciling ?  "" : ready ? "capacitorField" : "bg-orange-400"
  } else {
    color = ready ? "bg-green-300 dark:bg-teal-600" : (reconciling || dependencyNotReady) && !stalled ? "bg-blue-400 animate-pulse" : "bg-orange-400 animate-pulse"
    statusLabel = ready ? readyLabel : (reconciling || dependencyNotReady) && !stalled ? "Reconciling" : "Error"
    messageColor = ready ? "capacitorField" : (reconciling || dependencyNotReady) && !stalled ? "" : "bg-orange-400"
  }

  return (
    <div className="relative">
      <div className="font-medium">
        <span className={`absolute -left-4 top-1 rounded-full h-3 w-3 ${color} inline-block`}></span>
        {label &&
        <>
          <span>{statusLabel}</span>
          {readyCondition &&
            <span className='ml-1'><TimeLabel title={exactDate} date={parsed} /> ago</span>
          }
        </>
        }
      </div>
      {displayMessage &&
        <div className={`${messageColor} text-neutral-600 dark:text-neutral-400`}>
          {reconciling && reconcilingCondition &&
            <span title={reconcilingCondition.message}>{reconcilingCondition.message}</span>
          }
          {dependencyNotReady &&
            <span>Dependency not ready</span>
          }
          {readyCondition &&
          <span title={readyCondition.message}>{readyCondition.message}</span>
          }
        </div>
      }
    </div>
  )
}
