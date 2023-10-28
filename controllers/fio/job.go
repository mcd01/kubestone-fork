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
	"github.com/firepear/qsplit"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	"github.com/xridge/kubestone/pkg/k8s"
)

// NewJob creates a fio benchmark job
func NewJob(cr *perfv1alpha1.Fio) *batchv1.Job {
	objectMeta := metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}

	volumes := []corev1.Volume{
		{Name: "data", VolumeSource: cr.Spec.Volume.VolumeSource},
	}
	volumeMounts := []corev1.VolumeMount{
		{Name: "data", MountPath: "/data"},
	}

	var cmdLineArgs []string
	cmdLineArgs = append(cmdLineArgs, qsplit.ToStrings([]byte(cr.Spec.CmdLineArgs))...)

	job := k8s.NewPerfJob(objectMeta, "fio", cr.Spec.PodConfig)
	job.Spec.Template.Spec.Volumes = volumes
	for i := 0; i < len(job.Spec.Template.Spec.InitContainers); i++ {
		job.Spec.Template.Spec.InitContainers[i].VolumeMounts = volumeMounts
		if job.Spec.Template.Spec.InitContainers[i].Args == nil {
			job.Spec.Template.Spec.InitContainers[i].Args = cmdLineArgs
		}
	}
	for i := 0; i < len(job.Spec.Template.Spec.Containers); i++ {
		job.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts
		if job.Spec.Template.Spec.Containers[i].Args == nil {
			job.Spec.Template.Spec.Containers[i].Args = cmdLineArgs
		}
	}
	return job
}

// IsCrValid validates the given CR and raises error if semantic errors detected
// For fio, the VolumeSpec validity is checked
func IsCrValid(cr *perfv1alpha1.Fio) (valid bool, err error) {
	return cr.Spec.Volume.Validate()
}
