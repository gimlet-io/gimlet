package dx

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	cueyaml "cuelang.org/go/encoding/yaml"
)

func RenderCueToManifests(fileContent string) ([]string, error) {
	c := cuecontext.New()
	v := c.CompileString(fileContent)

	err := v.Validate()
	if err != nil {
		msg := errors.Details(err, nil)
		return []string{}, fmt.Errorf("cannot parse cue file: %s", msg)
	}

	configs := v.LookupPath(cue.ParsePath("configs"))
	if !configs.Exists() {
		return []string{}, fmt.Errorf("cue files should have a `configs` field that holds an array of Gimlet manfiests")
	}

	var manifests []string

	iter, _ := configs.List()
	for iter.Next() {
		m, err := cueyaml.Encode(iter.Value())
		if err != nil {
			return []string{}, err
		}
		manifests = append(manifests, string(m))
	}

	return manifests, nil
}
