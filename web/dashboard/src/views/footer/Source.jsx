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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/Source.jsx
*/

import React, { useState, useEffect, useRef } from 'react';
import { ReadyWidget } from './ReadyWidget'
import { ArtifactWidget } from './ArtifactWidget';
import { OCIArtifactWidget } from './OCIArtifactWidget';

export function Source(props) {
  const { capacitorClient, source, targetReference } = props;
  const ref = useRef(null);
  const [highlight, setHighlight] = useState(false)

  useEffect(() => {
    const matching = targetReference.objectNs === source.metadata.namespace &&
      targetReference.objectName === source.metadata.name &&
      targetReference.objectKind === source.kind
    setHighlight(matching);
    if (matching) {
      ref.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [source.metadata, source.kind, targetReference]);

  return (
    <div
      ref={ref}
      className={(highlight ? "ring-2 ring-indigo-600 ring-offset-2" : "") + " rounded-md border border-neutral-300 p-4 grid grid-cols-12 gap-x-4 bg-white shadow"}
      key={`${source.metadata.namespace}/${source.metadata.name}`}
    >
      <div className="col-span-2">
        <span className="block font-medium text-black">
          {source.metadata.name}
        </span>
        <span className="block text-neutral-600">
          {source.metadata.namespace}
        </span>
      </div>
      <div className="col-span-4">
        <ReadyWidget resource={source} displayMessage={true} />
      </div>
      <div className="col-span-5">
        {source.kind === 'GitRepository' &&
          <ArtifactWidget gitRepository={source} displayMessage={true} />
        }
        {source.kind === 'OCIRepository' &&
          <OCIArtifactWidget source={source} displayMessage={true} />
        }
        {source.kind === 'Bucket' &&
          <span>Bucket (TODO)</span>
        }
      </div>
      <div className="grid-cols-1 text-right">
        <button className="bg-transparent hover:bg-neutral-100 font-medium text-sm text-neutral-700 py-1 px-2 border border-neutral-300 rounded"
          onClick={() => capacitorClient.reconcile(source.kind, source.metadata.namespace, source.metadata.name)}
        >
          Reconcile
        </button>
      </div>
    </div>
  )
}