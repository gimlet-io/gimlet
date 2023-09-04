package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/agent/config"
	"github.com/gimlet-io/gimlet-cli/pkg/agent"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server/streaming"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	fmt.Println(logo())

	err := godotenv.Load(".env")
	if err != nil {
		logrus.Warnf("could not load .env file, relying on env vars")
	}

	config, err := config.Environ()
	if err != nil {
		logrus.Fatalln("main: invalid configuration")
	}

	initLogger(config)
	if logrus.IsLevelEnabled(logrus.TraceLevel) {
		logrus.Traceln(config.String())
	}

	if config.Host == "" {
		panic(fmt.Errorf("please provide the HOST variable"))
	}
	if config.AgentKey == "" {
		panic(fmt.Errorf("please provide the AGENT_KEY variable"))
	}
	if config.Env == "" {
		panic(fmt.Errorf("please provide the ENV variable"))
	}

	envName, namespace, err := parseEnvString(config.Env)
	if err != nil {
		panic(fmt.Errorf("invalid ENV variable. Format is env1=ns1,env2=ns2"))
	}

	if namespace != "" {
		logrus.Infof("Initializing %s kubeEnv in %s namespace scope", envName, namespace)
	} else {
		logrus.Infof("Initializing %s kubeEnv in cluster scope", envName)
	}

	k8sConfig, err := k8sConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		panic(err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		panic(err.Error())
	}

	kubeEnv := &agent.KubeEnv{
		Name:          envName,
		Namespace:     namespace,
		Client:        clientset,
		DynamicClient: dynamicClient,
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	podController := agent.PodController(kubeEnv, config.Host, config.AgentKey)
	deploymentController := agent.DeploymentController(kubeEnv, config.Host, config.AgentKey)
	ingressController := agent.IngressController(kubeEnv, config.Host, config.AgentKey)
	eventController := agent.EventController(kubeEnv, config.Host, config.AgentKey)
	gitRepositoryController := agent.GitRepositoryController(kubeEnv, config.Host, config.AgentKey)
	kustomizationController := agent.KustomizationController(kubeEnv, config.Host, config.AgentKey)
	helmReleaseController := agent.HelmReleaseController(kubeEnv, config.Host, config.AgentKey)
	go podController.Run(1, stopCh)
	go deploymentController.Run(1, stopCh)
	go ingressController.Run(1, stopCh)
	go eventController.Run(1, stopCh)
	go gitRepositoryController.Run(1, stopCh)
	go kustomizationController.Run(1, stopCh)
	go helmReleaseController.Run(1, stopCh)

	messages := make(chan *streaming.WSMessage)

	go serverCommunication(kubeEnv, config, messages, config.Host, config.AgentKey)
	go serverWSCommunication(config, messages)

	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// This goroutine executes a blocking receive for signals.
	// When it gets one it’ll print it out and then notify the program that it can finish.
	go func() {
		sig := <-signals
		logrus.Info(sig)
		done <- true
	}()

	logrus.Info("Initialized")
	<-done
	logrus.Info("Exiting")
}

func k8sConfig(config config.Config) (*rest.Config, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logrus.Infof("In-cluster-config didn't work (%s), loading from path in KUBECONFIG if set", err.Error())
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", config.KubeConfig)
		if err != nil {
			panic(err.Error())
		}
	}
	return k8sConfig, err
}

// helper function configures the logging.
func initLogger(c config.Config) {
	logrus.SetReportCaller(true)

	customFormatter := &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
		},
	}
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)

	if c.Logging.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if c.Logging.Trace {
		logrus.SetLevel(logrus.TraceLevel)
	}
}

func parseEnvString(envString string) (string, string, error) {
	if strings.Contains(envString, "=") {
		parts := strings.Split(envString, "=")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("")
		}
		return parts[0], parts[1], nil
	} else {
		return envString, "", nil
	}
}

func serverCommunication(
	kubeEnv *agent.KubeEnv,
	config config.Config,
	messages chan *streaming.WSMessage,
	gimletHost string,
	agentKey string,
) {
	for {
		done := make(chan bool)

		events, err := register(config.Host, kubeEnv.Name, kubeEnv.Namespace, config.AgentKey)
		if err != nil {
			logrus.Errorf("could not connect to Gimlet: %s", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}

		logrus.Info("Connected to Gimlet")
		go sendState(kubeEnv, config.Host, config.AgentKey)
		go sendEvents(kubeEnv, config.Host, config.AgentKey)
		go sendFluxState(kubeEnv, config.Host, config.AgentKey)

		runningLogStreams := NewRunningLogStreams()

		go func(events chan map[string]interface{}) {
			for {
				e, more := <-events
				if more {
					logrus.Debugf("event received: %v", e)
					switch e["action"] {
					case "refetch":
						go sendState(kubeEnv, config.Host, config.AgentKey)
						go sendEvents(kubeEnv, config.Host, config.AgentKey)
					case "podLogs":
						go podLogs(
							kubeEnv,
							e["namespace"].(string),
							e["serviceName"].(string),
							messages,
							runningLogStreams,
						)
					case "stopPodLogs":
						namespace := e["namespace"].(string)
						svc := e["serviceName"].(string)
						go runningLogStreams.Stop(namespace, svc)
					case "imageBuildTrigger":
						eString, _ := json.Marshal(e)
						var trigger streaming.ImageBuildTrigger
						_ = json.Unmarshal(eString, &trigger)

						go buildImage(gimletHost, agentKey, trigger, messages, config.ImageBuilderHost)
					}
				} else {
					logrus.Info("event stream closed")
					go runningLogStreams.StopAll()
					done <- true
					return
				}
			}
		}(events)

		<-done
		time.Sleep(time.Second * 3)
		logrus.Info("Disconnected from Gimlet")
	}
}

func sendState(kubeEnv *agent.KubeEnv, gimletHost string, agentKey string) {
	stacks, err := kubeEnv.Services("")
	if err != nil {
		logrus.Errorf("could not get state from k8s apiServer: %v", err)
		return
	}

	stacksString, err := json.Marshal(stacks)
	if err != nil {
		logrus.Errorf("could not serialize k8s state: %v", err)
		return
	}

	params := url.Values{}
	params.Add("name", kubeEnv.Name)
	reqUrl := fmt.Sprintf("%s/agent/state?%s", gimletHost, params.Encode())
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(stacksString))
	if err != nil {
		logrus.Errorf("could not create http request: %v", err)
		return
	}
	req.Header.Set("Authorization", "BEARER "+agentKey)
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.Errorf("could not send k8s state: %d - %v", resp.StatusCode, string(body))
		return
	}

	logrus.Info("init state sent")
}

func sendFluxState(kubeEnv *agent.KubeEnv, gimletHost string, agentKey string) {
	agent.SendFluxState(kubeEnv, gimletHost, agentKey)
	logrus.Info("init flux states sent")
}

func sendEvents(kubeEnv *agent.KubeEnv, gimletHost string, agentKey string) {
	events, err := kubeEnv.WarningEvents("")
	if err != nil {
		logrus.Errorf("could not get events from k8s apiServer: %v", err)
		return
	}

	eventsString, err := json.Marshal(events)
	if err != nil {
		logrus.Errorf("could not serialize k8s events: %v", err)
		return
	}

	reqUrl := fmt.Sprintf("%s/agent/events", gimletHost)
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(eventsString))
	if err != nil {
		logrus.Errorf("could not create http request: %v", err)
		return
	}
	req.Header.Set("Authorization", "BEARER "+agentKey)
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logrus.Errorf("could not send k8s events: %d - %v", resp.StatusCode, string(body))
		return
	}

	logrus.Info("init events sent")
}

func podLogs(
	kubeEnv *agent.KubeEnv,
	namespace string,
	serviceName string,
	messages chan *streaming.WSMessage,
	runningLogStreams *runningLogStreams,
) {

	svc, err := kubeEnv.Client.CoreV1().Services(namespace).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		logrus.Errorf("could not get services: %v", err)
		return
	}

	var integratedServices []v1.Service
	for _, s := range svc.Items {
		if _, ok := s.ObjectMeta.GetAnnotations()[agent.AnnotationGitRepository]; ok {
			integratedServices = append(integratedServices, s)
		}
	}

	allDeployments, err := kubeEnv.Client.AppsV1().Deployments(namespace).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		logrus.Errorf("could not get deployments: %v", err)
		return
	}

	allPods, err := kubeEnv.Client.CoreV1().Pods(namespace).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		logrus.Errorf("could not get pods: %v", err)
		return
	}

	for _, svc := range integratedServices {
		for _, deployment := range allDeployments.Items {
			if deployment.Name == serviceName {
				if agent.SelectorsMatch(deployment.Spec.Selector.MatchLabels, svc.Spec.Selector) {
					for _, pod := range allPods.Items {
						if agent.HasLabels(deployment.Spec.Selector.MatchLabels, pod.GetObjectMeta().GetLabels()) &&
							pod.Namespace == deployment.Namespace {
							containers := podContainers(pod.Spec)
							for _, container := range containers {
								go streamPodLogs(kubeEnv, namespace, pod.Name, container.Name, serviceName, messages, runningLogStreams)
							}
							return
						}
					}
				}
			}
		}
	}

	logrus.Debug("pod logs sent")
}

func httpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   20 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		},
	}
}

func streamPodLogs(
	kubeEnv *agent.KubeEnv,
	namespace string,
	pod string,
	containerName string,
	serviceName string,
	messages chan *streaming.WSMessage,
	runningLogStreams *runningLogStreams,
) {
	count := int64(100)
	podLogOpts := v1.PodLogOptions{
		Container:  containerName,
		TailLines:  &count,
		Follow:     true,
		Timestamps: true,
	}
	logsReq := kubeEnv.Client.CoreV1().Pods(namespace).GetLogs(pod, &podLogOpts)

	podLogs, err := logsReq.Stream(context.Background())
	if err != nil {
		logrus.Errorf("could not stream pod logs: %v", err)
		return
	}
	defer podLogs.Close()

	stopCh := make(chan int)
	runningLogStreams.Regsiter(stopCh, namespace, serviceName)

	go func() {
		<-stopCh
		podLogs.Close()
	}()

	sc := bufio.NewScanner(podLogs)
	for sc.Scan() {
		text := sc.Text()
		logrus.Infof(text)
		chunks := chunks(text, 1000)
		for _, chunk := range chunks {
			timestamp, message := parseMessage(chunk)
			serializedPayload, err := json.Marshal(streaming.PodLogWSMessage{
				Timestamp: timestamp,
				Container: containerName,
				Pod:       namespace + "/" + serviceName,
				Message:   message,
			})
			if err != nil {
				logrus.Error("cannot serialize payload", err)
			}

			msg := &streaming.WSMessage{
				Type:    "log",
				Payload: string(serializedPayload),
			}
			messages <- msg
		}
	}
}

func serverWSCommunication(config config.Config, messages chan *streaming.WSMessage) {
	for {
		u := webSocketURL(config.Host)

		bearerToken := "BEARER " + config.AgentKey
		c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
			"Authorization": []string{bearerToken},
		})
		if err != nil {
			logrus.Errorf("dial:%s", err.Error())
			time.Sleep(3 * time.Second)
			continue
		}

		logrus.Info("Connected ws")
		defer c.Close()

		done := make(chan struct{})

		go func() {
			defer close(done)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					logrus.Println("read:", err)
					return
				}
				logrus.Printf("recv: %s", message)
			}
		}()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			wsDisconnected := false

			select {
			case <-done:
				wsDisconnected = true
			case <-ticker.C:
				tick := &streaming.WSMessage{
					Type: "tick",
				}

				serializedMessage, err := json.Marshal(tick)
				if err != nil {
					logrus.Error("dial:", err)
				}

				err = c.WriteMessage(websocket.TextMessage, serializedMessage)
				if err != nil {
					logrus.Println("write:", err)
					return
				}
			case message := <-messages:
				serializedMessage, err := json.Marshal(message)
				if err != nil {
					logrus.Error("dial:", err)
				}

				err = c.WriteMessage(websocket.TextMessage, serializedMessage)
				if err != nil {
					logrus.Println("write:", err)
					return
				}
			}

			if wsDisconnected {
				logrus.Info("Disonnected ws")
				break
			}
		}
	}
}

func webSocketURL(host string) url.URL {
	urlSlice := strings.Split(host, "//")
	hostWithoutScheme := urlSlice[1]

	if strings.Contains(host, "https") {
		return url.URL{Scheme: "wss", Host: hostWithoutScheme, Path: "/agent/ws/"}
	}
	return url.URL{Scheme: "ws", Host: hostWithoutScheme, Path: "/agent/ws/"}
}

func chunks(str string, size int) []string {
	if len(str) <= size {
		return []string{str}
	}
	return append([]string{string(str[0:size])}, chunks(str[size:], size)...)
}

func podContainers(podSpec v1.PodSpec) (containers []v1.Container) {
	containers = append(containers, podSpec.InitContainers...)
	containers = append(containers, podSpec.Containers...)

	return containers
}

func parseMessage(chunk string) (string, string) {
	parts := strings.SplitN(chunk, " ", 2)

	return parts[0], parts[1]
}

func logo() string {
	return `
   _____ _____ __  __ _      ______ _______            _____ ______ _   _ _______
  / ____|_   _|  \/  | |    |  ____|__   __|     /\   / ____|  ____| \ | |__   __|
 | |  __  | | | \  / | |    | |__     | |       /  \ | |  __| |__  |  \| |  | |
 | | |_ | | | | |\/| | |    |  __|    | |      / /\ \| | |_ |  __| | .   |  | |
 | |__| |_| |_| |  | | |____| |____   | |     / ____ \ |__| | |____| |\  |  | |
  \_____|_____|_|  |_|______|______|  |_|    /_/    \_\_____|______|_| \_|  |_|

`
}
