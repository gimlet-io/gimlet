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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/Logs.jsx
*/

import React, { useState } from 'react';
import { ACTION_TYPE_CLEAR_PODLOGS } from '../../redux/redux';
import { Modal } from './capacitor/Modal'
import { SkeletonLoader } from './capacitor/SkeletonLoader'

export function Logs(props) {
  const { capacitorClient, store, namespace, deployment, containers } = props;
  const [showModal, setShowModal] = useState(false)
  const deploymentName = namespace + "/" + deployment
  const [logs, setLogs] = useState(store.getState().podLogs[deploymentName])
  store.subscribe(() => setLogs(store.getState().podLogs[deploymentName]));
  const [selected, setSelected] = useState("")

  const streamPodLogs = () => {
    capacitorClient.podLogsRequest(namespace, deployment)
  }

  const stopLogsHandler = () => {
    setShowModal(false);
    capacitorClient.stopPodlogsRequest(namespace, deployment);
    store.dispatch({
      type: ACTION_TYPE_CLEAR_PODLOGS, payload: {
        pod: namespace + "/" + deployment
      }
    });
  }

  return (
    <>
      {showModal &&
        <Modal
          stopHandler={stopLogsHandler}
          navBar={
            <LogsNav
              containers={containers}
              selected={selected}
              setSelected={setSelected}
            />
          }
        >
          {logs ?
            logs.filter(line => line.pod.includes(selected)).map((line, idx) => <p key={idx} className={`font-mono text-xs ${line.color}`}>{line.content}</p>)
            :
            <SkeletonLoader />
          }
        </Modal>
      }
      <button onClick={() => {
        setShowModal(true);
        streamPodLogs()
      }}
        className="transparentBtn">
        Logs
      </button>
    </>
  )
}

function LogsNav(props) {
  const { containers, selected, setSelected } = props;

  return (
    <div className="flex flex-wrap items-center overflow-auto mx-4 space-x-1">
      <button
        className={`${"" === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
        onClick={() => {
          setSelected("")
        }}
      >
        All pods
      </button>
      {
        containers?.map((container) => (
          <button
            key={container}
            title={container}
            className={`${container === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
            onClick={() => {
              setSelected(container)
            }}
          >
            {container}
          </button>
        ))
      }
    </div>
  )
}
