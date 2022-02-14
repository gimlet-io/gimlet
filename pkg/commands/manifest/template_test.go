package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/gimlet-io/gimlet-cli/pkg/commands"
	"github.com/gimlet-io/gimletd/dx"
	"github.com/stretchr/testify/assert"
)

const manifestWithRemoteHelmChart = `
app: myapp
env: staging
namespace: my-team
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.21.0
values:
  replicas: 1
  image:
    repository: myapp
    tag: 1.1.0
  ingress:
    host: myapp.staging.mycompany.com
    tlsEnabled: true
  volumes:
  - name: uploads
    path: /files
    size: 12Gi
    storageClass: efs-ftp-uploads
  - name: errors
    path: /tmp/err
    size: 12Gi
    storageClass: efs-ftp-errors
`

const manifestWithLocalChart = `
app: myapp
env: staging
namespace: my-team
chart:
  name: ../../../fixtures/localChart/hello-server
values:
  replicaCount: 2
`

const manifestWithPrivateGitRepoHTTPS = `
app: myapp
env: staging
namespace: my-team
chart:
  name: https://github.com/gimlet-io/onechart.git?sha=8e52597ae4fb4ed7888c819b3c77331622136aba&path=/charts/onechart/
values:
  replicas: 10
`

const manifestWithKustomizePatch = `
app: myapp
env: staging
namespace: my-team
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.10.0
values:
  replicas: 10  
strategicMergePatches: |
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: myapp
    namespace: my-team
  spec:
    template:
      spec:
        containers:
        - name: myapp
          volumeMounts:
            - name: azure-file
              mountPath: /azure-bucket
      volumes:
        - name: azure-file
          azureFile:
            secretName: my-azure-secret
            shareName: my-azure-share
            readOnly: false
  ---
`

const manifestWithKustomizeJsonPatch = `
app: myapp
env: staging
namespace: my-team
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.32.0
values:
  replicas: 10
  ingress:
    host: myapp.staging.mycompany.com
    tlsEnabled: true
json6902Patches:
- target:
    group: "networking.k8s.io"
    version: "v1"
    kind: "Ingress"
    name: "myapp"
  patch: |
    ---
    - op: replace
      path: /spec/rules/0/host
      value: myapp.com
    - op: replace
      path: /spec/tls/0/hosts/0
      value: myapp.com
`

const manifestWithChartAndRawYaml = `
app: myapp
env: staging
namespace: my-team
chart:
  repository: https://chart.onechart.dev
  name: onechart
  version: 0.10.0
values:
  replicas: 10
  image:
    repository: myapp
    tag: 1.1.0
  ingress:
    host: myapp.staging.mycompany.com
    tlsEnabled: true
manifests: |
  ---
  apiVersion: v1
  kind: Service
  metadata:
    name: myapp-svc-02
    namespace: my-team
  spec:
    ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: http
    selector:
      app.kubernetes.io/instance: myapp
      app.kubernetes.io/name: onechart
    type: LoadBalancer  
`

const manifestWithRawYamlandPatch = `
app: myapp
env: staging
namespace: my-team
manifests: |
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app.kubernetes.io/instance: myapp
      app.kubernetes.io/managed-by: Helm
      app.kubernetes.io/name: onechart
      helm.sh/chart: onechart-0.10.0
    name: myapp
    namespace: my-team
  spec:
    replicas: 10
    selector:
      matchLabels:
        app.kubernetes.io/instance: myapp
        app.kubernetes.io/name: onechart
    template:
      metadata:
        annotations:
          checksum/config: 01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b
        labels:
          app.kubernetes.io/instance: myapp
          app.kubernetes.io/name: onechart
      spec:
        containers:
        - image: myapp:1.1.0
          name: myapp
          ports:
          - containerPort: 80
            name: http
            protocol: TCP
          resources:
            limits:
              cpu: 200m
              memory: 200Mi
            requests:
              cpu: 200m
              memory: 200Mi
          securityContext: {}
          volumeMounts:
          - mountPath: /azure-bucket
            name: azure-file
        securityContext:
          fsGroup: 999
      volumes:
      - azureFile:
          readOnly: false
          secretName: my-azure-secret
          shareName: my-azure-share
        name: azure-file
strategicMergePatches: |
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: myapp
    namespace: my-team
  spec:
    template:
      spec:
        containers:
        - name: myapp
          volumeMounts:
            - name: azure-file
              mountPath: /azure-bucket
      volumes:
        - name: azure-file
          azureFile:
            secretName: my-azure-secret
            shareName: my-azure-share
            readOnly: false
  ---`

const manifestwithRaWYaml = `
app: myapp
env: staging
namespace: my-team
manifests: |
  ---
  apiVersion: v1
  kind: Service
  metadata:
    labels:
      app.kubernetes.io/instance: myapp
      app.kubernetes.io/managed-by: Helm
      app.kubernetes.io/name: onechart
      helm.sh/chart: onechart-0.10.0
    name: myapp
    namespace: my-team
  spec:
    ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: http
    selector:
      app.kubernetes.io/instance: myapp
      app.kubernetes.io/name: onechart
    type: ClusterIP
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app.kubernetes.io/instance: myapp
      app.kubernetes.io/managed-by: Helm
      app.kubernetes.io/name: onechart
      helm.sh/chart: onechart-0.10.0
    name: myapp
    namespace: my-team
  spec:
    replicas: 10
    selector:
      matchLabels:
        app.kubernetes.io/instance: myapp
        app.kubernetes.io/name: onechart
    template:
      metadata:
        annotations:
          checksum/config: 01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b
        labels:
          app.kubernetes.io/instance: myapp
          app.kubernetes.io/name: onechart
      spec:
        containers:
        - image: myapp:1.1.0
          name: myapp
          ports:
          - containerPort: 80
            name: http
            protocol: TCP
          resources:
            limits:
              cpu: 200m
              memory: 200Mi
            requests:
              cpu: 200m
              memory: 200Mi
          securityContext: {}
          volumeMounts:
          - mountPath: /azure-bucket
            name: azure-file
        securityContext:
          fsGroup: 999
      volumes:
      - azureFile:
          readOnly: false
          secretName: my-azure-secret
          shareName: my-azure-share
        name: azure-file
  ---
  apiVersion: networking.k8s.io/v1beta1
  kind: Ingress
  metadata:
    labels:
      app.kubernetes.io/instance: myapp
      app.kubernetes.io/managed-by: Helm
      app.kubernetes.io/name: onechart
      helm.sh/chart: onechart-0.10.0
    name: myapp
    namespace: my-team
  spec:
    rules:
    - host: myapp.staging.mycompany.com
      http:
        paths:
        - backend:
            serviceName: myapp
            servicePort: 80
    tls:
    - hosts:
      - myapp.staging.mycompany.com
      secretName: tls-myapp  
`

const cueTemplate = `
import "text/template"

_instances: [
  "first",
  "second",
]

configs: [ for instance in _instances {
  app:       template.Execute("myapp-{{ . }}", instance)
  env:       "production"
  namespace: "production"
  chart: {
    repository: "https://chart.onechart.dev"
    name:       "cron-job"
    version:    0.32
  }
  values: {
    image: {
      repository: "<account>.dkr.ecr.eu-west-1.amazonaws.com/myapp"
      tag:        "1.1.1"
    }
  }
}]
`

func Test_template(t *testing.T) {
	g := goblin.Goblin(t)

	args := strings.Split("gimlet manifest template", " ")

	g.Describe("gimlet manifest template", func() {
		g.It("Should template a manifest file with remote chart", func() {
			g.Timeout(60 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithRemoteHelmChart), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "myapp:1.1.0")).IsTrue("Templated manifest should have the image reference")
			//fmt.Println(string(templated))
		})

		g.It("Should template a manifest file with local chart", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithLocalChart), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "hello-server:v0.1.0")).IsTrue("Templated manifest should have the image reference")
			//fmt.Println(string(templated))
		})

		g.It("Should template a manifest file with a private git hosted chart", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithPrivateGitRepoHTTPS), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "replicas: 10")).IsTrue("should set replicas")
			//fmt.Println(string(templated))
		})
		g.It("Should template a manifest file with a kustomize patch", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithKustomizePatch), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "mountPath: /azure-bucket")).IsTrue("the spec should contain volumeMounts")
			// fmt.Println(string(templated))
		})
		g.It("Should template a manifest file with a kustomize json patch", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithKustomizeJsonPatch), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			// fmt.Println(string(templated))
			g.Assert(strings.Contains(string(templated), "host: myapp.com")).
				IsTrue("ingress url should have been replaced by kustomize json patch")
		})
		g.It("Should template a manifest file with Chart and raw yaml", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithChartAndRawYaml), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "type: LoadBalancer")).IsTrue("the service spec should contain type: LoadBalancer ")
			g.Assert(strings.Contains(string(templated), "app.kubernetes.io/managed-by: Helm")).IsTrue("the resources should contain app.kubernetes.io/managed-by: Helm label")
			// fmt.Println(string(templated))
		})
		g.It("Should template a manifest file with raw yaml and patch", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestWithRawYamlandPatch), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "mountPath: /azure-bucket")).IsTrue("the deployment spec should contain mountPath: /azure-bucket")
			// fmt.Println(string(templated))
		})
		g.It("Should template a manifest file with raw yaml", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(manifestwithRaWYaml), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "secretName: tls-myapp")).IsTrue("the ingress spec should contain secretName: tls-myapp")
			// fmt.Println(string(templated))
		})
		g.It("Should template a cue file", func() {
			g.Timeout(100 * time.Second)
			manifestFile, err := ioutil.TempFile("", "gimlet-cli-test-*.cue")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(manifestFile.Name())
			templatedFile, err := ioutil.TempFile("", "gimlet-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(templatedFile.Name())

			ioutil.WriteFile(manifestFile.Name(), []byte(cueTemplate), commands.File_RW_RW_R)
			args = append(args, "-f", manifestFile.Name())
			args = append(args, "-o", templatedFile.Name())

			err = commands.Run(&Command, args)
			g.Assert(err == nil).IsTrue(err)

			templated, err := ioutil.ReadFile(templatedFile.Name())
			g.Assert(err == nil).IsTrue(err)
			if err != nil {
				t.Fatal(err)
			}
			g.Assert(strings.Contains(string(templated), "myapp-first")).IsTrue("should render two manifests")
			g.Assert(strings.Contains(string(templated), "myapp-second")).IsTrue("should render two manifests")
		})
	})
}

func Test_ProcessCue(t *testing.T) {
	manifests, err := dx.RenderCueToManifests(cueTemplate)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(manifests))

	_, err = dx.RenderCueToManifests(`
a: "hello"
`)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "should have a `configs` field"))
}
