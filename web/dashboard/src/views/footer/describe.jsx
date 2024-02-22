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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/Describe.jsx
*/

import React, { useState } from 'react';
import { SkeletonLoader } from './SkeletonLoader'
import { Modal } from './Modal'
import { ACTION_TYPE_CLEAR_DETAILS } from '../../redux/redux';

export function Describe(props) {
  const { gimletClient, store, namespace, deployment, pods } = props;
  const [showModal, setShowModal] = useState(false)
  const [selected, setSelected] = useState(namespace + "/" + deployment)
  const [details, setDetails] = useState(store.getState().details);
  store.subscribe(() => setDetails(store.getState().details));

  const describeDeployment = () => {
    gimletClient.deploymentDetailsRequest(namespace, deployment)
    setSelected(namespace + "/" + deployment)
  }

  const describePod = (podNamespace, podName) => {
    gimletClient.podDetailsRequest(podNamespace, podName)
    setSelected(podNamespace + "/" + podName)
  }

  const closeDetailsHandler = () => {
    setShowModal(false)
    store.dispatch({
      type: ACTION_TYPE_CLEAR_DETAILS
    });
  }

  return (
    <>
      {showModal &&
        <Modal
          stopHandler={closeDetailsHandler}
          navBar={
            <DescribeNav
              deployment={deployment}
              pods={pods}
              describeDeployment={describeDeployment}
              describePod={describePod}
            />}
        >
          <code className='flex whitespace-pre items-center font-mono text-xs p-2 text-yellow-100 rounded'>
            {details[selected] ?? <SkeletonLoader />}
          </code>
        </Modal>
      }
      <button onClick={() => {
        setShowModal(true);
        describeDeployment()
      }}
        className="bg-transparent hover:bg-neutral-100 font-medium text-sm text-neutral-700 py-1 px-4 mr-2 border border-neutral-300 rounded">
        Describe
      </button>
    </>
  )
}

function DescribeNav(props) {
  const { deployment, pods, describeDeployment, describePod } = props;
  const [selected, setSelected] = useState(deployment)

  return (
    <div className="flex flex-wrap items-center overflow-auto mx-4 space-x-1">
      <button
        title={deployment}
        className={`${deployment === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
        onClick={() => {
          describeDeployment();
          setSelected(deployment)
        }}
      >
        Deployment
      </button>
      {
        pods?.map((pod) => {
          const podNamespace = pod.metadata ? pod.metadata.namespace : pod.namespace;
          const podName = pod.metadata ? pod.metadata.name : pod.name;

          return (
            <button
              key={podName}
              title={podName}
              className={`${podName === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
              onClick={() => {
                describePod(podNamespace, podName);
                setSelected(podName)
              }}
            >
              {podName}
            </button>)
        })
      }
    </div>
  )
}
