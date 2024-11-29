package kubeapi

import (
	"context"
	"io"
	"time"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
				LogChannel: make(chan string, 1),
			}
			for _, container := range pod.Spec.Containers {
				podObject.Containers = append(podObject.Containers, container.Name)
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
	for _, container := range pod.Containers {
		log.Debugf("Started goroutine :: (%s) %s/%s", ns, pod.Name, container)
		podLogOpts := corev1.PodLogOptions{
			SinceSeconds: &ScrapeInterval,
			Timestamps:   true,
			Container:    container,
		}
		req := clientset.CoreV1().Pods(ns).GetLogs(pod.Name, &podLogOpts)
		// https://stackoverflow.com/questions/53852530/how-to-get-logs-from-kubernetes-using-go#53870271
		go func(*rest.Request, int64) {
			for {
				result, err := req.Stream(context.Background())
				if err != nil {
					log.Warn(err)
				}
				buf := make([]byte, 2000)
				numBytes, err := result.Read(buf)
				if err == io.EOF {
					break
				}
				time.Sleep(time.Second * time.Duration(ScrapeInterval))
				if err != nil {
					log.Warn(err)
				}
				message := string(buf[:numBytes])
				pod.LogChannel <- message
			}
		}(req, ScrapeInterval)
	}
}

func StartGoRoutinesForPodLogFetching(clientset *kubernetes.Clientset, namespaces []Namespace) {
	for _, ns := range namespaces {
		for _, pod := range ns.Pods {
			go GetPodLogs(clientset, ns.Name, pod)
		}
	}
}
