package stack

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/epiclabs-io/diff3"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"gopkg.in/yaml.v3"
)

func GenerateAndWriteFiles(stackConfig dx.StackConfig, targetDir string) error {
	generatedFiles, err := GenerateFromStackYaml(stackConfig)
	if err != nil {
		return fmt.Errorf("cannot generate stack: %s", err.Error())
	}

	oldStackConfigPath := filepath.Join(filepath.Dir(targetDir), ".stack", "old")
	oldStackConfig, err := ReadStackConfig(oldStackConfigPath)
	if err != nil {
		oldStackConfig = stackConfig
	}
	previousGenerationFiles, err := GenerateFromStackYaml(oldStackConfig)
	if err != nil {
		return fmt.Errorf("cannot generate stack: %s", err.Error())
	}

	targetPath := filepath.Dir(targetDir)
	err = writeFilesAndPreserveCustomChanges(
		previousGenerationFiles,
		generatedFiles,
		targetPath,
	)
	if err != nil {
		return fmt.Errorf("cannot write stack: %s", err.Error())
	}

	err = keepStackConfigUsedForGeneration(targetDir, stackConfig)
	if err != nil {
		return fmt.Errorf("cannot write old stack config: %s", err.Error())
	}

	return nil
}

func ReadStackConfig(stackConfigPath string) (dx.StackConfig, error) {
	stackConfigYaml, err := ioutil.ReadFile(stackConfigPath)
	if err != nil {
		return dx.StackConfig{}, fmt.Errorf("cannot read stack config file: %s", err.Error())
	}

	var stackConfig dx.StackConfig
	err = yaml.Unmarshal(stackConfigYaml, &stackConfig)
	if err != nil {
		return dx.StackConfig{}, fmt.Errorf("cannot parse stack config file: %s", err.Error())
	}
	return stackConfig, nil
}

func writeFilesAndPreserveCustomChanges(
	previousGenerationFiles map[string]string,
	generatedFiles map[string]string,
	targetPath string,
) error {
	for path, updated := range generatedFiles { // write new or update existing files
		physicalPath := filepath.Join(targetPath, path)

		var existingContent string
		if _, err := os.Stat(physicalPath); err == nil {
			existingContentBytes, err := os.ReadFile(physicalPath)
			if err != nil {
				return fmt.Errorf("cannot read file %s: %s", path, err.Error())
			}
			existingContent = string(existingContentBytes)
		}

		var baseline string
		if val, ok := previousGenerationFiles[path]; ok {
			baseline = val
		}

		var mergedString string
		if existingContent != "" {
			merged, err := diff3.Merge(strings.NewReader(existingContent), strings.NewReader(baseline), strings.NewReader(updated), true, "Your custom settings", "From stack generate")
			if err != nil {
				return fmt.Errorf("cannot merge %s: %s", path, err.Error())
			}
			mergedBuffer := new(strings.Builder)
			_, err = io.Copy(mergedBuffer, merged.Result)
			if err != nil {
				return fmt.Errorf("cannot merge %s: %s", path, err.Error())
			}

			mergedString = mergedBuffer.String()
			if !strings.HasSuffix(mergedString, "\n") {
				mergedString = mergedString + "\n"
			}
		} else {
			mergedString = updated
		}

		err := os.MkdirAll(filepath.Dir(physicalPath), 0775)
		if err != nil {
			return fmt.Errorf("cannot write stack: %s", err.Error())
		}
		err = ioutil.WriteFile(physicalPath, []byte(mergedString), 0664)
		if err != nil {
			return fmt.Errorf("cannot write stack: %s", err.Error())
		}
	}

	for path := range previousGenerationFiles { // delete missing files
		if _, ok := generatedFiles[path]; !ok {
			physicalPath := filepath.Join(targetPath, path)
			err := os.Remove(physicalPath)
			if err != nil {
				return fmt.Errorf("cannot clean up file: %s", err.Error())
			}
		}
	}

	return nil
}

func keepStackConfigUsedForGeneration(
	stackConfigPath string,
	stackConfig dx.StackConfig,
) error {
	stackBackupPath := filepath.Join(filepath.Dir(stackConfigPath), ".stack", "old")
	err := os.MkdirAll(filepath.Dir(stackBackupPath), 0775)
	if err != nil {
		return err
	}
	return WriteStackConfig(stackConfig, stackBackupPath)
}

func WriteStackConfig(stackConfig dx.StackConfig, stackConfigPath string) error {
	updatedStackConfigBuffer := bytes.NewBufferString("")
	e := yaml.NewEncoder(updatedStackConfigBuffer)
	e.SetIndent(2)
	e.Encode(stackConfig)

	updatedStackConfigString := "---\n" + updatedStackConfigBuffer.String()
	return ioutil.WriteFile(stackConfigPath, []byte(updatedStackConfigString), 0666)
}
