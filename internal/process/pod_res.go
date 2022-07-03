package process

import (
	"context"
	"fmt"
	k8client "k8res/internal/k8s/client"
	"k8res/pkg/config"
	"k8res/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes/typed/core/v1"
)

//PodResStore ex: [pod-name][request][cpu] = int64
type PodResStore map[string]map[string]map[string]int64

func GetPodRes(k8 *k8client.K8s, store PodResStore) error {
	var podClient client.PodInterface
	var pvcClient client.PersistentVolumeClaimInterface
	namespaces := config.GetStringSlice("app.namespaces")
	ctx := context.TODO()


	for _, ns := range namespaces {
		if ns == "all" {
			podClient = k8.ClientSet.CoreV1().Pods(metav1.NamespaceAll)
			pvcClient = k8.ClientSet.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll)
		} else {
			podClient = k8.ClientSet.CoreV1().Pods(ns)
			pvcClient = k8.ClientSet.CoreV1().PersistentVolumeClaims(ns)
		}
		pods, err := podClient.List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != "Running" {
				logger.Infof("pod %s is not running", pod.Name)
				continue
			}
			// requests & limits
			if _, ok := store[pod.Name]; !ok {
				store[pod.Name] = make(map[string]map[string]int64)
			}
			for _, container := range pod.Spec.Containers {
				if _, ok := store[pod.Name]["request"]; !ok {
					store[pod.Name]["request"] = make(map[string]int64)
				}
				if _, ok := store[pod.Name]["limit"]; !ok {
					store[pod.Name]["limit"] = make(map[string]int64)
				}

				if container.Resources.Requests.Cpu() != nil {
					store[pod.Name]["request"]["cpu"] += container.Resources.Requests.Cpu().Value()
				}
				if container.Resources.Requests.Memory() != nil {
					store[pod.Name]["request"]["mem"] += container.Resources.Requests.Memory().Value()
				}
				if container.Resources.Limits.Cpu() != nil {
					store[pod.Name]["limit"]["cpu"] += container.Resources.Limits.Cpu().Value()
				}
				if container.Resources.Limits.Memory() != nil {
					store[pod.Name]["limit"]["mem"] += container.Resources.Limits.Memory().Value()
				}
			}

			// disk
			for _, volume := range pod.Spec.Volumes {
				if volume.VolumeSource.PersistentVolumeClaim != nil {
					pvc, err := pvcClient.Get(ctx, volume.VolumeSource.PersistentVolumeClaim.ClaimName, metav1.GetOptions{});
					if err != nil {
						return err
					}
					if pvc.Spec.Resources.Requests.Storage() != nil {
						store[pod.Name]["request"]["disk"] += pvc.Spec.Resources.Requests.Storage().Value()
					}
					if pvc.Spec.Resources.Limits.Storage() != nil {
						store[pod.Name]["limit"]["disk"] += pvc.Spec.Resources.Limits.Storage().Value()
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
				if _, ok := store[pod.Name]["usage"]; !ok {
					store[pod.Name]["usage"] = make(map[string]int64)
				}

				if container.Usage.Cpu() != nil {
					store[pod.Name]["usage"]["cpu"] += container.Usage.Cpu().Value()
				}
				if container.Usage.Memory() != nil {
					store[pod.Name]["usage"]["mem"] += container.Usage.Memory().Value()
				}
				if container.Usage.Storage() != nil {
					store[pod.Name]["usage"]["disk"] += container.Usage.Storage().Value()
				}
			}
		}
	}
	return nil
}

func ExportPodRes(store PodResStore) {
	fmt.Println()
	for pod, res := range store {
		fmt.Print(pod, ", ")
		fmt.Print(res["request"]["cpu"], ", ")
		fmt.Print(res["request"]["mem"], ", ")
		fmt.Print(res["request"]["disk"], ", ")
		fmt.Print(res["limit"]["cpu"], ", ")
		fmt.Print(res["limit"]["mem"], ", ")
		fmt.Print(res["limit"]["disk"], ", ")
		fmt.Print(res["usage"]["cpu"], ", ")
		fmt.Print(res["usage"]["mem"], ", ")
		fmt.Print(res["usage"]["disk"])
		fmt.Println()
	}
}