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
package iperf3

import (
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ksapi "github.com/xridge/kubestone/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Server Deployment", func() {
	Describe("created from CR", func() {
		var cr ksapi.Iperf3
		var deployment *appsv1.Deployment

		BeforeEach(func() {
			tolerationSeconds := int64(17)
			cr = ksapi.Iperf3{
				Spec: ksapi.Iperf3Spec{
					HostNetwork: true,
					ServerConfiguration: ksapi.BenchmarkConfigurationSpec{
						CmdLineArgs: "--testing --things",
						PodConfig: ksapi.PodConfigurationSpec{
							PodLabels: map[string]string{"labels": "are", "really": "useful"},
							PodScheduling: ksapi.PodSchedulingSpec{
								Affinity: &corev1.Affinity{
									NodeAffinity: &corev1.NodeAffinity{
										RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
											NodeSelectorTerms: []corev1.NodeSelectorTerm{
												{
													MatchExpressions: []corev1.NodeSelectorRequirement{
														{
															Key:      "mutated",
															Operator: corev1.NodeSelectorOperator(corev1.NodeSelectorOpIn),
															Values:   []string{"nano-virus"},
														},
													},
												},
											},
										},
									},
								},
								Tolerations: []corev1.Toleration{
									{
										Key:               "genetic-code",
										Operator:          corev1.TolerationOperator(corev1.TolerationOpExists),
										Value:             "distressed",
										Effect:            corev1.TaintEffect(corev1.TaintEffectNoExecute),
										TolerationSeconds: &tolerationSeconds,
									},
								},
								NodeSelector: map[string]string{
									"atomized": "spiral",
								},
							},
						},
					},
				},
			}
			deployment = NewServerDeployment(&cr)
		})

		Context("with default settings", func() {
			It("--server mode is enabled", func() {
				Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(
					ContainElement("--server"))
			})
			It("--port's value is specified", func() {
				Expect(strings.Join(deployment.Spec.Template.Spec.Containers[0].Args, " ")).To(
					ContainSubstring("--port " + strconv.Itoa(Iperf3ServerPort)))
			})
			It("should not contain --udp flag", func() {
				Expect(deployment.Spec.Template.Spec.Containers[0].Args).NotTo(
					ContainElement("--udp"))
			})
		})

		Context("with UDP mode specified", func() {
			cr.Spec.UDP = true
			deployment := NewServerDeployment(&cr)
			It("should contain --udp flag in iperf args", func() {
				Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(
					ContainElement("--udp"))
			})
		})

		Context("with podLabels specified", func() {
			It("should contain all podLabels", func() {
				for k, v := range cr.Spec.ServerConfiguration.PodConfig.PodLabels {
					Expect(deployment.Spec.Template.ObjectMeta.Labels).To(
						HaveKeyWithValue(k, v))
				}
			})
		})

		Context("with podAffinity specified", func() {
			It("should match with Affinity", func() {
				Expect(deployment.Spec.Template.Spec.Affinity).To(
					Equal(cr.Spec.ServerConfiguration.PodConfig.PodScheduling.Affinity))
			})
			It("should match with Tolerations", func() {
				Expect(deployment.Spec.Template.Spec.Tolerations).To(
					Equal(cr.Spec.ServerConfiguration.PodConfig.PodScheduling.Tolerations))
			})
			It("should match with NodeSelector", func() {
				Expect(deployment.Spec.Template.Spec.NodeSelector).To(
					Equal(cr.Spec.ServerConfiguration.PodConfig.PodScheduling.NodeSelector))
			})
		})

		Context("with HostNetwork specified", func() {
			It("should match with HostNetwork", func() {
				Expect(deployment.Spec.Template.Spec.HostNetwork).To(
					Equal(cr.Spec.HostNetwork))
			})
		})
	})
})
