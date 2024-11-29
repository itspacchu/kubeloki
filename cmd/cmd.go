package cmd

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/itspacchu/kubeloki/cmd/kubeapi"
	"github.com/itspacchu/kubeloki/cmd/loki"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var Reload bool = false

func Run() error {
	log.Info("Running Kubeloki")
	log.SetLevel(log.DebugLevel)
	log.Info("Fetching Kube details")
	kubeconfig := os.Getenv("HOME") + "/.kube/config"
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
	exit := make(chan os.Signal)
	err = StartSendingLogs(clientset)
	<-exit // wait on Ctrl+C
	return err
}

func StartSendingLogs(clientset *kubernetes.Clientset) error {
	namespaces, err := kubeapi.GetKubeDetails(clientset)
	if err != nil {
		return err
	}
	kubeapi.StartGoRoutinesForPodLogFetching(clientset, namespaces)
	loki.StartGoRoutinesForLokiSending(namespaces)
	go CheckIfThereAreNewNamespaces(clientset, &namespaces)
	return nil
}

func CheckIfThereAreNewNamespaces(clientset *kubernetes.Clientset, namespaces *[]kubeapi.Namespace) {
	for {
		currentLen := len(*namespaces)
		fetchNsLen, _ := clientset.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
		if currentLen == len(fetchNsLen.Items) {
			Reload = false
		} else {
			log.Infof("There might be new namespaces (old:%d) (new:%d) Reloading!", currentLen, len(fetchNsLen.Items))
			Reload = true
			*namespaces, _ = kubeapi.GetKubeDetails(clientset)
			kubeapi.PrintNSList(*namespaces)
		}
		time.Sleep(5 * time.Second)
	}
}
