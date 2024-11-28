package cmd

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/itspacchu/kubeloki/cmd/kubeapi"
	"github.com/itspacchu/kubeloki/cmd/loki"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	PodLogChannels = []kubeapi.PodLogObject{}
)

func GetKubeDetails(log *log.Logger) error {
	log.Info("Fetching Kube details")
	kubeconfig := os.Getenv("HOME") + "/.kube/config" //default kubeconfig path
	_, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Error("Kubeconfig file not found!")
		return err
	}
	log.Info(kubeconfig)
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
