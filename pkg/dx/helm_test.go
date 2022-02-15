package dx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SplitHelmOutput(t *testing.T) {

	templatedManifests := `
---
# Source: cron-job/templates/cronJob.yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
	name: myapp-first
	namespace: production
---
# Source: cron-job/templates/deployment.yaml
apiVersion: batch/v1beta1
kind: Deployment
metadata:
	name: myapp-second
	namespace: production
`

	files := SplitHelmOutput(map[string]string{"manifest.yaml": templatedManifests})
	assert.Equal(t, 2, len(files))
	assert.True(t, strings.Contains(files["cronJob.yaml"], "myapp-first"))
	assert.True(t, strings.Contains(files["deployment.yaml"], "myapp-second"))
}

func Test_SplitHelmOutput_two_files_same_name(t *testing.T) {

	templatedManifests := `
---
# Source: cron-job/templates/cronJob.yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
	name: myapp-first
	namespace: production
---
# Source: cron-job/templates/cronJob.yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
	name: myapp-second
	namespace: production
`

	files := SplitHelmOutput(map[string]string{"manifest.yaml": templatedManifests})
	assert.Equal(t, 1, len(files))
	assert.True(t, strings.Contains(files["cronJob.yaml"], "myapp-first"))
	assert.True(t, strings.Contains(files["cronJob.yaml"], "myapp-second"))
}
