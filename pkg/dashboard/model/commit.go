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

package model

// Commit represents a Github commit
type Commit struct {
	// ID for this commit
	// required: true
	ID        int64          `json:"id" meddler:"id,pk"`
	Repo      string         `json:"-" meddler:"repo"`
	SHA       string         `json:"sha"  meddler:"sha"`
	URL       string         `json:"url"  meddler:"url"`
	Author    string         `json:"author"  meddler:"author"`
	AuthorPic string         `json:"author_pic"  meddler:"author_pic"`
	Tags      []string       `json:"tags,omitempty"    meddler:"tags,json"`
	Status    CombinedStatus `json:"status,omitempty"    meddler:"status,json"`
	Message   string         `json:"message"  meddler:"message"`
	Created   int64          `json:"created"  meddler:"created"`
}

type CombinedStatus struct {
	State    string   `json:"state"`
	Contexts []Status `json:"statuses"`
}

type Status struct {
	Context     string `json:"context"`
	CreatedAt   string `json:"createdAt"`
	State       string `json:"state"`
	TargetUrl   string `json:"targetURL"`
	Description string `json:"description"`
}
