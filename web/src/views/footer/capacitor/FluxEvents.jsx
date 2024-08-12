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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/FluxEvents.jsx
*/

import { NavigationButton } from './NavigationButton'
import { TimeLabel } from './TimeLabel'
import { format } from "date-fns";
import { useState } from 'react';

function FluxEvents(props) {
  const { events, handleNavigationSelect } = props
  const [filter, setFilter] = useState(false)

  let filteredEvents = events;
  if (filter) {
    filteredEvents = filteredEvents.filter(e => e.type === "Warning")
  }

  return (
    <div className="space-y-4">
      <button className={(filter ? "text-blue-50 bg-blue-500 dark:bg-blue-800" : "bg-neutral-50 dark:bg-neutral-600 text-neutral-600 dark:text-neutral-200") + " rounded-full px-3"}
        onClick={() => setFilter(!filter)}
      >
        Filter errors
      </button>
      <div className="card flow-root p-4">
        <div className="overflow-x-auto">
          <div className="inline-block min-w-full py-2 align-middle">
            <table className="min-w-full divide-y divide-neutral-300 dark:divide-neutral-500">
              <thead>
                <tr>
                  <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold sm:pl-0">
                    Last Seen
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold">
                    Object
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold">
                    Type
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold">
                    Reason
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold">
                    Message
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-200 dark:divide-neutral-600">
                {filteredEvents.map((e, index) => {
                  return (
                  <tr key={index} className={e.type === "Warning" ? "bg-orange-400 dark:bg-" : ""}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium sm:pl-0">
                      <LastSeen event={e} />
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-neutral-600 dark:text-neutral-400">
                      <NavigationButton handleNavigation={() => handleNavigationSelect(e.involvedObjectKind === "Kustomization" ? "Kustomizations" : "Sources", e.involvedObjectNamespace, e.involvedObject, e.involvedObjectKind)}>
                        {e.involvedObjectKind}: {e.involvedObjectNamespace}/{e.involvedObject}
                      </NavigationButton>
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-neutral-600 dark:text-neutral-400">{e.type}</td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-neutral-600 dark:text-neutral-400">{e.reason}</td>
                    <td className="px-3 py-4 text-sm text-neutral-600 dark:text-neutral-400">{e.message}</td>
                  </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}

function LastSeen(props) {
  const { event } = props

  const firstTimestampSince = event.eventTime !== "0001-01-01T00:00:00Z" ? event.eventTime : event.firstTimestamp
  const firstTimestampSinceParsed = Date.parse(firstTimestampSince, "yyyy-MM-dd'T'HH:mm:ss");
  const firstTimestampSinceExactDate = format(firstTimestampSinceParsed, 'MMMM do yyyy, h:mm:ss a O')

	if (event.series) {
    const lastObservedTimeParsed = Date.parse(event.series.lastObservedTime, "yyyy-MM-dd'T'HH:mm:ss");
    const lastObservedTimeExactDate = format(lastObservedTimeParsed, 'MMMM do yyyy, h:mm:ss a O')
	  return (
      <span>
        <TimeLabel title={lastObservedTimeExactDate} date={lastObservedTimeParsed} />
        <span className='px-1'>ago (x{event.series.count} over</span>
        <TimeLabel title={firstTimestampSinceExactDate} date={firstTimestampSinceParsed} />
        )
      </span>
    )
	} else if (event.count > 1) {
    const lastTimestampParsed = Date.parse(event.lastTimestamp, "yyyy-MM-dd'T'HH:mm:ss");
    const lastTimestampExactDate = format(lastTimestampParsed, 'MMMM do yyyy, h:mm:ss a O')
    return (
      <span>
        <TimeLabel title={lastTimestampExactDate} date={lastTimestampParsed} />
        <span className='px-1'>ago (x{event.count} over</span>
        <TimeLabel title={firstTimestampSinceExactDate} date={firstTimestampSinceParsed} />
        )
      </span>
    )
	} else {
		return (<span><TimeLabel title={firstTimestampSinceExactDate} date={firstTimestampSinceParsed} /> ago</span>)
	}
}

export default FluxEvents;
