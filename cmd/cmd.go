package cmd

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/itspacchu/kubeloki/cmd/kubeapi"
	"github.com/itspacchu/kubeloki/cmd/loki"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespaces = []kubeapi.Namespace{}
)

func GetKubeDetails(log *log.Logger) error {
	log.Info("Fetching Kube details")
	kubeconfig := os.Getenv("HOME") + "/.kube/config" //default kubeconfig path
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Error("Kubeconfig file not found!")
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("Unable to create clientset")
		return err
	}
	namespaces = kubeapi.GetNamespaces(clientset)

	if err := loki.PublishLoki(loki.LokiMessage{
		LogLine:       "Something",
		Namespace:     "has",
		PodName:       "Changed",
		ContainerName: "Within me",
		Timestamp:     time.Now(),
	}); err != nil {
		log.Warn("Unable to send to loki endpoint")
	}
	return nil
}
