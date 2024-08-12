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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/HelmRelease.jsx
*/

import React, { useState, useRef, useEffect } from 'react';
import { ReadyWidget } from './ReadyWidget'
import { HelmRevisionWidget } from './HelmRevisionWidget';

export function HelmRelease(props) {
  const { capacitorClient, item, targetReference, handleNavigationSelect } = props;
  const ref = useRef(null);
  const [highlight, setHighlight] = useState(false)

  useEffect(() => {
    const matching = targetReference.objectNs === item.metadata.namespace && targetReference.objectName === item.metadata.name
    setHighlight(matching);
    if (matching) {
      ref.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [item.metadata, targetReference]);

  return (
    <div
      ref={ref}
      className={(highlight ? "ring-2 ring-indigo-600 ring-offset-2" : "") + " card p-4 grid grid-cols-12 gap-x-4"}
      key={`hr-${item.metadata.namespace}/${item.metadata.name}`}
    >
      <div className="col-span-2">
        <span className="block font-medium">
          {item.metadata.name}
        </span>
        <span className="block text-neutral-600 dark:text-neutral-400">
          {item.metadata.namespace}
        </span>
      </div>
      <div className="col-span-4">
        <span className="block"><ReadyWidget resource={item} displayMessage={true} label="Reconciled" /></span>
      </div>
      <div className="col-span-5">
        <div className="font-medium text-neutral-700 dark:text-neutral-200"><HelmRevisionWidget helmRelease={item} withHistory={true} handleNavigationSelect={handleNavigationSelect} /></div>
      </div>
      <div className="grid-cols-1 text-right">
        <button className="transparentBtn !px-2"
          onClick={() => capacitorClient.reconcile("helmrelease", item.metadata.namespace, item.metadata.name)}
        >
          Reconcile
        </button>
      </div>
    </div>)
}
