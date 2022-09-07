package dx

import (
	"bytes"
	"encoding/json"
	"errors"

	"gopkg.in/yaml.v3"
)

// GitEvent represents the git event that produced the artifact
type GitEvent int

const (
	// Push artifact is produced by a git push event
	Push GitEvent = iota
	// Tag artifact is produced by a git tag event
	Tag
	// PR artifact is produced by a pull request event
	PR
)

func (s GitEvent) String() string {
	return toString[s]
}

func GitEventFromString(eventString string) (*GitEvent, error) {
	if event, ok := toID[eventString]; ok {
		return &event, nil
	}
	return nil, errors.New("wrong input")
}

var toString = map[GitEvent]string{
	Push: "push",
	Tag:  "tag",
	PR:   "pr",
}

var toID = map[string]GitEvent{
	"push": Push,
	"tag":  Tag,
	"pr":   PR,
}

// MarshalJSON marshals the enum as a quoted json string
func (s GitEvent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshalls a quoted json string to the enum value
func (s *GitEvent) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Push' in this case.
	*s = toID[j]
	return nil
}

// MarshalYAML marshals the enum as a quoted yaml string
func (s GitEvent) MarshalYAML() (interface{}, error) {
	return toString[s], nil
}

// UnmarshalYAML unmarshalls a quoted yaml string to the enum value
func (s *GitEvent) UnmarshalYAML(n *yaml.Node) error {
	var j string
	err := yaml.Unmarshal([]byte(n.Value), &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Push' in this case.
	*s = toID[j]
	return nil
}

func PushPtr() *GitEvent {
	push := Push
	return &push
}

func TagPtr() *GitEvent {
	tag := Tag
	return &tag
}

func PRPtr() *GitEvent {
	pr := PR
	return &pr
}
