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

import { format } from "date-fns";
import { TimeLabel } from './TimeLabel'

export function ArtifactWidget(props) {
  const { gitRepository } = props
  const artifact = gitRepository.status.artifact

  const revision = artifact.revision
  const hash = revision.slice(revision.indexOf(':') + 1);
  const url = gitRepository.spec.url.slice(gitRepository.spec.url.indexOf('@') + 1)
  const branch = gitRepository.spec.ref.branch

  const parsed = Date.parse(artifact.lastUpdateTime, "yyyy-MM-dd'T'HH:mm:ss");
  const exactDate = format(parsed, 'MMMM do yyyy, h:mm:ss a O')

  return (
    <>
      <div className="capacitorField font-medium text-neutral-700 dark:text-neutral-200">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512" className="h4 w-4 inline fill-current"><path d="M320 336a80 80 0 1 0 0-160 80 80 0 1 0 0 160zm156.8-48C462 361 397.4 416 320 416s-142-55-156.8-128H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H163.2C178 151 242.6 96 320 96s142 55 156.8 128H608c17.7 0 32 14.3 32 32s-14.3 32-32 32H476.8z" /></svg>
        <span className="pl-1">
          <a href={`https://${url}/commit/${hash}`} target="_blank" rel="noopener noreferrer">
            {hash.slice(0, 8)} committed <TimeLabel title={exactDate} date={parsed} />
          </a>
        </span>
      </div>
      <span className="block capacitorField text-neutral-600 dark:text-neutral-400">
        <span className='font-mono bg-neutral-100 dark:bg-neutral-700 text-neutral-700 dark:text-neutral-300 px-1 rounded'>{branch}</span>
        <span className='px-1'>@</span>
        <a href={`https://${url}`} target="_blank" rel="noopener noreferrer">{url}</a>
      </span>
    </>
  )
}
