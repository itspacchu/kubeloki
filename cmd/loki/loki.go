package loki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/itspacchu/kubeloki/cmd/kubeapi"
)

const lokiAddress string = ""

type LokiMessage struct {
	Timestamp     time.Time
	LogLine       string
	Namespace     string
	PodName       string
	ContainerName string
}

// TODO: Have a function that reads a channel
func PublishLoki(msg LokiMessage) {
	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"job":       "kubernetes",
					"namespace": msg.Namespace,
					"pod":       msg.PodName,
					"container": msg.ContainerName,
				},
				"values": [][]string{
					{fmt.Sprintf("%d", msg.Timestamp.UnixNano()), msg.LogLine},
				},
			},
		},
	}

	reqJSON, err := json.Marshal(payload)
	if err != nil {
		log.Warn("Unable to unmarshal data")
		return
	}
	resp, err := http.Post(lokiAddress, "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		log.Warnf("Log Failed to push :: (%s) %s -- %s", msg.Namespace, msg.PodName, msg)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Warnf("HTTP Status %d", resp.StatusCode)
		return
	}
}

func StreamToLoki(strChannel chan string, podName string, nameSpace string, containerName string) {
	for msg := range strChannel {
		lm := LokiMessage{
			Timestamp:     time.Now(),
			LogLine:       msg,
			Namespace:     nameSpace,
			PodName:       podName,
			ContainerName: containerName,
		}
		go PublishLoki(lm)
	}
}

func StartGoRoutinesForLokiSending(namespaces []kubeapi.Namespace) {
	for _, ns := range namespaces {
		for _, pod := range ns.Pods {
			go StreamToLoki(pod.LogChannel, pod.Name, ns.Name, pod.Containers[0]) //TODO: Send container info seperation as well
		}
	}
}
