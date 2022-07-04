package process

import (
	"context"
	"fmt"
	k8client "k8res/internal/k8s/client"
	"k8res/pkg/config"
	"k8res/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes/typed/core/v1"
)

//AllPodResStore ex: [ns][podName]
type AllPodResStore map[string]map[string]PodResStore

//PodResStore ex: [request/limit/usage][cpu/mem/disk][normal/min/mas] = int64
type PodResStore map[string]map[string]map[string]int64

func GetPodRes(k8 *k8client.K8s, store AllPodResStore) error {
	var podClient client.PodInterface
	var pvcClient client.PersistentVolumeClaimInterface
	var podStore PodResStore
	var err error

	ctx := context.TODO()
	nsClient := k8.ClientSet.CoreV1().Namespaces()
	allNamespaces, err := nsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	usedNamespaces := getNamespaces(allNamespaces.Items)

	for _, ns := range usedNamespaces {
		podClient = k8.ClientSet.CoreV1().Pods(ns)
		pvcClient = k8.ClientSet.CoreV1().PersistentVolumeClaims(ns)
		pods, err := podClient.List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != "Running" {
				logger.Debugf("pod %s is not running", pod.Name)
				continue
			}
			if _, ok := store[pod.Namespace]; !ok {
				store[pod.Namespace] = make(map[string]PodResStore)
			}
			if _, ok := store[pod.Namespace][pod.Name]; !ok {
				store[pod.Namespace][pod.Name] = make(PodResStore)
			}
			podStore = store[pod.Namespace][pod.Name]

			podStoreInit(podStore)
			resetNormalCount(podStore)

			// requests & limits
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests.Cpu() != nil {
					podStore["request"]["cpu"]["normal"] += container.Resources.Requests.Cpu().MilliValue()
				}
				if container.Resources.Requests.Memory() != nil {
					podStore["request"]["mem"]["normal"] += container.Resources.Requests.Memory().Value()
				}
				if container.Resources.Limits.Cpu() != nil {
					podStore["limit"]["cpu"]["normal"] += container.Resources.Limits.Cpu().MilliValue()
				}
				if container.Resources.Limits.Memory() != nil {
					podStore["limit"]["mem"]["normal"] += container.Resources.Limits.Memory().Value()
				}
			}

			// disk
			for _, volume := range pod.Spec.Volumes {
				if volume.VolumeSource.PersistentVolumeClaim != nil {
					pvc, err := pvcClient.Get(ctx, volume.VolumeSource.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
					if err != nil {
						return err
					}
					if pvc.Spec.Resources.Requests.Storage() != nil {
						podStore["request"]["disk"]["normal"] += pvc.Spec.Resources.Requests.Storage().Value()
					}
					if pvc.Spec.Resources.Limits.Storage() != nil {
						podStore["limit"]["disk"]["normal"] += pvc.Spec.Resources.Limits.Storage().Value()
					}
				}
			}

			// usage
			mc := k8.MetricsClient.MetricsV1beta1().PodMetricses(pod.Namespace)
			podMetrics, err := mc.Get(ctx, pod.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			for _, container := range podMetrics.Containers {
				if container.Usage.Cpu() != nil {
					podStore["usage"]["cpu"]["normal"] += container.Usage.Cpu().MilliValue()
				}
				if container.Usage.Memory() != nil {
					podStore["usage"]["mem"]["normal"] += container.Usage.Memory().Value()
				}
				if container.Usage.Storage() != nil {
					podStore["usage"]["disk"]["normal"] += container.Usage.Storage().Value()
				}
			}
			updateMinMaxUsage(podStore)
		}
	}
	return nil
}

func resetNormalCount(podStore PodResStore) {
	podStore["request"]["cpu"]["normal"] = 0
	podStore["request"]["mem"]["normal"] = 0
	podStore["request"]["disk"]["normal"] = 0
	podStore["limit"]["cpu"]["normal"] = 0
	podStore["limit"]["mem"]["normal"] = 0
	podStore["limit"]["disk"]["normal"] = 0
	podStore["usage"]["cpu"]["normal"] = 0
	podStore["usage"]["mem"]["normal"] = 0
	podStore["usage"]["disk"]["normal"] = 0
}

func podStoreInit(podStore PodResStore) {
	if _, ok := podStore["request"]; !ok {
		podStore["request"] = make(map[string]map[string]int64)
	}
	if _, ok := podStore["request"]["cpu"]; !ok {
		podStore["request"]["cpu"] = make(map[string]int64)
	}
	if _, ok := podStore["request"]["mem"]; !ok {
		podStore["request"]["mem"] = make(map[string]int64)
	}
	if _, ok := podStore["request"]["disk"]; !ok {
		podStore["request"]["disk"] = make(map[string]int64)
	}

	if _, ok := podStore["limit"]; !ok {
		podStore["limit"] = make(map[string]map[string]int64)
	}
	if _, ok := podStore["limit"]["cpu"]; !ok {
		podStore["limit"]["cpu"] = make(map[string]int64)
	}
	if _, ok := podStore["limit"]["mem"]; !ok {
		podStore["limit"]["mem"] = make(map[string]int64)
	}
	if _, ok := podStore["limit"]["disk"]; !ok {
		podStore["limit"]["disk"] = make(map[string]int64)
	}

	if _, ok := podStore["usage"]; !ok {
		podStore["usage"] = make(map[string]map[string]int64)
	}
	if _, ok := podStore["usage"]["cpu"]; !ok {
		podStore["usage"]["cpu"] = make(map[string]int64)
	}
	if _, ok := podStore["usage"]["mem"]; !ok {
		podStore["usage"]["mem"] = make(map[string]int64)
	}
	if _, ok := podStore["usage"]["disk"]; !ok {
		podStore["usage"]["disk"] = make(map[string]int64)
	}
}

func updateMinMaxUsage(podStore PodResStore) {
	if podStore["usage"]["cpu"]["min"] == 0 {
		podStore["usage"]["cpu"]["min"] = podStore["usage"]["cpu"]["normal"]
	}
	if podStore["usage"]["cpu"]["max"] == 0 {
		podStore["usage"]["cpu"]["max"] = podStore["usage"]["cpu"]["normal"]
	}
	if podStore["usage"]["mem"]["min"] == 0 {
		podStore["usage"]["mem"]["min"] = podStore["usage"]["mem"]["normal"]
	}
	if podStore["usage"]["mem"]["max"] == 0 {
		podStore["usage"]["mem"]["max"] = podStore["usage"]["mem"]["normal"]
	}
	if podStore["usage"]["disk"]["min"] == 0 {
		podStore["usage"]["disk"]["min"] = podStore["usage"]["disk"]["normal"]
	}
	if podStore["usage"]["disk"]["max"] == 0 {
		podStore["usage"]["disk"]["max"] = podStore["usage"]["disk"]["normal"]
	}

	// update min max
	if podStore["usage"]["cpu"]["normal"] < podStore["usage"]["cpu"]["min"] {
		podStore["usage"]["cpu"]["min"] = podStore["usage"]["cpu"]["normal"]
	}
	if podStore["usage"]["cpu"]["normal"] > podStore["usage"]["cpu"]["max"] {
		podStore["usage"]["cpu"]["max"] = podStore["usage"]["cpu"]["normal"]
	}

	if podStore["usage"]["mem"]["normal"] < podStore["usage"]["mem"]["min"] {
		podStore["usage"]["mem"]["min"] = podStore["usage"]["mem"]["normal"]
	}
	if podStore["usage"]["mem"]["normal"] > podStore["usage"]["mem"]["max"] {
		podStore["usage"]["mem"]["max"] = podStore["usage"]["mem"]["normal"]
	}

	if podStore["usage"]["disk"]["normal"] < podStore["usage"]["disk"]["min"] {
		podStore["usage"]["disk"]["min"] = podStore["usage"]["disk"]["normal"]
	}
	if podStore["usage"]["disk"]["normal"] > podStore["usage"]["disk"]["max"] {
		podStore["usage"]["disk"]["max"] = podStore["usage"]["disk"]["normal"]
	}
}

func getNamespaces(allNamespaces []corev1.Namespace) []string {
	var usedNamespaceName []string
	var allNamespacesName []string
	usedNamespaceNamesSet := make(map[string]bool)

	for _, namespace := range allNamespaces {
		allNamespacesName = append(allNamespacesName, namespace.Name)
	}
	for _, namespace := range config.GetStringSlice("app.namespaces") {
		usedNamespaceNamesSet[namespace] = true
	}
	if _, ok := usedNamespaceNamesSet["all"]; ok {
		return allNamespacesName
	}
	for _, ns := range allNamespacesName {
		if _, ok := usedNamespaceNamesSet[ns]; ok {
			usedNamespaceName = append(usedNamespaceName, ns)
		}
	}
	return usedNamespaceName
}

func ExportPodRes(store AllPodResStore) {
	fmt.Println()
	for ns, pods := range store {
		for name, pod := range pods {
			fmt.Print(ns, ", ")
			fmt.Print(name, ", ")
			fmt.Print(pod["request"]["cpu"]["normal"], ", ")
			fmt.Print(pod["request"]["mem"]["normal"], ", ")
			fmt.Print(pod["request"]["disk"]["normal"], ", ")
			fmt.Print(pod["limit"]["cpu"]["normal"], ", ")
			fmt.Print(pod["limit"]["mem"]["normal"], ", ")
			fmt.Print(pod["limit"]["disk"]["normal"], ", ")
			fmt.Print(pod["usage"]["cpu"]["min"], ", ")
			fmt.Print(pod["usage"]["cpu"]["normal"], ", ")
			fmt.Print(pod["usage"]["cpu"]["max"], ", ")
			fmt.Print(pod["usage"]["mem"]["min"], ", ")
			fmt.Print(pod["usage"]["mem"]["normal"], ", ")
			fmt.Print(pod["usage"]["mem"]["max"], ", ")
			fmt.Print(pod["usage"]["disk"]["normal"], ", ")
			fmt.Println()
		}
	}
}
