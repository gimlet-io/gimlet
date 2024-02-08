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

	"github.com/gimlet-io/gimlet-cli/pkg/agent"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func buildImage(gimletHost, agentKey, buildId string, trigger dx.ImageBuildRequest, messages chan *streaming.WSMessage, imageBuilderHost string) {
	tarFile, err := ioutil.TempFile("/tmp", "source-*.tar.gz")
	if err != nil {
		logrus.Errorf("cannot get temp file: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}
	defer tarFile.Close()

	reqUrl := fmt.Sprintf("%s/agent/imagebuild/%s", gimletHost, buildId)
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

	imageBuilder(
		tarFile.Name(),
		imageBuilderHost,
		trigger,
		messages,
		buildId,
	)
}

func imageBuilder(
	path string, url string,
	trigger dx.ImageBuildRequest,
	messages chan *streaming.WSMessage,
	buildId string,
) {
	request, err := newfileUploadRequest(url, map[string]string{
		"image": trigger.Image,
		"tag":   trigger.Tag,
		"app":   trigger.App,
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
	reader := bufio.NewReader(body)
	// first := true
	for {
		line, err := reader.ReadBytes('\n')
		sb.WriteString(string(line))
		if err != nil {
			if err == io.EOF {
				break
			}

			logrus.Errorf("cannot stream build logs: %s", err)
			streamImageBuildEvent(messages, userLogin, imageBuildId, "error", sb.String())
			return
		}

		// if first || sb.Len() > 300 {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
		sb.Reset()
		// first = false
		// }
	}

	lastLine := sb.String()
	if strings.HasSuffix(lastLine, "IMAGE BUILT") {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "success", lastLine)
		return
	} else {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "notBuilt", lastLine)
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
) {
	reqUrl := fmt.Sprintf("%s/agent/imagebuild/%s", gimletHost, buildId)
	jobName := fmt.Sprintf("kaniko-%d", rand.Uint32())
	job := generateJob(trigger, jobName, reqUrl)
	_, err := kubeEnv.Client.BatchV1().Jobs("infrastructure").Create(context.TODO(), job, meta_v1.CreateOptions{})
	if err != nil {
		logrus.Errorf("cannot apply job: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	var pods *corev1.PodList
	err = wait.PollImmediate(1*time.Second, 20*time.Second, func() (done bool, err error) {
		pods, err = kubeEnv.Client.CoreV1().Pods("infrastructure").List(context.TODO(), meta_v1.ListOptions{
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
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		logrus.Errorf("cannot get pods: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
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
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
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
	reader := bufio.NewReader(podLogs)
	for {
		line, err := reader.ReadBytes('\n')
		sb.WriteString(string(line))
		if err != nil {
			if err == io.EOF {
				break
			}

			logrus.Errorf("cannot stream build logs: %s", err)
			streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
			break
		}

		if strings.Contains(strings.ToLower(sb.String()), "error") {
			streamImageBuildEvent(messages, userLogin, imageBuildId, "notBuilt", sb.String())
			break
		}

		streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
		sb.Reset()
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
	for {
		line, err := reader.ReadBytes('\n')
		sb.WriteString(string(line))
		if err != nil {
			if err == io.EOF {
				break
			}

			logrus.Errorf("cannot stream build logs: %s", err)
			streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
			break
		}

		lastLine = string(line)
		streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
		sb.Reset()
	}

	if strings.Contains(lastLine, "Pushed") {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "success", "")
		return
	} else {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "notBuilt", "")
		return
	}
}
