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

package fio

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/firepear/qsplit"
	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	"github.com/xridge/kubestone/controllers/common"
)

// NewJob creates a fio benchmark job
func NewJob(cr *perfv1alpha1.Fio) *batchv1.Job {
	objectMeta := metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}

	fioCmdLineArgs := []string{}
	fioCmdLineArgs = append(fioCmdLineArgs,
		qsplit.ToStrings([]byte(cr.Spec.CmdLineArgs))...)
	fioCmdLineArgs = append(fioCmdLineArgs, cr.Spec.BuiltinJobFiles...)
	for i := 0; i < len(cr.Spec.CustomJobFiles); i++ {
		fioCmdLineArgs = append(fioCmdLineArgs, "/custom-jobs/"+CustomJobName(i))
	}

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	if len(cr.Spec.CustomJobFiles) > 0 {
		volumes = append(volumes, corev1.Volume{
			Name: "custom-jobs",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cr.Name,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name: "custom-jobs", MountPath: "/custom-jobs",
		})
	}
	if cr.Spec.Volume != nil {
		volumeSource := cr.Spec.Volume.VolumeSource
		if cr.Spec.Volume.PersistentVolumeClaim != nil {
			// If a PVC was constructed, use that instead of the given volume source
			volumeSource = corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: cr.Name,
				},
			}
		}
		volumes = append(volumes, corev1.Volume{
			Name: "data", VolumeSource: volumeSource,
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name: "data", MountPath: "/data",
		})
	}

	job := common.NewPerfJob(objectMeta, "fio", cr.Spec.Image, cr.Spec.PodConfig)
	job.Spec.Template.Spec.Volumes = volumes
	job.Spec.Template.Spec.Containers[0].Args = fioCmdLineArgs
	job.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	return job
}
