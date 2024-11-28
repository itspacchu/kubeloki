package kubeapi

import (
	"bufio"
	"context"
	"time"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodLogObject struct {
	Name       string
	Containers []string
	LogChannel chan string
}

func (p PodLogObject) Print() {
	log.Debug("+--" + p.Name)
}

type Namespace struct {
	Name           string
	Pods           []PodLogObject
	ScrapeInterval time.Duration
}

func (ns Namespace) Print() {
	log.Debug(ns.Name)
	for _, item := range ns.Pods {
		item.Print()
	}
}

func PrintNSList(nsList []Namespace) {
	for _, item := range nsList {
		item.Print()
	}
}

func GetNamespaces(clientset *kubernetes.Clientset) ([]Namespace, error) {
	coreNamespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	genNamespaceList := make([]Namespace, len(coreNamespaces.Items))
	for nsIndex, namespace := range coreNamespaces.Items {
		podList, err := clientset.CoreV1().Pods(namespace.Name).List(context.Background(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		podObjectList := make([]PodLogObject, len(podList.Items))
		for podIndex, pod := range podList.Items {
			podObject := PodLogObject{
				Name:       pod.Name,
				LogChannel: make(chan string),
			}
			podObjectList[podIndex] = podObject
		}
		genNamespaceList[nsIndex].Name = namespace.Name
		genNamespaceList[nsIndex].Pods = podObjectList
	}
	return genNamespaceList, nil
}

func GetKubeDetails(clientset *kubernetes.Clientset) ([]Namespace, error) {
	if namespaces, err := GetNamespaces(clientset); err != nil {
		return nil, err
	} else {
		return namespaces, nil
	}
}

// for a goroutine
func GetPodLogs(clientset *kubernetes.Clientset, ns string, pod PodLogObject) {
	var ScrapeInterval int64 = 1
	podLogOpts := corev1.PodLogOptions{
		SinceSeconds: &ScrapeInterval,
		Timestamps:   true,
		Container:    all,
	}
	req := clientset.CoreV1().Pods(ns).GetLogs(pod.Name, &podLogOpts)
	result, err := req.Stream(context.Background())
	if err != nil {
		log.Warn(err)
	}
	// https://stackoverflow.com/questions/53852530/how-to-get-logs-from-kubernetes-using-go#53870271
	defer result.Close()
	scanner := bufio.NewScanner(result)
	for scanner.Scan() {
		pod.LogChannel <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Warn(err)
	}
}

func StartGoRoutines(clientset *kubernetes.Clientset, namespaces []Namespace) {
	for _, ns := range namespaces {
		for _, pod := range ns.Pods {
			log.Debugf("Started goroutine :: (%s) %s\n", ns.Name, pod.Name)
			go GetPodLogs(clientset, ns.Name, pod)
		}
	}
}
