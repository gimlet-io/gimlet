package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/sirupsen/logrus"
)

func buildImage(
	gimletHost, agentKey, imageBuildId, image, tag, app, userLogin string,
	messages chan *streaming.WSMessage,
) {
	tarFile, err := ioutil.TempFile("/tmp", "source-*.tar.gz")
	if err != nil {
		logrus.Errorf("cannot get temp file: %s", err)
		return
	}
	defer tarFile.Close()

	reqUrl := fmt.Sprintf("%s/agent/imagebuild/%s", gimletHost, imageBuildId)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		logrus.Errorf("could not create http request: %v", err)
		return
	}
	req.Header.Set("Authorization", "BEARER "+agentKey)
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("could not download tarfile: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.Errorf("could download tar file: %d - %v", resp.StatusCode, string(body))
		return
	}

	_, err = io.Copy(tarFile, resp.Body)
	if err != nil {
		logrus.Errorf("could not download tarfile: %s", err)
		return
	}

	imageBuilder(
		tarFile.Name(),
		// "http://image-builder.infrastructure.svc.cluster.local:9000/build-image",
		"http://127.0.0.1:8001/build-image",
		image, tag, app,
		userLogin,
		imageBuildId,
		messages,
	)
}

func imageBuilder(
	path string, url string,
	image string, tag string, app string,
	userLogin string,
	imageBuildId string,
	messages chan *streaming.WSMessage,
) {
	request, err := newfileUploadRequest(url, map[string]string{
		"image": image,
		"tag":   tag,
		"app":   app,
	}, "data", path)
	if err != nil {
		logrus.Errorf("cannot upload file: %s", err)
		streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
		return
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logrus.Errorf("cannot upload file: %s", err)
		streamImageBuildEvent(messages, userLogin, imageBuildId, "error", "")
		return
	}

	streamImageBuilderLogs(resp.Body, messages, userLogin, imageBuildId)
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
	first := true
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

		if first || sb.Len() > 1000 {
			streamImageBuildEvent(messages, userLogin, imageBuildId, "running", sb.String())
			sb.Reset()
			first = false
		}
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
