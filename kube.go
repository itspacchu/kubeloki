package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func LogsOfPod(pod corev1.Pod) error {
	fmt.Printf("[%d] %s\n", os.Getpid(), pod.Name)
	podLogOpts := corev1.PodLogOptions{
		SinceSeconds: &TIME_SCRAPE,

		Timestamps: true,
	}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs := req.Do(context.Background())
	rawPodLogs, err := podLogs.Raw() // TODO: storing logs in memory is not efficient!
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("[Err] Something went wrong in fetching logs for (%s) %s", pod.Namespace, pod.Name)
	}
	fmt.Printf("got %d lines .. parsing and sending ...", len(rawPodLogs))
	for _, logLine := range strings.Split(string(rawPodLogs), "\n") {
		logInd := strings.Split(logLine, " ")
		log := logInd[1:]
		layout := time.RFC3339Nano
		timestamp, _ := time.Parse(layout, logInd[0])
		go sendToLoki(fmt.Sprintf("%s", log), timestamp, pod.Namespace, pod.Name)
	}
	fmt.Println("Waiting on Goroutine completions! Goroutine count: ", runtime.NumGoroutine())
	fmt.Printf("[DONE:%d] %s\n", os.Getpid(), pod.Name)
	return nil
}

func PodsInNamespace(namespace corev1.Namespace) {
	pods, err := clientset.CoreV1().Pods(namespace.Name).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Creating Goroutines for %s\n", namespace.Name)
	for _, pod := range pods.Items {
		go LogsOfPod(pod)
	}
	for {
		if runtime.NumGoroutine() < 2 {
			break
		} else {
			fmt.Println("Current GOROUTINE count: ", runtime.NumGoroutine())
			time.Sleep(time.Duration(30) * time.Second)
		}
	}
}
