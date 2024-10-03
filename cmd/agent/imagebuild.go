package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet/pkg/agent"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func buildImage(kubeEnv *agent.KubeEnv, gimletHost, agentKey, buildId string, trigger dx.ImageBuildRequest, messages chan *streaming.WSMessage, imageBuilderHost string) {
	tarFile, err := ioutil.TempFile("/tmp", "source-*.tar.gz")
	if err != nil {
		logrus.Errorf("cannot get temp file: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}
	defer tarFile.Close()

	reqUrl := fmt.Sprintf("%s/agent/imagebuild/%s?access_token=%s", gimletHost, buildId, agentKey)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		logrus.Errorf("could not create http request: %v", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}
	req.Header.Set("Authorization", "BEARER "+agentKey)
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("could not download tarfile: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.Errorf("could download tar file: %d - %v", resp.StatusCode, string(body))
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	_, err = io.Copy(tarFile, resp.Body)
	if err != nil {
		logrus.Errorf("could not download tarfile: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	configJson, err := getSecretConfigJson(kubeEnv, trigger.Registry)
	if err != nil {
		logrus.Errorf("could not get secret configjson: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	imageBuilder(
		tarFile.Name(),
		imageBuilderHost,
		configJson,
		trigger,
		messages,
		buildId,
	)
}

func getSecretConfigJson(kubeEnv *agent.KubeEnv, registry string) (string, error) {
	pushSecret, err := kubeEnv.Client.CoreV1().Secrets("infrastructure").Get(context.TODO(), strings.ToLower(registry)+"-pushsecret", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(pushSecret.Data["config.json"]), nil
}

func imageBuilder(
	path string, url string,
	configJson string,
	trigger dx.ImageBuildRequest,
	messages chan *streaming.WSMessage,
	buildId string,
) {
	request, err := newfileUploadRequest(url, map[string]string{
		"image":      trigger.Image,
		"tag":        trigger.Tag,
		"app":        trigger.App,
		"configJson": configJson,
	}, "data", path)
	if err != nil {
		logrus.Errorf("cannot upload file: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logrus.Errorf("cannot upload file: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	streamImageBuilderLogs(resp.Body, messages, trigger.TriggeredBy, buildId)
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func streamImageBuilderLogs(
	body io.ReadCloser,
	messages chan *streaming.WSMessage,
	userLogin string,
	imageBuildId string,
) {
	defer body.Close()

	var sb strings.Builder
	var lastLine string
	reader := bufio.NewReader(body)
	logCh := make(chan string)

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			lastLine = string(line)
			logCh <- lastLine
			if err != nil {
				if err == io.EOF {
					close(logCh)
					break
				}

				logrus.Errorf("cannot stream build logs: %s", err)
				streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
				return
			}
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		logEnded := false

		select {
		case logLine, ok := <-logCh:
			if !ok {
				logEnded = true
				streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
				sb.Reset()
				break
			}
			sb.WriteString(string(logLine))

			if sb.Len() > 4000 {
				streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
				sb.Reset()
			}
		case <-ticker.C:
			streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
			sb.Reset()
		}

		if logEnded {
			break
		}
	}

	if strings.HasSuffix(lastLine, "IMAGE BUILT") {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "success", "")
		return
	} else {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "notBuilt", "")
		return
	}
}

func streamImageBuildEvent(messages chan *streaming.WSMessage, userLogin string, imageBuildId string, status string, logLine string) {
	serializedPayload, err := json.Marshal(streaming.ImageBuildStatusWSMessage{
		ClientId: userLogin,
		BuildId:  imageBuildId,
		Status:   status,
		LogLine:  string(logLine),
	})
	if err != nil {
		logrus.Error("cannot serialize payload", err)
	}

	msg := &streaming.WSMessage{
		Type:    "imageBuildLogs",
		Payload: string(serializedPayload),
	}
	messages <- msg
}

func dockerfileImageBuild(
	kubeEnv *agent.KubeEnv,
	gimletHost, buildId string,
	trigger dx.ImageBuildRequest,
	messages chan *streaming.WSMessage,
	agentKey string,
) {
	reqUrl := fmt.Sprintf("%s/agent/imagebuild/%s?access_token=%s", gimletHost, buildId, agentKey)
	jobName := fmt.Sprintf("kaniko-%d", rand.Uint32())
	job := generateJob(trigger, jobName, reqUrl)
	job = mountPushSecret(job, trigger.Registry)
	job = mountRegistryCertSecret(kubeEnv.Client, job, trigger)
	_, err := kubeEnv.Client.BatchV1().Jobs("infrastructure").Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		logrus.Errorf("cannot apply job: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	var pods *corev1.PodList
	err = wait.PollImmediate(1*time.Second, 30*time.Second, func() (done bool, err error) {
		pods, err = kubeEnv.Client.CoreV1().Pods("infrastructure").List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", jobName),
		})
		if err != nil {
			return false, err
		}

		if len(pods.Items) == 0 {
			return false, nil
		}

		for _, containerStatus := range pods.Items[0].Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "running", fmt.Sprintf("%s: %s\n", pods.Items[0].Name, containerStatus.State.Waiting.Reason))
			}
		}

		if pods.Items[0].Status.Phase == corev1.PodPending {
			streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "running", fmt.Sprintf("%s: %s\n", pods.Items[0].Name, corev1.PodPending))
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		pods, err = kubeEnv.Client.CoreV1().Pods("infrastructure").List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", jobName),
		})
		if err != nil {
			logrus.Error(err)
			streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "ERROR: could not start up image build pods")
			return
		}

		if len(pods.Items) == 0 {
			logrus.Error(fmt.Errorf("found zero pods"))
			streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "ERROR: could not start up image build pods")
			return
		}

		if pods.Items[0].Status.Phase == corev1.PodPending {
			events, _ := kubeEnv.Client.CoreV1().Events("infrastructure").List(
				context.TODO(),
				metav1.ListOptions{FieldSelector: "involvedObject.name=" + pods.Items[0].ObjectMeta.Name, TypeMeta: metav1.TypeMeta{Kind: "Pod"}})
			for _, item := range events.Items {
				logrus.Error(fmt.Errorf("kaniko pod pending: %s ", item.Message))
				streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "ERROR: kaniko pod stuck in Pending state: "+item.Message+"\n")
			}
			for _, c := range pods.Items[0].Status.Conditions {
				logrus.Error(fmt.Errorf("kaniko pod pending: %s ", c.Message))
				streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "ERROR: kaniko pod stuck in Pending state: "+c.Message+"\n")
			}
		} else {
			streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "ERROR: unknown error. Could not start up image build pods, check for kaniko pods in the infrastructure namespace")
		}

		dp := metav1.DeletePropagationBackground
		err := kubeEnv.Client.BatchV1().Jobs("infrastructure").Delete(context.TODO(), jobName, metav1.DeleteOptions{
			PropagationPolicy: &dp,
		})
		if err != nil {
			logrus.Errorf("could not delete job %s: %s", jobName, err)
		}

		return
	}

	pod := pods.Items[0]
	streamInitContainerLogs(kubeEnv, messages, pod.Name, pod.Spec.InitContainers[0].Name, trigger.TriggeredBy, buildId)
	streamLogs(kubeEnv, messages, pod.Name, pod.Spec.Containers[0].Name, trigger.TriggeredBy, buildId)
}

func generateJob(trigger dx.ImageBuildRequest, name, sourceUrl string) *batchv1.Job {
	var ttlSecondsAfterFinished int32 = 60
	var backOffLimit int32 = 0
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "download-source",
							Image: "alpine",
							Command: []string{
								"/bin/sh",
							},
							Args: []string{
								"-c",
								fmt.Sprintf(`apk update; apk add curl; apk add tar && mkdir /source && curl -X GET -H "Content-Type: application/json" %s -o /source/source.tar.gz && tar xvf source/source.tar.gz -C /workspace`, sourceUrl),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/workspace",
									Name:      "workspace",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: "gcr.io/kaniko-project/executor:latest",
							Args: []string{
								fmt.Sprintf("--dockerfile=/workspace/%s", trigger.Dockerfile),
								"--context=dir:///workspace",
								fmt.Sprintf("--destination=%s:%s", trigger.Image, trigger.Tag),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/workspace",
									Name:      "workspace",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu":    resource.MustParse("4"),
									"memory": resource.MustParse("8Gi"),
								},
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse("1000m"),
									"memory": resource.MustParse("1000Mi"),
								},
							},
						},
					},
					RestartPolicy: "Never",
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									SizeLimit: &[]resource.Quantity{
										resource.MustParse("500Mi"),
									}[0],
								},
							},
						},
					},
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}
}

func mountPushSecret(job *batchv1.Job, registry string) *batchv1.Job {
	if registry == "" {
		return job
	}

	job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		MountPath: "/kaniko/.docker",
		Name:      "pushsecret",
	})
	optional := false
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "pushsecret",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: strings.ToLower(registry) + "-pushsecret",
				Optional:   &optional,
			},
		},
	})

	return job
}

func mountRegistryCertSecret(client kubernetes.Interface, job *batchv1.Job, trigger dx.ImageBuildRequest) *batchv1.Job {
	if trigger.Registry == "" {
		return job
	}

	_, err := client.CoreV1().Secrets("infrastructure").Get(context.TODO(), strings.ToLower(trigger.Registry)+"-registrycert", metav1.GetOptions{})
	if err != nil {
		return job
	}

	url := strings.Split(trigger.Image, "/")[0]

	job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		MountPath: "/kaniko/mounted-certs/",
		Name:      "registrycert",
	})
	optional := false
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "registrycert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: strings.ToLower(trigger.Registry) + "-registrycert",
				Optional:   &optional,
			},
		},
	})

	job.Spec.Template.Spec.Containers[0].Args = append(job.Spec.Template.Spec.Containers[0].Args,
		fmt.Sprintf("--registry-certificate=%s=/kaniko/mounted-certs/certificate.cert", url),
	)

	return job
}

func streamInitContainerLogs(kubeEnv *agent.KubeEnv,
	messages chan *streaming.WSMessage,
	pod, container string,
	userLogin, imageBuildId string,
) {
	count := int64(100)
	logsReq := kubeEnv.Client.CoreV1().Pods("infrastructure").GetLogs(pod, &corev1.PodLogOptions{
		Container: container,
		TailLines: &count,
		Follow:    true,
	})

	podLogs, err := logsReq.Stream(context.Background())
	if err != nil {
		logrus.Errorf("could not stream pod logs: %v", err)
		streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
		return
	}
	defer podLogs.Close()

	var sb strings.Builder
	var lastLine string
	reader := bufio.NewReader(podLogs)
	logCh := make(chan string)

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			logCh <- string(line)
			if err != nil {
				if err == io.EOF {
					close(logCh)
					break
				}

				logrus.Errorf("cannot stream build logs: %s", err)
				streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
				break
			}
			lastLine = string(line)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		logEnded := false

		select {
		case logLine, ok := <-logCh:
			if !ok {
				logEnded = true
				streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
				sb.Reset()
				break
			}
			sb.WriteString(string(logLine))

			if sb.Len() > 4000 {
				streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
				sb.Reset()
			}
		case <-ticker.C:
			streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
			sb.Reset()
		}

		if logEnded {
			break
		}
	}

	if strings.Contains(lastLine, "error") {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "notBuilt", sb.String())
		return
	}
}

func streamLogs(kubeEnv *agent.KubeEnv,
	messages chan *streaming.WSMessage,
	pod, container string,
	userLogin, imageBuildId string,
) {
	count := int64(100)
	logsReq := kubeEnv.Client.CoreV1().Pods("infrastructure").GetLogs(pod, &corev1.PodLogOptions{
		Container: container,
		TailLines: &count,
		Follow:    true,
	})

	podLogs, err := logsReq.Stream(context.Background())
	if err != nil {
		logrus.Errorf("could not stream pod logs: %v", err)
		streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
		return
	}
	defer podLogs.Close()

	var sb strings.Builder
	var lastLine string
	reader := bufio.NewReader(podLogs)
	logCh := make(chan string)

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			logCh <- string(line)
			if err != nil {
				if err == io.EOF {
					close(logCh)
					break
				}

				logrus.Errorf("cannot stream build logs: %s", err)
				streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
				break
			}
			lastLine = string(line)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		logEnded := false

		select {
		case logLine, ok := <-logCh:
			if !ok {
				logEnded = true
				streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
				sb.Reset()
				break
			}
			sb.WriteString(string(logLine))

			if sb.Len() > 4000 {
				streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
				sb.Reset()
			}
		case <-ticker.C:
			streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
			sb.Reset()
		}

		if logEnded {
			break
		}
	}

	if strings.Contains(lastLine, "Pushed") {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "success", "")
		return
	} else {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "notBuilt", "")
		return
	}
}
