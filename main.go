package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	clientset   *kubernetes.Clientset
	TIME_SCRAPE int64
	// TODO: Expose to env variable
	lokiAddress string = "http://172.16.16.4:32401/loki/api/v1/push"
)

func main() {
	log.Println("[INFO] Started Kubeapi Loki Interface")
	log.Printf("[INFO] LOKI Server %s\n", lokiAddress)
	// TODO: Add Redis to store when last push was made for particular pod
	log.Printf("[INFO] REDIS Server ...\n")
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	TIME_SCRAPE = 600

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config) // TODO: Pass Kubeconfig as a file
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	for _, namespace := range namespaces.Items {
		PodsInNamespace(namespace)
	}
}

func sendToLoki(logs string, ts time.Time, namespace string, pod string) error {
	timestamp := fmt.Sprintf("%d", ts.UnixNano())
	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"job":       "kubernetes",
					"namespace": namespace,
					"pod":       pod,
				},
				"values": [][]string{
					{timestamp, logs},
				},
			},
		},
	}

	reqJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("[Err] Unable to unmarshal for (%s) %s", namespace, pod)
	}
	resp, err := http.Post(lokiAddress, "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("[Err] Unable to push loki endpoint (%s) %s", namespace, pod)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 { //204
		return fmt.Errorf("[Err] HTTP Status %d (%s) %s", resp.StatusCode, namespace, pod)
	}
	return nil
}
