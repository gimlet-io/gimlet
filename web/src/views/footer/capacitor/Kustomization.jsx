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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/Kustomization.jsx
*/

import React, { useState, useEffect, useMemo, useRef } from 'react';
import { ReadyWidget } from './ReadyWidget'
import jp from 'jsonpath'
import { NavigationButton } from './NavigationButton'
import { findSource } from './utils';

export function Kustomization(props) {
  const { capacitorClient, item, fluxState, targetReference, handleNavigationSelect } = props;
  const ref = useRef(null);
  const [highlight, setHighlight] = useState(false)

  useEffect(() => {
    const matching = targetReference.objectNs === item.metadata.namespace && targetReference.objectName === item.metadata.name
    setHighlight(matching);
    if (matching) {
      ref.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [item.metadata.name, item.metadata.namespace, targetReference]);

  const sources = useMemo(() => {
    const sources = [];
    if (fluxState.ociRepositories) {
      sources.push(...fluxState.ociRepositories)
      sources.push(...fluxState.gitRepositories)
      sources.push(...fluxState.buckets)
    }
    return [...sources].sort((a, b) => a.metadata.name.localeCompare(b.metadata.name));
  }, [fluxState]);
  const source = findSource(sources, item)

  return (
    <div
      className={(highlight ? "ring-2 ring-indigo-600 ring-offset-2" : "") + " card p-4 grid grid-cols-12 gap-x-4"}
      key={`${item.metadata.namespace}/${item.metadata.name}`}
    >
      <div className="col-span-2">
        <span className="block font-medium">
          {item.metadata.name}
        </span>
        <span className="block text-neutral-600 dark:text-neutral-400">
          {item.metadata.namespace}
        </span>
      </div>
      <div className="col-span-4 text-base">
        <span className="block"><ReadyWidget resource={item} displayMessage={true} label="Applied" /></span>
      </div>
      <div className="col-span-5">
        <div className="text-base font-medium field">
          <RevisionWidget
            kustomization={item}
            source={source}
            handleNavigationSelect={handleNavigationSelect}
            inFooter={true}
          />
        </div>
        { source.kind !== 'OCIRepository' &&
        <span className='rounded font-mono bg-neutral-100 dark:bg-neutral-700 px-1'>{item.spec.path}</span>
        }
      </div>
      <div className="grid-cols-1 text-right">
        <button className="transparentBtn !px-2"
          onClick={() => capacitorClient.reconcile("kustomization", item.metadata.namespace, item.metadata.name)}
        >
          Reconcile
        </button>
      </div>
    </div>
  )
}

export function RevisionWidget(props) {
  const { kustomization, source, handleNavigationSelect, inFooter } = props

  const appliedRevision = kustomization.status.lastAppliedRevision
  const appliedHash = appliedRevision ? appliedRevision.slice(appliedRevision.indexOf(':') + 1) : "";

  const lastAttemptedRevision = kustomization.status.lastAttemptedRevision
  const lastAttemptedHash = lastAttemptedRevision ? lastAttemptedRevision.slice(lastAttemptedRevision.indexOf(':') + 1) : "";

  const readyConditions = jp.query(kustomization.status, '$..conditions[?(@.type=="Ready")]');
  const readyCondition = readyConditions.length === 1 ? readyConditions[0] : undefined
  const ready = readyCondition && readyConditions[0].status === "True"

  const readyTransitionTime = readyCondition ? readyCondition.lastTransitionTime : undefined
  const parsed = Date.parse(readyTransitionTime, "yyyy-MM-dd'T'HH:mm:ss");
  const fiveMinutesAgo = new Date();
  fiveMinutesAgo.setMinutes(fiveMinutesAgo.getMinutes() - 5);
  const stalled = fiveMinutesAgo > parsed

  // const reconcilingConditions = jp.query(kustomization.status, '$..conditions[?(@.type=="Reconciling")]');
  // const reconcilingCondition = reconcilingConditions.length === 1 ? reconcilingConditions[0] : undefined
  // const reconciling = reconcilingCondition && reconcilingConditions[0].status === "True"

  const url = source.spec.url.slice(source.spec.url.indexOf('@') + 1)

  const navigationHandler = inFooter ?
    () => handleNavigationSelect("Sources", source.metadata.namespace, source.metadata.name, source.kind) :
    () => handleNavigationSelect("Kustomizations", kustomization.metadata.namespace, kustomization.metadata.name)

  return (
    <>
    { !ready && stalled &&
      <span className='bg-orange-400'>
        <span>Last Attempted: </span>
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512" className="h4 w-4 inline fill-current"><path d="M320 336a80 80 0 1 0 0-160 80 80 0 1 0 0 160zm156.8-48C462 361 397.4 416 320 416s-142-55-156.8-128H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H163.2C178 151 242.6 96 320 96s142 55 156.8 128H608c17.7 0 32 14.3 32 32s-14.3 32-32 32H476.8z"/></svg>
        <span className="pl-1">
          <a href={`https://${url}/commit/${lastAttemptedHash}`} target="_blank" rel="noopener noreferrer">
            {lastAttemptedHash.slice(0, 8)}
          </a>
        </span>
        <NavigationButton handleNavigation={navigationHandler}>
          &nbsp;({`${source.metadata.namespace}/${source.metadata.name}`})
        </NavigationButton>
      </span>
    }
    <span className={`${ready ? '' : 'font-normal'} field`}>
      { !ready &&
      <span>Currently Applied: </span>
      }
      { source.kind === 'OCIRepository' &&
      <NavigationButton handleNavigation={navigationHandler}>
       {appliedRevision}
       <div className='text-left'>({`${source.metadata.namespace}/${source.metadata.name}`})</div>
      </NavigationButton>
      }
      { source.kind !== 'OCIRepository' &&
      <>
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512" className="h4 w-4 inline fill-current"><path d="M320 336a80 80 0 1 0 0-160 80 80 0 1 0 0 160zm156.8-48C462 361 397.4 416 320 416s-142-55-156.8-128H32c-17.7 0-32-14.3-32-32s14.3-32 32-32H163.2C178 151 242.6 96 320 96s142 55 156.8 128H608c17.7 0 32 14.3 32 32s-14.3 32-32 32H476.8z"/></svg>
        <span className="pl-1">
          <a href={`https://${url}/commit/${appliedHash}`} target="_blank" rel="noopener noreferrer">
            {appliedHash.slice(0, 8)}
          </a>
        </span>
        <NavigationButton handleNavigation={navigationHandler}>
          &nbsp;({`${source.metadata.namespace}/${source.metadata.name}`})
        </NavigationButton>
      </>
      }
    </span>
    </>
  )
}
