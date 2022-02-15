package dx

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v3"
)

// CleanupEvent represents events that cause an app instance cleanup
type CleanupEvent int

const (
	// BranchDeleted indicates if a git branch is deleted
	BranchDeleted CleanupEvent = iota
)

func (s CleanupEvent) String() string {
	return cleanupEventToString[s]
}

var cleanupEventToString = map[CleanupEvent]string{
	BranchDeleted: "branchDeleted",
}

var cleanupEventToID = map[string]CleanupEvent{
	"branchDeleted": BranchDeleted,
}

// MarshalJSON marshals the enum as a quoted json string
func (s CleanupEvent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(cleanupEventToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshalls a quoted json string to the enum value
func (s *CleanupEvent) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Push' in this case.
	*s = cleanupEventToID[j]
	return nil
}

// MarshalYAML marshals the enum as a quoted yaml string
func (s CleanupEvent) MarshalYAML() (interface{}, error) {
	return cleanupEventToString[s], nil
}

// UnmarshalYAML unmarshalls a quoted yaml string to the enum value
func (s *CleanupEvent) UnmarshalYAML(n *yaml.Node) error {
	var j string
	err := yaml.Unmarshal([]byte(n.Value), &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Push' in this case.
	*s = cleanupEventToID[j]
	return nil
}
