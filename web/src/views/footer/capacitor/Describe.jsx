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

import React, { useState, useEffect } from 'react';
import { SkeletonLoader } from './SkeletonLoader'
import { Modal } from './Modal'
import { ACTION_TYPE_CLEAR_DETAILS } from '../../../redux/redux';

export function Describe(props) {
  const { capacitorClient, store, resource, namespace, name, deployment, pods } = props;
  const [showModal, setShowModal] = useState(false)
  const [selected, setSelected] = useState("")
  const [content, setContent] = useState()
  store.subscribe(() => {
    console.log(store.getState().detail)
    setContent(store.getState().details[selected])
  });

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
              resource={resource}
              namespace={namespace}
              name={name}
              deployment={deployment}
              pods={pods}
              capacitorClient={capacitorClient}
              selected={selected}
              setSelected={setSelected}
            />}
        >
          <code key={selected} className='text-left flex whitespace-pre font-mono text-xs p-2 text-yellow-100 rounded'>
            {content ?? <SkeletonLoader />}
          </code>
        </Modal>
      }
      <button onClick={() => {
        setShowModal(true);
      }}
        className="transparentBtn w-24">
        Describe
      </button>
    </>
  )
}

function DescribeNav(props) {
  const { capacitorClient, selected, setSelected } = props;
  const { resource, namespace, name, deployment, pods } = props;

  const describeResource = (r, ns, n) => {
    capacitorClient.describe(r, ns, n)
    setSelected(`${r}/${ns}/${n}`)
  }

  useEffect(() => {
    if (resource) {
      setSelected(`${resource}/${namespace}/${name}`)
      describeResource(resource, namespace, name);
    } else {
      setSelected(deployment)
      describeResource("Deployment", namespace, deployment);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex flex-wrap items-center overflow-auto mx-4 space-x-1">
      {deployment &&
      <button
        title={deployment}
        className={`${deployment === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
        onClick={() => {
          setSelected(deployment)
          describeResource("Deployment", namespace, deployment);
        }}
      >
        Deployment
      </button>
      }
      {pods?.map((pod) => {
          const podNamespace = pod.metadata ? pod.metadata.namespace : pod.namespace;
          const podName = pod.metadata ? pod.metadata.name : pod.name;

          return (
            <button
              key={podName}
              title={podName}
              className={`${podName === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
              onClick={() => {
                setSelected(podName)
                describeResource("Pod", podNamespace, podName);
              }}
            >
              {podName}
            </button>)
        })
      }
      {resource &&
      <button
        title={`${resource}/${namespace}/${name}`}
        className={`${`${resource}/${namespace}/${name}` === selected ? 'bg-white' : 'hover:bg-white bg-neutral-300'} my-2 inline-block rounded-full py-1 px-2 font-medium text-xs leading-tight text-neutral-700`}
        onClick={() => {
          setSelected(`${resource}/${namespace}/${name}`)
          describeResource(resource, namespace, name);
        }}
      >
        {`${resource}/${namespace}/${name}`}
      </button>
      }
    </div>
  )
}
