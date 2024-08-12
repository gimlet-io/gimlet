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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/SkeletonLoader.jsx
*/

export const SkeletonLoader = () => {
  return (
    <div className="w-full max-w-4xl animate-pulse space-y-3">
      <div className="h-2 bg-neutral-700 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-3/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-4/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-4/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-3/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-1/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-1/6"></div>
      <div className="h-2 bg-neutral-700 rounded w-2/5"></div>
      <div className="h-2 bg-neutral-700 rounded w-3/5"></div>
    </div>
  )
}
