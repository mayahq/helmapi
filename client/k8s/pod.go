package k8s

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type PodSummary struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	Uid       string      `json:"uid"`
	CreatedAt metav1.Time `json:"createdAt"`
	OwnerId   string      `json:"ownerId"`
	Node      string      `json:"node"`
	Status    string      `json:"status"`
	HostIP    string      `json:"hostIP"`
	PodIP     string      `json:"podIP"`
	StartTime metav1.Time `json:"startTime"`
}

type PodListResult struct {
	Pods     []PodSummary `json:"pods"`
	Continue string       `json:"continue"`
}

type PodDetailsResult struct {
	Details v1.Pod     `json:"details"`
	Summary PodSummary `json:"summary"`
}

func convertMapToQueryString(mapToConv map[string]string) string {
	expressions := make([]string, len(mapToConv))

	i := 0
	for key, value := range mapToConv {
		expressions[i] = key + "=" + value
	}

	return strings.Join(expressions, ",")
}

func getClient(configLocation string) (typev1.CoreV1Interface, error) {
	kubeconfig := filepath.Clean(configLocation)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1(), nil
}

func getSummaryFromPod(pod *v1.Pod) PodSummary {
	meta := pod.ObjectMeta
	spec := pod.Spec
	status := pod.Status

	var startTime metav1.Time
	if len(status.ContainerStatuses) > 0 {
		startTime = status.ContainerStatuses[0].State.Running.StartedAt
	}

	return PodSummary{
		Name:      meta.Name,
		Namespace: meta.Namespace,
		Uid:       string(meta.UID),
		CreatedAt: meta.CreationTimestamp,
		OwnerId:   meta.Labels["userRuntimeOwner"],
		Node:      spec.NodeName,
		Status:    string(status.Phase),
		HostIP:    status.HostIP,
		PodIP:     status.PodIP,
		StartTime: startTime,
	}
}

func GetPodsBySelector(
	ctx context.Context,
	namespace string,
	selector string,
	limit int64,
	cont string,
) (
	PodListResult, error,
) {
	kubeconfig := os.Getenv("KUBECONFIG")
	k8sClient, err := getClient(kubeconfig)
	if err != nil {
		return PodListResult{}, err
	}

	listOptions := metav1.ListOptions{
		LabelSelector: selector,
		Limit:         limit,
		Continue:      cont,
	}

	pods, perr := k8sClient.Pods(namespace).List(ctx, listOptions)
	if perr != nil {
		return PodListResult{}, err
	}

	result := make([]PodSummary, len(pods.Items))
	for i, pod := range pods.Items {
		result[i] = getSummaryFromPod(&pod)
	}

	podRes := PodListResult{
		Pods:     result,
		Continue: pods.Continue,
	}
	// log.Println(pods.Items[0])

	return podRes, nil
}

func GetPodByName(ctx context.Context, namespace string, name string) (PodDetailsResult, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	k8sClient, err := getClient(kubeconfig)
	if err != nil {
		return PodDetailsResult{}, err
	}

	pod, perr := k8sClient.Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if perr != nil {
		return PodDetailsResult{}, perr
	}

	result := PodDetailsResult{
		Details: *pod,
		Summary: getSummaryFromPod(pod),
	}

	return result, nil
}
