package kubeapi

import (
	"time"

	"k8s.io/client-go/kubernetes"
)

type PodLogObject struct {
	Name       string
	Containers []string
	LogChannel chan string
}

type Namespace struct {
	Name           string
	Pods           []PodLogObject
	ScrapeInterval time.Duration
}

func GetNamespaces(clientset *kubernetes.Clientset) []Namespace {
	
	return []Namespace{}
}
