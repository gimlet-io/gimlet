package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	buildId string,
	trigger dx.ImageBuildRequest,
	messages chan *streaming.WSMessage,
) {
	job := generateJob(trigger)
	_, err := kubeEnv.Client.BatchV1().Jobs("default").Create(context.TODO(), job, meta_v1.CreateOptions{})
	if err != nil {
		logrus.Errorf("cannot apply job: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", "")
		return
	}

	var pods *corev1.PodList
	err = wait.PollImmediate(100*time.Millisecond, 20*time.Second, func() (done bool, err error) {
		pods, err = kubeEnv.Client.CoreV1().Pods("default").List(context.TODO(), meta_v1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kaniko-%s", trigger.App),
		})
		if err != nil {
			return false, err
		}

		if len(pods.Items) == 0 {
			return false, nil
		}

		if pods.Items[0].Status.Phase == corev1.PodFailed {
			return false, fmt.Errorf("pod failed")
		}

		if pods.Items[0].Status.Phase != corev1.PodRunning {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		logrus.Errorf("poll: %s", err)
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "error", err.Error())
		return
	}

	pod := pods.Items[0]
	done := streamLogs(kubeEnv, messages, pod.Name, pod.Spec.InitContainers[0].Name, trigger.TriggeredBy, buildId)
	if !done {
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "notBuilt", "")
		return
	}

	done = streamLogs(kubeEnv, messages, pod.Name, pod.Spec.Containers[0].Name, trigger.TriggeredBy, buildId)

	if done {
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "success", "")
	} else {
		streamImageBuildEvent(messages, trigger.TriggeredBy, buildId, "notBuilt", "")
	}

	// TODO cleanup kaniko on error
	backgroundDeletion := meta_v1.DeletePropagationBackground
	kubeEnv.Client.BatchV1().Jobs("default").Delete(context.TODO(), fmt.Sprintf("kaniko-%s", trigger.App), meta_v1.DeleteOptions{
		PropagationPolicy: &backgroundDeletion,
	})
}

func generateJob(trigger dx.ImageBuildRequest) *batchv1.Job {
	var backOffLimit int32 = 0
	return &batchv1.Job{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: fmt.Sprintf("kaniko-%s", trigger.App),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "download-source",
							Image: "nixery.dev/shell/curl/unzip",
							Command: []string{
								"/bin/sh",
							},
							Args: []string{
								"-c",
								fmt.Sprintf("mkdir /source && curl -L %s -o /source/source.zip && unzip /source/source.zip -d /workspace", trigger.AppSource),
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
								fmt.Sprintf("--dockerfile=%s-%s/%s", trigger.App, trigger.Tag, trigger.Dockerfile),
								fmt.Sprintf("--context=dir:///workspace/%s-%s", trigger.App, trigger.Tag),
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

func streamLogs(kubeEnv *agent.KubeEnv,
	messages chan *streaming.WSMessage,
	pod, container string,
	userLogin, imageBuildId string,
) bool {
	count := int64(100)
	logsReq := kubeEnv.Client.CoreV1().Pods("default").GetLogs(pod, &corev1.PodLogOptions{
		Container: container,
		TailLines: &count,
		Follow:    true,
	})

	podLogs, err := logsReq.Stream(context.Background())
	if err != nil {
		logrus.Errorf("could not stream pod logs: %v", err)
		streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
		return false
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
			streamImageBuildEvent(messages, userLogin, imageBuildId, "error", sb.String())
			return false
		}

		// if first || sb.Len() > 300 {
		streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
		sb.Reset()
		// first = false
		// }

		if strings.Contains(sb.String(), "pushed") {
			streamImageBuildEvent(messages, userLogin, imageBuildId, "success", sb.String())
			break
		}
	}

	return true
}
