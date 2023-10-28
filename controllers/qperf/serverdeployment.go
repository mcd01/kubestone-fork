/*
Copyright 2019 The xridge kubestone contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package qperf

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/util/intstr"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
)

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;create;delete;watch

func serverDeploymentName(cr *perfv1alpha1.Qperf) string {
	return cr.Name
}

// NewServerDeployment create a qperf server deployment from the
// provided Qperf Benchmark Definition.
func NewServerDeployment(cr *perfv1alpha1.Qperf) *appsv1.Deployment {
	replicas := int32(1)

	labels := map[string]string{
		"kubestone.xridge.io/app":     "qperf",
		"kubestone.xridge.io/cr-name": cr.Name,
	}
	// Let's be nice and don't mutate CRs label field
	for k, v := range cr.Spec.ServerConfiguration.PodConfig.PodLabels {
		labels[k] = v
	}

	cmdLineArgs := []string{"--listen_port", strconv.Itoa(perfv1alpha1.QperfPort)}

	// Qperf Server does not like if probe connections are made to the port,
	// therefore we are checking if the port if open or not via shell script
	// the solution does not assume to have netstat installed in the container
	readinessAwkCmd := fmt.Sprintf("BEGIN{err=1}toupper($2)~/:%04X$/{err=0}END{exit err}", perfv1alpha1.QperfPort)

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serverDeploymentName(cr),
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: cr.Spec.ServerConfiguration.PodConfig.InitContainers,
					Containers:     cr.Spec.ServerConfiguration.PodConfig.Containers,
					Affinity:       cr.Spec.ServerConfiguration.PodConfig.PodScheduling.Affinity,
					Tolerations:    cr.Spec.ServerConfiguration.PodConfig.PodScheduling.Tolerations,
					NodeSelector:   cr.Spec.ServerConfiguration.PodConfig.PodScheduling.NodeSelector,
					HostNetwork:    cr.Spec.HostNetwork,
				},
			},
		},
	}

	for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
		if deployment.Spec.Template.Spec.Containers[i].Name == "main" {
			deployment.Spec.Template.Spec.Containers[i].Command = []string{"qperf"}
			deployment.Spec.Template.Spec.Containers[i].Args = cmdLineArgs
			deployment.Spec.Template.Spec.Containers[i].Ports = []corev1.ContainerPort{
				{
					Name:          "qperf-server",
					ContainerPort: perfv1alpha1.QperfPort,
					Protocol:      corev1.ProtocolTCP,
				},
			}
			deployment.Spec.Template.Spec.Containers[i].ReadinessProbe = &corev1.Probe{
				Handler: corev1.Handler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"awk",
							readinessAwkCmd,
							"/proc/1/net/tcp",
							"/proc/1/net/tcp6",
						},
					},
				},
				InitialDelaySeconds: 5,
				TimeoutSeconds:      2,
				PeriodSeconds:       2,
			}
		}
	}

	return &deployment
}
