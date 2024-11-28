package kubeapi

type PodLogObject struct {
	Name       string
	Containers []string
	LogChannel chan string
}
