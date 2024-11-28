package loki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const lokiAddress string = ""

type LokiMessage struct {
	Timestamp     time.Time
	LogLine       string
	Namespace     string
	PodName       string
	ContainerName string
}

//TODO: Have a function that reads a channel

func PublishLoki(msg LokiMessage) error {
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
		return fmt.Errorf("Unable to unmarshal data")
	}
	resp, err := http.Post(lokiAddress, "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("Unable to push to Loki endpoint")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP Status %d", resp.StatusCode)
	}
	return nil
}
