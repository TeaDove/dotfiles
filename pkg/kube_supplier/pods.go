package kube_supplier

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"maps"
	"slices"
	"strings"
)

type ContainerInfo struct {
	Name    string
	PodName string

	Image            string
	State            string
	Ready            bool
	CpuUsageMicro    uint64
	MemUsageMegaByte uint64
}

func getPodContainerName(pod, container string) string {
	return fmt.Sprintf("%s-%s", pod, container)
}

func (r *Supplier) GetContainerInfo(ctx context.Context) ([]*ContainerInfo, error) {
	pods, err := r.kc.CoreV1().Pods(r.namespace).List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pods")
	}

	var containersInfo = make(map[string]*ContainerInfo, len(pods.Items))

	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			containerInfo := containersInfo[getPodContainerName(pod.Name, container.Name)]
			if containerInfo == nil {
				containerInfo = &ContainerInfo{
					PodName: pod.Name,
					Name:    container.Name,
					Image:   container.Image,
					Ready:   container.Ready,
				}
				containersInfo[getPodContainerName(pod.Name, container.Name)] = containerInfo
			}

			if container.State.Running != nil {
				containerInfo.State = "running"
				continue
			}
			if container.State.Waiting != nil {
				containerInfo.State = container.State.Waiting.Message
				continue
			}
			if container.State.Terminated != nil {
				containerInfo.State = container.State.Terminated.Message
				continue
			}
		}
	}

	podMetrics, err := r.mc.MetricsV1beta1().PodMetricses(r.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers

		for _, container := range podContainers {
			containerInfo, ok := containersInfo[getPodContainerName(podMetric.Name, container.Name)]
			if !ok {
				continue
			}

			containerInfo.CpuUsageMicro = uint64(container.Usage.Cpu().ScaledValue(resource.Micro))
			containerInfo.MemUsageMegaByte = uint64(container.Usage.Memory().ScaledValue(resource.Mega))
		}
	}

	return slices.SortedFunc(maps.Values(containersInfo), func(info *ContainerInfo, info2 *ContainerInfo) int {
		return strings.Compare(info.PodName, info2.PodName)
	}), nil
}
