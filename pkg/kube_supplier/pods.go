package kube_supplier

import (
	"cmp"
	"context"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"maps"
	"slices"
	"strings"
	"time"
)

type ContainerInfo struct {
	Name    string
	PodName string

	Image     string
	State     string
	Ready     bool
	CreatedAt time.Time

	CpuUsageMilli        uint64
	MemUsageKiloByte     uint64
	CpuRequestedMilli    uint64
	MemRequestedKiloByte uint64
}

func getPodContainerName(pod, container string) string {
	return fmt.Sprintf("%s-%s", pod, container)
}

func (r *Supplier) GetContainersInfo(ctx context.Context) ([]*ContainerInfo, error) {
	pods, err := r.kc.CoreV1().Pods(r.namespace).List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pods")
	}

	var containersInfo = make(map[string]*ContainerInfo, len(pods.Items))

	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {

			containerInfo := containersInfo[getPodContainerName(pod.Name, container.Name)]
			if containerInfo == nil {
				containerInfo = &ContainerInfo{}
				containersInfo[getPodContainerName(pod.Name, container.Name)] = containerInfo
			}

			containerInfo.Name = container.Name
			containerInfo.PodName = pod.Name
			containerInfo.Ready = container.Ready
			containerInfo.Image = container.Image
			containerInfo.CreatedAt = pod.CreationTimestamp.Time

			if len(pod.Status.Conditions) != 0 {
				slices.SortFunc(pod.Status.Conditions, func(a, b corev1.PodCondition) int {
					return cmp.Compare(b.LastTransitionTime.Unix(), a.LastTransitionTime.Unix())
				})

				containerInfo.State = string(pod.Status.Conditions[0].Type)
			}
		}

		for _, container := range pod.Spec.Containers {
			containerInfo := containersInfo[getPodContainerName(pod.Name, container.Name)]
			if containerInfo == nil {
				containerInfo = &ContainerInfo{}
				containersInfo[getPodContainerName(pod.Name, container.Name)] = containerInfo
			}

			if container.Resources.Requests.Memory() != nil {
				containerInfo.MemRequestedKiloByte = uint64(container.Resources.Requests.Memory().ScaledValue(resource.Kilo))
			}
			if container.Resources.Requests.Cpu() != nil {
				containerInfo.CpuRequestedMilli = uint64(container.Resources.Requests.Cpu().ScaledValue(resource.Milli))
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

			containerInfo.CpuUsageMilli = uint64(container.Usage.Cpu().ScaledValue(resource.Milli))
			containerInfo.MemUsageKiloByte = uint64(container.Usage.Memory().ScaledValue(resource.Kilo))
		}
	}

	return slices.SortedFunc(maps.Values(containersInfo), func(a *ContainerInfo, b *ContainerInfo) int {
		return strings.Compare(getPodContainerName(a.PodName, a.Name), getPodContainerName(b.PodName, b.Name))
	}), nil
}

type DeploymentInfo struct {
	Name string

	Images         []string
	ImageUpdatedAt time.Time
	PrevImages     []string

	Ready             bool
	ReadyReplicas     uint64
	RequestedReplicas uint64
	UpdatedAt         time.Time
}

func containersToImages(containers []corev1.Container) []string {
	images := make([]string, 0)
	for _, container := range containers {
		images = append(images, container.Image)
	}

	return images
}

func (r *Supplier) GetDeploymentInfo(ctx context.Context) ([]*DeploymentInfo, error) {
	deployments, err := r.kc.AppsV1().Deployments(r.namespace).List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pods")
	}

	var deploymentsInfo = make(map[string]*DeploymentInfo, len(deployments.Items))

	for _, deploy := range deployments.Items {
		deploymentInfo := DeploymentInfo{
			Name:              deploy.Name,
			Ready:             deploy.Status.ReadyReplicas == deploy.Status.Replicas,
			ReadyReplicas:     uint64(deploy.Status.ReadyReplicas),
			RequestedReplicas: uint64(deploy.Status.Replicas),
			UpdatedAt:         deploy.CreationTimestamp.Time,
			Images:            containersToImages(deploy.Spec.Template.Spec.Containers),
		}
		deploymentsInfo[deploy.Name] = &deploymentInfo

		replicasSet, err := r.kc.AppsV1().ReplicaSets(r.namespace).List(ctx, metav1.ListOptions{Limit: 30, LabelSelector: fmt.Sprintf("app=%s", deploy.Name)})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get replicasets")
		}

		for idx, replicaSet := range replicasSet.Items {
			images := containersToImages(replicaSet.Spec.Template.Spec.Containers)
			if strings.Join(images, ",") == strings.Join(deploymentInfo.Images, ",") {
				continue
			}

			deploymentInfo.PrevImages = containersToImages(replicaSet.Spec.Template.Spec.Containers)
			if idx == 0 {
				deploymentInfo.ImageUpdatedAt = deploy.CreationTimestamp.Time
			} else {
				deploymentInfo.ImageUpdatedAt = replicasSet.Items[idx-1].CreationTimestamp.Time
			}
			break
		}
	}

	return slices.SortedFunc(maps.Values(deploymentsInfo), func(a *DeploymentInfo, b *DeploymentInfo) int {
		return strings.Compare(a.Name, b.Name)
	}), nil
}
