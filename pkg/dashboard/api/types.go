// Copyright 2019 Laszlo Fogas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Pod struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
	Logs              string `json:"logs"`
}

func (p *Pod) FQN() string {
	return p.Namespace + "/" + p.Name
}

type Deployment struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Pods          []*Pod `json:"pods,omitempty"`
	SHA           string `json:"sha"`
	CommitMessage string `json:"commitMessage"`
}

func (d *Deployment) FQN() string {
	return d.Namespace + "/" + d.Name
}

type Ingress struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	URL       string `json:"url"`
}

type Env struct {
	Name   string   `json:"name"`
	Stacks []*Stack `json:"stacks"`
}

type Stack struct {
	Repo       string      `json:"repo"`
	Env        string      `json:"env"`
	Service    *Service    `json:"service"`
	Deployment *Deployment `json:"deployment,omitempty"`
	Ingresses  []*Ingress  `json:"ingresses,omitempty"`
}

type StackUpdate struct {
	Event   string `json:"event"`
	Repo    string `json:"repo"`
	Env     string `json:"env"`
	Subject string `json:"subject"`
	Svc     string `json:"svc"`

	// Pod
	Status     string `json:"status"`
	Deployment string `json:"deployment"`
	ErrorCause string `json:"errorCause"`
	Logs       string `json:"logs"`

	// Deployment
	SHA           string `json:"sha"`
	CommitMessage string `json:"commitMessage"` // only used in streamed update to frontend

	// Ingress
	URL string `json:"url"`

	// Service
	Stacks []*Stack `json:"stacks"`
}

type Tag struct {
	SHA  string `json:"sha"`
	Name string `json:"name"`
}
