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
import React, { useState } from 'react';
import { SkeletonLoader } from './skeletonLoader'
import { Modal } from './modal'
import { ACTION_TYPE_CLEAR_DEPLOYMENT_DETAILS } from '../../redux/redux';

export function Describe(props) {
  const { gimletClient, store, namespace, deployment, pods } = props;
  const [showModal, setShowModal] = useState(false)
  const dep = namespace + "/" + deployment;
  const [details, setDetails] = useState(store.getState().details[dep]);
  store.subscribe(() => setDetails(store.getState().details[dep]));

  const describeDeployment = () => {
    gimletClient.deploymentDetailsRequest(namespace, deployment)
  }

  // TODO handle show pod describe
  const describePod = (podNamespace, podName) => {
    gimletClient.podDetailsRequest(podNamespace, podName)
  }

  const closeDetailsHandler = () => {
    setShowModal(false)
    store.dispatch({
      type: ACTION_TYPE_CLEAR_DEPLOYMENT_DETAILS, payload: {
        deployment: namespace + "/" + deployment
      }
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
            {details ?? <SkeletonLoader />}
          </code>
        </Modal>
      }
      <button onClick={() => {
        setShowModal(true);
        describeDeployment()
      }}
        className="bg-transparent hover:bg-neutral-100 font-medium text-sm text-neutral-700 py-1 px-4 border border-neutral-300 rounded">
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
          // TODO
          if (!pod.metadata) {
            return (
              <button
                key={pod.name}
                title={pod.name}
                className={`${pod.name === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
                onClick={() => {
                  describePod(pod.namespace, pod.name);
                  setSelected(pod.name)
                }}
              >
                {pod.name}
              </button>)
          }

          return (
            <button
              key={pod.metadata.name}
              title={pod.metadata.name}
              className={`${pod.metadata.name === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
              onClick={() => {
                describePod(pod.metadata.namespace, pod.metadata.name);
                setSelected(pod.metadata.name)
              }}
            >
              {pod.metadata.name}
            </button>)
        })
      }
    </div>
  )
}
