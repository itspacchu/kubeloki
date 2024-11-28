package cmd

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/itspacchu/kubeloki/cmd/kubeapi"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Run() error {
	log.Info("Running Kubeloki")
	log.SetLevel(log.DebugLevel)
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
	namespaces, err := kubeapi.GetKubeDetails(clientset)
	if err != nil {
		return err
	}
	kubeapi.StartGoRoutines(clientset, namespaces)
	exit := make(chan os.Signal)
	// Launch GoRoutines for each channel
	kubeapi.PrintNSList(namespaces)
	<-exit // wait on Ctrl+C
	return nil
}
