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

Original version: https://github.com/gimlet-io/capacitor/blob/main/web/src/Service.jsx
Trimmed everything that is not used
*/

export function Pod(props) {
  const {pod} = props;

  let textColor;
  let color;
  let pulsar;
  switch (pod.status.phase) {
    case 'Running':
      textColor = 'text-green-900 dark:text-teal-400'
      color = 'bg-green-300 dark:bg-teal-600';
      pulsar = '';
      break;
    case 'PodInitializing':
    case 'ContainerCreating':
    case 'Pending':
      textColor = 'text-blue-900'
      color = 'bg-blue-300';
      pulsar = 'animate-pulse';
      break;
    case 'Terminating':
      color = 'bg-neutral-500';
      pulsar = 'animate-pulse';
      break;
    default:
      textColor = 'text-neutral-900 dark:text-red-400'
      color = 'bg-red-600 dark:bg-red-800';
      pulsar = '';
      break;
  }

  return (
    <span className={`inline-block mr-1 mt-2 shadow-lg ${textColor} ${color} ${pulsar} font-bold px-2 cursor-default`} title={`${pod.metadata.name} - ${pod.status.phase}`}>
      {pod.metadata.name}
    </span>
  );
}
