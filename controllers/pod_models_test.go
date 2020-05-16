/*
 * pod_models_test.go
 *
 * This source file is part of the FoundationDB open source project
 *
 * Copyright 2019 Apple Inc. and the FoundationDB project authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	"fmt"

	fdbtypes "github.com/FoundationDB/fdb-kubernetes-operator/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("pod_models", func() {
	var cluster *fdbtypes.FoundationDBCluster
	var err error

	BeforeEach(func() {
		cluster = createDefaultCluster()
	})

	Describe("GetPod", func() {
		var pod *corev1.Pod
		Context("with a basic storage instance", func() {
			BeforeEach(func() {
				pod, err = GetPod(context.TODO(), cluster, "storage", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should contain the instance's metadata", func() {
				Expect(pod.Namespace).To(Equal("my-ns"))
				Expect(pod.Name).To(Equal(fmt.Sprintf("%s-storage-1", cluster.Name)))
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "storage-1",
				}))
			})

			It("should contain the instance's pod spec", func() {
				spec, err := GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Spec).To(Equal(*spec))
			})
		})

		Context("with a cluster controller instance", func() {
			BeforeEach(func() {
				pod, err = GetPod(context.TODO(), cluster, "cluster_controller", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should contain the instance's metadata", func() {
				Expect(pod.Name).To(Equal(fmt.Sprintf("%s-cluster-controller-1", cluster.Name)))
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "cluster_controller",
					"fdb-instance-id":   "cluster_controller-1",
				}))
			})

			It("should contain the instance's pod spec", func() {
				spec, err := GetPodSpec(cluster, "cluster_controller", 1)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Spec).To(Equal(*spec))
			})
		})

		Context("with an instance ID prefix", func() {
			BeforeEach(func() {
				cluster.Spec.InstanceIDPrefix = "dc1"
				pod, err = GetPod(context.TODO(), cluster, "storage", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not include the prefix in the instance name", func() {
				Expect(pod.Name).To(Equal(fmt.Sprintf("%s-storage-1", cluster.Name)))
			})

			It("should contain the prefix in the instance labels labels", func() {
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "dc1-storage-1",
				}))
			})
		})

		Context("with custom annotations", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {PodTemplate: &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"fdb-annotation": "value1",
						},
					},
				}}}
				pod, err = GetPod(context.TODO(), cluster, "storage", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should add the annotations to the metadata", func() {
				hash, err := GetPodSpecHash(cluster, pod.Labels["fdb-process-class"], 1, &pod.Spec)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.ObjectMeta.Annotations).To(Equal(map[string]string{
					"fdb-annotation":                     "value1",
					"foundationdb.org/last-applied-spec": hash,
				}))
			})
		})

		Context("with custom annotations from the spec.PodTemplate field", func() {
			BeforeEach(func() {
				cluster.Spec.PodTemplate = &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"fdb-annotation": "value1",
						},
					},
				}
				pod, err = GetPod(context.TODO(), cluster, "storage", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should add the annotations to the metadata", func() {
				hash, err := GetPodSpecHash(cluster, pod.Labels["fdb-process-class"], 1, &pod.Spec)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.ObjectMeta.Annotations).To(Equal(map[string]string{
					"fdb-annotation":                     "value1",
					"foundationdb.org/last-applied-spec": hash,
				}))
			})
		})

		Context("with custom labels", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {PodTemplate: &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"fdb-label": "value2",
						},
					},
				}}}
				pod, err = GetPod(context.TODO(), cluster, "storage", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should add the labels to the metadata", func() {
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "storage-1",
					"fdb-label":         "value2",
				}))
			})
		})

		Context("with custom labels from the Spec.PodTemplate field", func() {
			BeforeEach(func() {
				cluster.Spec.PodTemplate = &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"fdb-label": "value2",
						},
					},
				}
				pod, err = GetPod(context.TODO(), cluster, "storage", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should add the labels to the metadata", func() {
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "storage-1",
					"fdb-label":         "value2",
				}))
			})
		})

		Context("with custom labels from the Spec.PodLabels field", func() {
			BeforeEach(func() {
				cluster.Spec.PodLabels = map[string]string{
					"fdb-label": "value3",
				}

				pod, err = GetPod(context.TODO(), cluster, "storage", 1, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should put the labels on the pod", func() {
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "storage-1",
					"fdb-label":         "value3",
				}))

			})
		})
	})

	Describe("GetPodSpec", func() {
		var spec *corev1.PodSpec

		Context("with a basic storage instance", func() {
			BeforeEach(func() {
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should have the built-in init container", func() {
				Expect(len(spec.InitContainers)).To(Equal(1))
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
				Expect(initContainer.Args).To(Equal([]string{
					"--copy-file",
					"fdb.cluster",
					"--input-monitor-conf",
					"fdbmonitor.conf",
					"--copy-binary",
					"fdbserver",
					"--copy-binary",
					"fdbcli",
					"--main-container-version",
					"6.2.20",
					"--init-mode",
				}))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))
				Expect(initContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "config-map", MountPath: "/var/input-files"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/output-files"},
				}))
				Expect(initContainer.ReadinessProbe).To(BeNil())
			})

			It("should have two containers", func() {
				Expect(len(spec.Containers)).To(Equal(2))
			})

			It("should have the main foundationdb container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb:%s", cluster.Spec.Version)))
				Expect(mainContainer.Command).To(Equal([]string{"sh", "-c"}))
				Expect(mainContainer.Args).To(Equal([]string{
					"fdbmonitor --conffile /var/dynamic-conf/fdbmonitor.conf" +
						" --lockfile /var/dynamic-conf/fdbmonitor.lockfile",
				}))

				Expect(mainContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
				}))

				Expect(*mainContainer.Resources.Limits.Cpu()).To(Equal(resource.MustParse("1")))
				Expect(*mainContainer.Resources.Limits.Memory()).To(Equal(resource.MustParse("1Gi")))
				Expect(*mainContainer.Resources.Requests.Cpu()).To(Equal(resource.MustParse("1")))
				Expect(*mainContainer.Resources.Requests.Memory()).To(Equal(resource.MustParse("1Gi")))

				Expect(len(mainContainer.VolumeMounts)).To(Equal(3))

				Expect(mainContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "data", MountPath: "/var/fdb/data"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/dynamic-conf"},
					corev1.VolumeMount{Name: "fdb-trace-logs", MountPath: "/var/log/fdb-trace-logs"},
				}))

				Expect(*mainContainer.SecurityContext.ReadOnlyRootFilesystem).To(BeTrue())
			})

			It("should have the sidecar container", func() {
				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
				Expect(sidecarContainer.Args).To(Equal([]string{
					"--copy-file",
					"fdb.cluster",
					"--input-monitor-conf",
					"fdbmonitor.conf",
					"--copy-binary",
					"fdbserver",
					"--copy-binary",
					"fdbcli",
					"--main-container-version",
					"6.2.20",
				}))
				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: ""},
				}))
				Expect(sidecarContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "config-map", MountPath: "/var/input-files"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/output-files"},
				}))
				Expect(sidecarContainer.ReadinessProbe).To(Equal(&corev1.Probe{
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.IntOrString{IntVal: 8080},
						},
					},
				}))

				Expect(*sidecarContainer.SecurityContext.ReadOnlyRootFilesystem).To(BeTrue())
			})

			It("should have the built-in volumes", func() {
				Expect(len(spec.Volumes)).To(Equal(4))
				Expect(spec.Volumes[0]).To(Equal(corev1.Volume{
					Name: "data",
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("%s-storage-1-data", cluster.Name),
					}},
				}))
				Expect(spec.Volumes[1]).To(Equal(corev1.Volume{
					Name:         "dynamic-conf",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
				Expect(spec.Volumes[2]).To(Equal(corev1.Volume{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{Key: "fdbmonitor-conf-storage", Path: "fdbmonitor.conf"},
							corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
						},
					}},
				}))
				Expect(spec.Volumes[3]).To(Equal(corev1.Volume{
					Name:         "fdb-trace-logs",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
			})

			It("should have no affinity rules", func() {
				Expect(spec.Affinity).To(BeNil())
			})
		})

		Context("with custom resources", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {PodTemplate: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse("2"),
										"memory": resource.MustParse("8Gi"),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse("4"),
										"memory": resource.MustParse("16Gi"),
									},
								},
							},
						},
					},
				}}}
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should set the resources on the main container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(*mainContainer.Resources.Limits.Cpu()).To(Equal(resource.MustParse("4")))
				Expect(*mainContainer.Resources.Limits.Memory()).To(Equal(resource.MustParse("16Gi")))
				Expect(*mainContainer.Resources.Requests.Cpu()).To(Equal(resource.MustParse("2")))
				Expect(*mainContainer.Resources.Requests.Memory()).To(Equal(resource.MustParse("8Gi")))
			})
		})

		Context("with custom resources from the Spec.PodTemplate field", func() {
			BeforeEach(func() {
				cluster.Spec.PodTemplate = &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse("2"),
										"memory": resource.MustParse("8Gi"),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse("4"),
										"memory": resource.MustParse("16Gi"),
									},
								},
							},
						},
					},
				}
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should set the resources on the main container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(*mainContainer.Resources.Limits.Cpu()).To(Equal(resource.MustParse("4")))
				Expect(*mainContainer.Resources.Limits.Memory()).To(Equal(resource.MustParse("16Gi")))
				Expect(*mainContainer.Resources.Requests.Cpu()).To(Equal(resource.MustParse("2")))
				Expect(*mainContainer.Resources.Requests.Memory()).To(Equal(resource.MustParse("8Gi")))
			})
		})

		Context("with custom resources from the Spec.Resources field", func() {
			BeforeEach(func() {
				cluster.Spec.Resources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("2"),
						"memory": resource.MustParse("8Gi"),
					},
					Limits: corev1.ResourceList{
						"cpu":    resource.MustParse("4"),
						"memory": resource.MustParse("16Gi"),
					},
				}
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should set the resources on the main container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(*mainContainer.Resources.Limits.Cpu()).To(Equal(resource.MustParse("4")))
				Expect(*mainContainer.Resources.Limits.Memory()).To(Equal(resource.MustParse("16Gi")))
				Expect(*mainContainer.Resources.Requests.Cpu()).To(Equal(resource.MustParse("2")))
				Expect(*mainContainer.Resources.Requests.Memory()).To(Equal(resource.MustParse("8Gi")))
			})
		})

		Context("with no volume", func() {
			BeforeEach(func() {
				cluster.Spec.VolumeSize = "0"
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should use an EmptyDir volume", func() {
				Expect(spec.Volumes[0]).To(Equal(corev1.Volume{
					Name:         "data",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
			})
		})

		Context("with a host-based fault domain", func() {
			BeforeEach(func() {
				cluster.Spec.FaultDomain = fdbtypes.FoundationDBClusterFaultDomain{}
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should set the fault domain information in the sidecar environment", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))

			})

			It("should set the pod affinity", func() {
				Expect(spec.Affinity).To(Equal(&corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							corev1.WeightedPodAffinityTerm{
								Weight: 1,
								PodAffinityTerm: corev1.PodAffinityTerm{
									TopologyKey: "kubernetes.io/hostname",
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"fdb-cluster-name":  cluster.Name,
											"fdb-process-class": "storage",
										},
									},
								},
							},
						},
					},
				}))
			})
		})

		Context("with a custom fault domain", func() {
			BeforeEach(func() {

				cluster.Spec.FaultDomain = fdbtypes.FoundationDBClusterFaultDomain{
					Key:       "rack",
					ValueFrom: "$RACK",
				}
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should set the fault domain information in the sidecar environment", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))

			})

			It("should set the pod affinity", func() {
				Expect(spec.Affinity).To(Equal(&corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							corev1.WeightedPodAffinityTerm{
								Weight: 1,
								PodAffinityTerm: corev1.PodAffinityTerm{
									TopologyKey: "rack",
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"fdb-cluster-name":  cluster.Name,
											"fdb-process-class": "storage",
										},
									},
								},
							},
						},
					},
				}))
			})
		})

		Context("with cross-Kubernetes replication", func() {
			BeforeEach(func() {
				cluster.Spec.FaultDomain = fdbtypes.FoundationDBClusterFaultDomain{
					Key:   "foundationdb.org/kubernetes-cluster",
					Value: "kc2",
				}
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should set the fault domain information in the sidecar environment", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", Value: "kc2"},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))
			})

			It("should leave the pod affinity empty", func() {
				Expect(spec.Affinity).To(BeNil())
			})
		})

		Context("with custom containers", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {PodTemplate: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{corev1.Container{
							Name:    "test-container",
							Image:   "foundationdb/" + cluster.Name,
							Command: []string{"echo", "test1"},
						}},
						Containers: []corev1.Container{corev1.Container{
							Name:    "test-container",
							Image:   "foundationdb/" + cluster.Name,
							Command: []string{"echo", "test2"},
						}},
					},
				}}}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have both init containers", func() {
				Expect(len(spec.InitContainers)).To(Equal(2))

				testInitContainer := spec.InitContainers[0]
				Expect(testInitContainer.Name).To(Equal("test-container"))
				Expect(testInitContainer.Image).To(Equal("foundationdb/" + cluster.Name))
				Expect(testInitContainer.Command).To(Equal([]string{"echo", "test1"}))

				initContainer := spec.InitContainers[1]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
			})

			It("should have all three containers", func() {
				Expect(len(spec.Containers)).To(Equal(3))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb:%s", cluster.Spec.Version)))

				testContainer := spec.Containers[1]
				Expect(testContainer.Name).To(Equal("test-container"))
				Expect(testContainer.Image).To(Equal("foundationdb/" + cluster.Name))
				Expect(testContainer.Command).To(Equal([]string{"echo", "test2"}))

				sidecarContainer := spec.Containers[2]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
			})
		})

		Context("with custom containers from the Spec.PodTemplate field", func() {
			BeforeEach(func() {
				cluster.Spec.PodTemplate = &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{corev1.Container{
							Name:    "test-container",
							Image:   "foundationdb/" + cluster.Name,
							Command: []string{"echo", "test1"},
						}},
						Containers: []corev1.Container{corev1.Container{
							Name:    "test-container",
							Image:   "foundationdb/" + cluster.Name,
							Command: []string{"echo", "test2"},
						}},
					},
				}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have both init containers", func() {
				Expect(len(spec.InitContainers)).To(Equal(2))

				testInitContainer := spec.InitContainers[0]
				Expect(testInitContainer.Name).To(Equal("test-container"))
				Expect(testInitContainer.Image).To(Equal("foundationdb/" + cluster.Name))
				Expect(testInitContainer.Command).To(Equal([]string{"echo", "test1"}))

				initContainer := spec.InitContainers[1]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
			})

			It("should have all three containers", func() {
				Expect(len(spec.Containers)).To(Equal(3))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb:%s", cluster.Spec.Version)))

				testContainer := spec.Containers[1]
				Expect(testContainer.Name).To(Equal("test-container"))
				Expect(testContainer.Image).To(Equal("foundationdb/" + cluster.Name))
				Expect(testContainer.Command).To(Equal([]string{"echo", "test2"}))

				sidecarContainer := spec.Containers[2]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
			})
		})

		Context("with custom containers from the Spec.Containers field", func() {
			BeforeEach(func() {
				cluster.Spec.InitContainers = []corev1.Container{corev1.Container{
					Name:    "test-container",
					Image:   "foundationdb/" + cluster.Name,
					Command: []string{"echo", "test1"},
				}}
				cluster.Spec.Containers = []corev1.Container{corev1.Container{
					Name:    "test-container",
					Image:   "foundationdb/" + cluster.Name,
					Command: []string{"echo", "test2"},
				}}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have both init containers", func() {
				Expect(len(spec.InitContainers)).To(Equal(2))
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))

				testInitContainer := spec.InitContainers[1]
				Expect(testInitContainer.Name).To(Equal("test-container"))
				Expect(testInitContainer.Image).To(Equal("foundationdb/" + cluster.Name))
				Expect(testInitContainer.Command).To(Equal([]string{"echo", "test1"}))
			})

			It("should have all three containers", func() {
				Expect(len(spec.Containers)).To(Equal(3))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb:%s", cluster.Spec.Version)))

				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))

				testContainer := spec.Containers[2]
				Expect(testContainer.Name).To(Equal("test-container"))
				Expect(testContainer.Image).To(Equal("foundationdb/" + cluster.Name))
				Expect(testContainer.Command).To(Equal([]string{"echo", "test2"}))
			})
		})

		Context("with custom environment", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {PodTemplate: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb",
								Env: []corev1.EnvVar{
									corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
									corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/secrets/cert.pem"},
									corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
								},
							},
							corev1.Container{
								Name: "foundationdb-kubernetes-sidecar",
								Env: []corev1.EnvVar{
									corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env"},
								},
							},
						},
						InitContainers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb-kubernetes-init",
								Env: []corev1.EnvVar{
									corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env-init"},
								},
							},
						},
					},
				}}}

				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the environment variables on the containers", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env-init"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
				}))

				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: ""},
				}))
			})
		})

		Context("with custom environment from the Spec.PodTemplate field", func() {
			BeforeEach(func() {
				cluster.Spec.PodTemplate = &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb",
								Env: []corev1.EnvVar{
									corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
									corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/secrets/cert.pem"},
									corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
								},
							},
							corev1.Container{
								Name: "foundationdb-kubernetes-sidecar",
								Env: []corev1.EnvVar{
									corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env"},
								},
							},
						},
						InitContainers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb-kubernetes-init",
								Env: []corev1.EnvVar{
									corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env-init"},
								},
							},
						},
					},
				}

				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the environment variables on the containers", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env-init"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
				}))

				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: ""},
				}))
			})
		})

		Context("with custom environment from the Spec.MainContainer field", func() {
			BeforeEach(func() {
				cluster.Spec.MainContainer.Env = []corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
				}
				cluster.Spec.SidecarContainer.Env = []corev1.EnvVar{
					corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env"},
				}

				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the environment on the containers", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
				}))

				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "ADDITIONAL_ENV_FILE", Value: "/var/custom-env"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: ""},
				}))
			})
		})

		Context("with TLS for the sidecar", func() {
			BeforeEach(func() {
				cluster.Spec.SidecarContainer.EnableTLS = true
				cluster.Spec.SidecarContainer.Env = []corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
				}

				cluster.Spec.SidecarContainer.PeerVerificationRules = "S.CN=foundationdb.org"

				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("passes the TLS environment to the init container", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Args).To(Equal([]string{
					"--copy-file",
					"fdb.cluster",
					"--input-monitor-conf",
					"fdbmonitor.conf",
					"--copy-binary",
					"fdbserver",
					"--copy-binary",
					"fdbcli",
					"--main-container-version",
					"6.2.20",
					"--init-mode",
				}))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))
			})

			It("passes the TLS environment to the sidecar", func() {
				Expect(len(spec.Containers)).To(Equal(2))

				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Args).To(Equal([]string{
					"--copy-file",
					"fdb.cluster",
					"--input-monitor-conf",
					"fdbmonitor.conf",
					"--copy-binary",
					"fdbserver",
					"--copy-binary",
					"fdbcli",
					"--main-container-version",
					cluster.Spec.Version,
					"--tls",
				}))
				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_TLS_CERTIFICATE_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_TLS_KEY_FILE", Value: "/var/secrets/cert.pem"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: "S.CN=foundationdb.org"},
				}))
			})
		})

		Context("with custom volumes", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {PodTemplate: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{corev1.Volume{
							Name: "test-secrets",
							VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
								SecretName: "test-secrets",
							}},
						}},
						Containers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb",
								VolumeMounts: []corev1.VolumeMount{corev1.VolumeMount{
									Name:      "test-secrets",
									MountPath: "/var/secrets",
								}},
							},
						},
					},
				}}}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds volumes to the container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "test-secrets", MountPath: "/var/secrets"},
					corev1.VolumeMount{Name: "data", MountPath: "/var/fdb/data"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/dynamic-conf"},
					corev1.VolumeMount{Name: "fdb-trace-logs", MountPath: "/var/log/fdb-trace-logs"},
				}))
			})

			It("does not add volumes to the sidecar container", func() {
				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))

				Expect(sidecarContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "config-map", MountPath: "/var/input-files"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/output-files"},
				}))

			})

			It("adds volumes to the pod spec", func() {
				Expect(len(spec.Volumes)).To(Equal(5))
				Expect(spec.Volumes[0]).To(Equal(corev1.Volume{
					Name: "test-secrets",
					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
						SecretName: "test-secrets",
					}},
				}))
				Expect(spec.Volumes[1]).To(Equal(corev1.Volume{
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("%s-storage-1-data", cluster.Name),
					}},
					Name: "data",
				}))
				Expect(spec.Volumes[2]).To(Equal(corev1.Volume{
					Name:         "dynamic-conf",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
				Expect(spec.Volumes[3]).To(Equal(corev1.Volume{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{Key: "fdbmonitor-conf-storage", Path: "fdbmonitor.conf"},
							corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
						},
					}},
				}))
				Expect(spec.Volumes[4]).To(Equal(corev1.Volume{
					Name:         "fdb-trace-logs",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
			})
		})

		Context("with custom volumes from the Spec.PodTemplate field", func() {
			BeforeEach(func() {
				cluster.Spec.PodTemplate = &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{corev1.Volume{
							Name: "test-secrets",
							VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
								SecretName: "test-secrets",
							}},
						}},
						Containers: []corev1.Container{
							corev1.Container{
								Name: "foundationdb",
								VolumeMounts: []corev1.VolumeMount{corev1.VolumeMount{
									Name:      "test-secrets",
									MountPath: "/var/secrets",
								}},
							},
						},
					},
				}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds volumes to the container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "test-secrets", MountPath: "/var/secrets"},
					corev1.VolumeMount{Name: "data", MountPath: "/var/fdb/data"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/dynamic-conf"},
					corev1.VolumeMount{Name: "fdb-trace-logs", MountPath: "/var/log/fdb-trace-logs"},
				}))
			})

			It("does not add volumes to the sidecar container", func() {
				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))

				Expect(sidecarContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "config-map", MountPath: "/var/input-files"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/output-files"},
				}))

			})

			It("adds volumes to the pod spec", func() {
				Expect(len(spec.Volumes)).To(Equal(5))
				Expect(spec.Volumes[0]).To(Equal(corev1.Volume{
					Name: "test-secrets",
					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
						SecretName: "test-secrets",
					}},
				}))
				Expect(spec.Volumes[1]).To(Equal(corev1.Volume{
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("%s-storage-1-data", cluster.Name),
					}},
					Name: "data",
				}))
				Expect(spec.Volumes[2]).To(Equal(corev1.Volume{
					Name:         "dynamic-conf",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
				Expect(spec.Volumes[3]).To(Equal(corev1.Volume{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{Key: "fdbmonitor-conf-storage", Path: "fdbmonitor.conf"},
							corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
						},
					}},
				}))
				Expect(spec.Volumes[4]).To(Equal(corev1.Volume{
					Name:         "fdb-trace-logs",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
			})
		})

		Context("with custom volumes from the Spec.Volumes field", func() {
			BeforeEach(func() {
				cluster.Spec.Volumes = []corev1.Volume{corev1.Volume{
					Name: "test-secrets",
					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
						SecretName: "test-secrets",
					}},
				}}
				cluster.Spec.MainContainer.VolumeMounts = []corev1.VolumeMount{corev1.VolumeMount{
					Name:      "test-secrets",
					MountPath: "/var/secrets",
				}}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds volumes to the container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "data", MountPath: "/var/fdb/data"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/dynamic-conf"},
					corev1.VolumeMount{Name: "fdb-trace-logs", MountPath: "/var/log/fdb-trace-logs"},
					corev1.VolumeMount{Name: "test-secrets", MountPath: "/var/secrets"},
				}))
			})

			It("does not add volumes to the sidecar container", func() {
				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))

				Expect(sidecarContainer.VolumeMounts).To(Equal([]corev1.VolumeMount{
					corev1.VolumeMount{Name: "config-map", MountPath: "/var/input-files"},
					corev1.VolumeMount{Name: "dynamic-conf", MountPath: "/var/output-files"},
				}))

			})

			It("adds volumes to the pod spec", func() {
				Expect(len(spec.Volumes)).To(Equal(5))
				Expect(spec.Volumes[0]).To(Equal(corev1.Volume{
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("%s-storage-1-data", cluster.Name),
					}},
					Name: "data",
				}))
				Expect(spec.Volumes[1]).To(Equal(corev1.Volume{
					Name:         "dynamic-conf",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
				Expect(spec.Volumes[2]).To(Equal(corev1.Volume{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{Key: "fdbmonitor-conf-storage", Path: "fdbmonitor.conf"},
							corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
						},
					}},
				}))
				Expect(spec.Volumes[3]).To(Equal(corev1.Volume{
					Name:         "fdb-trace-logs",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}))
				Expect(spec.Volumes[4]).To(Equal(corev1.Volume{
					Name: "test-secrets",
					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
						SecretName: "test-secrets",
					}},
				}))
			})
		})

		Context("with custom sidecar version", func() {
			BeforeEach(func() {
				cluster.Spec.SidecarVersions = map[string]int{
					cluster.Spec.Version: 2,
					"6.1.0":              3,
				}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("sets the images on the containers", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-2", cluster.Spec.Version)))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb:%s", cluster.Spec.Version)))

				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Image).To(Equal(initContainer.Image))
			})
		})

		Context("with custom sidecar version with deprecated field", func() {
			BeforeEach(func() {
				cluster.Spec.SidecarVersion = 2
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("sets the images on the containers", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-2", cluster.Spec.Version)))

				mainContainer := spec.Containers[0]
				Expect(mainContainer.Name).To(Equal("foundationdb"))
				Expect(mainContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb:%s", cluster.Spec.Version)))

				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Image).To(Equal(initContainer.Image))
			})
		})

		Context("with a custom security context", func() {
			BeforeEach(func() {

				podSecurityContext := &corev1.PodSecurityContext{FSGroup: new(int64)}
				*podSecurityContext.FSGroup = 5000
				mainSecurityContext := &corev1.SecurityContext{RunAsGroup: new(int64), RunAsUser: new(int64)}
				*mainSecurityContext.RunAsGroup = 3000
				*mainSecurityContext.RunAsUser = 4000
				sidecarSecurityContext := &corev1.SecurityContext{RunAsGroup: new(int64), RunAsUser: new(int64)}
				*sidecarSecurityContext.RunAsGroup = 1000
				*sidecarSecurityContext.RunAsUser = 2000

				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {PodTemplate: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						SecurityContext: podSecurityContext,
						Containers: []corev1.Container{
							corev1.Container{
								Name:            "foundationdb",
								SecurityContext: mainSecurityContext,
							},
							corev1.Container{
								Name:            "foundationdb-kubernetes-sidecar",
								SecurityContext: sidecarSecurityContext,
							},
						},
						InitContainers: []corev1.Container{
							corev1.Container{
								Name:            "foundationdb-kubernetes-init",
								SecurityContext: sidecarSecurityContext,
							},
						},
					},
				}}}

				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the security contexts", func() {

				Expect(*spec.SecurityContext.FSGroup).To(Equal(int64(5000)))

				Expect(len(spec.InitContainers)).To(Equal(1))
				initContainer := spec.InitContainers[0]
				Expect(*initContainer.SecurityContext.RunAsGroup).To(Equal(int64(1000)))
				Expect(*initContainer.SecurityContext.RunAsUser).To(Equal(int64(2000)))

				mainContainer := spec.Containers[0]
				Expect(*mainContainer.SecurityContext.RunAsGroup).To(Equal(int64(3000)))
				Expect(*mainContainer.SecurityContext.RunAsUser).To(Equal(int64(4000)))

				sidecarContainer := spec.Containers[1]
				Expect(*sidecarContainer.SecurityContext.RunAsGroup).To(Equal(int64(1000)))
				Expect(*sidecarContainer.SecurityContext.RunAsUser).To(Equal(int64(2000)))

			})
		})

		Context("with a custom security context from the Spec.PodTemplate field", func() {
			BeforeEach(func() {

				podSecurityContext := &corev1.PodSecurityContext{FSGroup: new(int64)}
				*podSecurityContext.FSGroup = 5000
				mainSecurityContext := &corev1.SecurityContext{RunAsGroup: new(int64), RunAsUser: new(int64)}
				*mainSecurityContext.RunAsGroup = 3000
				*mainSecurityContext.RunAsUser = 4000
				sidecarSecurityContext := &corev1.SecurityContext{RunAsGroup: new(int64), RunAsUser: new(int64)}
				*sidecarSecurityContext.RunAsGroup = 1000
				*sidecarSecurityContext.RunAsUser = 2000

				cluster.Spec.PodTemplate = &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						SecurityContext: podSecurityContext,
						Containers: []corev1.Container{
							corev1.Container{
								Name:            "foundationdb",
								SecurityContext: mainSecurityContext,
							},
							corev1.Container{
								Name:            "foundationdb-kubernetes-sidecar",
								SecurityContext: sidecarSecurityContext,
							},
						},
						InitContainers: []corev1.Container{
							corev1.Container{
								Name:            "foundationdb-kubernetes-init",
								SecurityContext: sidecarSecurityContext,
							},
						},
					},
				}

				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the security contexts", func() {

				Expect(*spec.SecurityContext.FSGroup).To(Equal(int64(5000)))

				Expect(len(spec.InitContainers)).To(Equal(1))
				initContainer := spec.InitContainers[0]
				Expect(*initContainer.SecurityContext.RunAsGroup).To(Equal(int64(1000)))
				Expect(*initContainer.SecurityContext.RunAsUser).To(Equal(int64(2000)))

				mainContainer := spec.Containers[0]
				Expect(*mainContainer.SecurityContext.RunAsGroup).To(Equal(int64(3000)))
				Expect(*mainContainer.SecurityContext.RunAsUser).To(Equal(int64(4000)))

				sidecarContainer := spec.Containers[1]
				Expect(*sidecarContainer.SecurityContext.RunAsGroup).To(Equal(int64(1000)))
				Expect(*sidecarContainer.SecurityContext.RunAsUser).To(Equal(int64(2000)))

			})
		})

		Context("with a custom security context from the Spec.PodSecurityContext field", func() {
			BeforeEach(func() {
				cluster.Spec.PodSecurityContext = &corev1.PodSecurityContext{FSGroup: new(int64)}
				*cluster.Spec.PodSecurityContext.FSGroup = 5000

				cluster.Spec.MainContainer.SecurityContext = &corev1.SecurityContext{RunAsGroup: new(int64), RunAsUser: new(int64)}
				*cluster.Spec.MainContainer.SecurityContext.RunAsGroup = 3000
				*cluster.Spec.MainContainer.SecurityContext.RunAsUser = 4000
				cluster.Spec.SidecarContainer.SecurityContext = &corev1.SecurityContext{RunAsGroup: new(int64), RunAsUser: new(int64)}
				*cluster.Spec.SidecarContainer.SecurityContext.RunAsGroup = 1000
				*cluster.Spec.SidecarContainer.SecurityContext.RunAsUser = 2000

				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the security contexts", func() {
				Expect(*spec.SecurityContext.FSGroup).To(Equal(int64(5000)))

				Expect(len(spec.InitContainers)).To(Equal(1))
				initContainer := spec.InitContainers[0]
				Expect(*initContainer.SecurityContext.RunAsGroup).To(Equal(int64(1000)))
				Expect(*initContainer.SecurityContext.RunAsUser).To(Equal(int64(2000)))

				mainContainer := spec.Containers[0]
				Expect(*mainContainer.SecurityContext.RunAsGroup).To(Equal(int64(3000)))
				Expect(*mainContainer.SecurityContext.RunAsUser).To(Equal(int64(4000)))

				sidecarContainer := spec.Containers[1]
				Expect(*sidecarContainer.SecurityContext.RunAsGroup).To(Equal(int64(1000)))
				Expect(*sidecarContainer.SecurityContext.RunAsUser).To(Equal(int64(2000)))
			})
		})

		Context("with an instance ID prefix", func() {
			BeforeEach(func() {
				cluster.Spec.InstanceIDPrefix = "dc1"
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should put the prefix in the instance ID", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "dc1-storage-1"},
				}))
			})
		})

		Context("with serevice account token disabled", func() {
			BeforeEach(func() {
				var automount = false
				cluster.Spec.AutomountServiceAccountToken = &automount
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("disables the token in the spec", func() {
				Expect(*spec.AutomountServiceAccountToken).To(BeFalse())
			})
		})

		Context("with command line arguments for the sidecar", func() {
			BeforeEach(func() {
				cluster.Spec.Version = Versions.WithCommandLineVariablesForSidecar.String()
				cluster.Spec.SidecarVariables = []string{"FAULT_DOMAIN", "ZONE"}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should add the arguments to the init container", func() {
				Expect(len(spec.InitContainers)).To(Equal(1))
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
				Expect(initContainer.Args).To(Equal([]string{
					"--copy-file", "fdb.cluster", "--input-monitor-conf", "fdbmonitor.conf",
					"--copy-binary", "fdbserver", "--copy-binary", "fdbcli",
					"--main-container-version", cluster.Spec.Version,
					"--substitute-variable", "FAULT_DOMAIN", "--substitute-variable", "ZONE",
					"--init-mode",
				}))
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))
			})

			It("should add the arguments to the sidecar", func() {
				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))

				Expect(sidecarContainer.Args).To(Equal([]string{
					"--copy-file", "fdb.cluster", "--input-monitor-conf", "fdbmonitor.conf",
					"--copy-binary", "fdbserver", "--copy-binary", "fdbcli",
					"--main-container-version", cluster.Spec.Version,
					"--substitute-variable", "FAULT_DOMAIN", "--substitute-variable", "ZONE",
				}))

				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: ""},
				}))
			})

			It("should not include the sidecar conf in the config map", func() {
				Expect(spec.Volumes[2]).To(Equal(corev1.Volume{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{Key: "fdbmonitor-conf-storage", Path: "fdbmonitor.conf"},
							corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
						},
					}},
				}))
			})
		})

		Context("with environment variables for the sidecar", func() {
			BeforeEach(func() {
				cluster.Spec.Version = Versions.WithEnvironmentVariablesForSidecar.String()
				cluster.Spec.SidecarVariables = []string{"FAULT_DOMAIN", "ZONE"}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds the environment variables to the init container", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Name).To(Equal("foundationdb-kubernetes-init"))
				Expect(initContainer.Args).To(BeNil())
				Expect(initContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "COPY_ONCE", Value: "1"},
					corev1.EnvVar{Name: "SIDECAR_CONF_DIR", Value: "/var/input-files"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
				}))
			})

			It("adds the environment variables to the sidecar container", func() {
				sidecarContainer := spec.Containers[1]
				Expect(sidecarContainer.Name).To(Equal("foundationdb-kubernetes-sidecar"))
				Expect(sidecarContainer.Args).To(BeNil())
				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "SIDECAR_CONF_DIR", Value: "/var/input-files"},
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: ""},
				}))

				Expect(len(spec.Volumes)).To(Equal(4))

				Expect(spec.Volumes[2]).To(Equal(corev1.Volume{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{Key: "fdbmonitor-conf-storage", Path: "fdbmonitor.conf"},
							corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
							corev1.KeyToPath{Key: "sidecar-conf", Path: "config.json"},
						},
					}},
				}))
			})
		})

		Context("with custom map", func() {
			BeforeEach(func() {
				cluster.Spec.ConfigMap = &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "config1"}}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds config-map volume that refers to custom map", func() {
				Expect(spec.Volumes[2].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal(fmt.Sprintf("%s-%s", cluster.Name, "config1")))
			})
		})

		Context("with no custom map", func() {
			BeforeEach(func() {
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})
			It("adds config-map volume that refers to custom map", func() {
				Expect(spec.Volumes[2].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal(fmt.Sprintf("%s-%s", cluster.Name, "config")))
			})
		})

		Context("with custom pvc", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {VolumeClaim: &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "claim1"}}}}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds data volume that refers to custom pvc", func() {
				Expect(spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName).To(Equal(fmt.Sprintf("%s-storage-1-%s", cluster.Name, "claim1")))
			})
		})

		Context("with custom pvc from the Spec.VolumeClaim field", func() {
			BeforeEach(func() {
				cluster.Spec.VolumeClaim = &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "claim1"}}
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds data volume that refers to custom pvc", func() {
				Expect(spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName).To(Equal(fmt.Sprintf("%s-storage-1-%s", cluster.Name, "claim1")))
			})
		})

		Context("with no custom pvc", func() {
			BeforeEach(func() {
				spec, err = GetPodSpec(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})
			It("adds data volume that refers to default pvc", func() {
				Expect(spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName).To(Equal(fmt.Sprintf("%s-storage-1-%s", cluster.Name, "data")))
			})
		})

		Context("with a custom CA", func() {
			BeforeEach(func() {
				cluster.Spec.TrustedCAs = []string{"Test"}
				spec, err = GetPodSpec(cluster, "storage", 1)
			})

			It("should pass the CA file to the main container", func() {
				mainContainer := spec.Containers[0]
				Expect(mainContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
					corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/dynamic-conf/ca.pem"},
				}))
			})

			It("should pass the CA file to the init container", func() {
				initContainer := spec.InitContainers[0]
				Expect(initContainer.Args).To(Equal([]string{
					"--copy-file",
					"fdb.cluster",
					"--copy-file",
					"ca.pem",
					"--input-monitor-conf",
					"fdbmonitor.conf",
					"--copy-binary",
					"fdbserver",
					"--copy-binary",
					"fdbcli",
					"--main-container-version",
					"6.2.20",
					"--init-mode",
				}))
			})

			It("should pass the CA to the sidecar container", func() {
				sidecarContainer := spec.Containers[1]

				Expect(sidecarContainer.Args).To(Equal([]string{
					"--copy-file",
					"fdb.cluster",
					"--copy-file",
					"ca.pem",
					"--input-monitor-conf",
					"fdbmonitor.conf",
					"--copy-binary",
					"fdbserver",
					"--copy-binary",
					"fdbcli",
					"--main-container-version",
					"6.2.20",
				}))

				Expect(sidecarContainer.Env).To(Equal([]corev1.EnvVar{
					corev1.EnvVar{Name: "FDB_PUBLIC_IP", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"},
					}},
					corev1.EnvVar{Name: "FDB_MACHINE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_ZONE_ID", ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
					}},
					corev1.EnvVar{Name: "FDB_INSTANCE_ID", Value: "storage-1"},
					corev1.EnvVar{Name: "FDB_TLS_VERIFY_PEERS", Value: ""},
					corev1.EnvVar{Name: "FDB_TLS_CA_FILE", Value: "/var/input-files/ca.pem"},
				}))
			})

			It("should have the CA file in the config map volume", func() {
				Expect(len(spec.Volumes)).To(Equal(4))
				Expect(spec.Volumes[2]).To(Equal(corev1.Volume{
					Name: "config-map",
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
						Items: []corev1.KeyToPath{
							corev1.KeyToPath{Key: "fdbmonitor-conf-storage", Path: "fdbmonitor.conf"},
							corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
							corev1.KeyToPath{Key: "ca-file", Path: "ca.pem"},
						},
					}},
				}))
			})
		})
	})

	Describe("GetPvc", func() {
		var pvc *corev1.PersistentVolumeClaim

		Context("with a basic storage instance", func() {
			BeforeEach(func() {
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the metadata on the PVC", func() {
				Expect(pvc.Namespace).To(Equal("my-ns"))
				Expect(pvc.Name).To(Equal(fmt.Sprintf("%s-storage-1-data", cluster.Name)))
				Expect(pvc.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "storage-1",
				}))
			})

			It("should set the spec on the PVC", func() {
				Expect(pvc.Spec).To(Equal(corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": resource.MustParse("128G"),
						},
					},
				}))
			})
		})

		Context("with a custom storage size", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {VolumeClaim: &corev1.PersistentVolumeClaim{
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse("64G"),
							},
						},
					},
				}}}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the storage size on the resources", func() {
				Expect(pvc.Spec.Resources).To(Equal(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse("64G"),
					},
				}))
			})
		})

		Context("with a custom storage size from the Spec.VolumeClaim field", func() {
			BeforeEach(func() {
				cluster.Spec.VolumeClaim = &corev1.PersistentVolumeClaim{
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse("64G"),
							},
						},
					},
				}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the storage size on the resources", func() {
				Expect(pvc.Spec.Resources).To(Equal(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse("64G"),
					},
				}))
			})
		})

		Context("with a custom storage size from the VolumeSize field", func() {
			BeforeEach(func() {
				cluster.Spec.VolumeSize = "64G"
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the storage size on the resources", func() {
				Expect(pvc.Spec.Resources).To(Equal(corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse("64G"),
					},
				}))
			})
		})

		Context("with custom metadata", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {VolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"fdb-annotation": "value1",
						},
						Labels: map[string]string{
							"fdb-label": "value2",
						},
					},
				}}}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the metadata on the PVC", func() {
				Expect(pvc.Namespace).To(Equal("my-ns"))
				Expect(pvc.Name).To(Equal(fmt.Sprintf("%s-storage-1-data", cluster.Name)))
				Expect(pvc.ObjectMeta.Annotations).To(Equal(map[string]string{
					"fdb-annotation":                     "value1",
					"foundationdb.org/last-applied-spec": "f0c8a45ea6c3dd26c2dc2b5f3c699f38d613dab273d0f8a6eae6abd9a9569063",
				}))
				Expect(pvc.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "storage-1",
					"fdb-label":         "value2",
				}))
			})
		})

		Context("with custom metadata from the Spec.VolumeClaim field", func() {
			BeforeEach(func() {
				cluster.Spec.VolumeClaim = &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"fdb-annotation": "value1",
						},
						Labels: map[string]string{
							"fdb-label": "value2",
						},
					},
				}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the metadata on the PVC", func() {
				Expect(pvc.Namespace).To(Equal("my-ns"))
				Expect(pvc.Name).To(Equal(fmt.Sprintf("%s-storage-1-data", cluster.Name)))
				Expect(pvc.ObjectMeta.Annotations).To(Equal(map[string]string{
					"fdb-annotation":                     "value1",
					"foundationdb.org/last-applied-spec": "f0c8a45ea6c3dd26c2dc2b5f3c699f38d613dab273d0f8a6eae6abd9a9569063",
				}))
				Expect(pvc.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "storage-1",
					"fdb-label":         "value2",
				}))
			})
		})

		Context("with a value size of 0", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {VolumeClaim: &corev1.PersistentVolumeClaim{
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse("0"),
							},
						},
					},
				}}}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a nil PVC", func() {
				Expect(pvc).To(BeNil())
			})
		})

		Context("with a value size of 0 with deprecated field", func() {
			BeforeEach(func() {
				cluster.Spec.VolumeSize = "0"
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a nil PVC", func() {
				Expect(pvc).To(BeNil())
			})
		})

		Context("with a custom storage class", func() {
			var class string
			BeforeEach(func() {
				class = "local"
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {VolumeClaim: &corev1.PersistentVolumeClaim{
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &class,
					},
				}}}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the storage class on the PVC", func() {
				Expect(pvc.Spec.StorageClassName).To(Equal(&class))
			})
		})

		Context("with a custom storage class from the Spec.VolumeClaim field", func() {
			var class string
			BeforeEach(func() {
				class = "local"
				cluster.Spec.VolumeClaim = &corev1.PersistentVolumeClaim{
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &class,
					},
				}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the storage class on the PVC", func() {
				Expect(pvc.Spec.StorageClassName).To(Equal(&class))
			})
		})

		Context("with a custom storage class from the Spec.StorageClass field", func() {
			var class string
			BeforeEach(func() {
				class = "local"
				cluster.Spec.StorageClass = &class
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set the storage class on the PVC", func() {
				Expect(pvc.Spec.StorageClassName).To(Equal(&class))
			})
		})

		Context("for a stateless instance", func() {
			BeforeEach(func() {
				pvc, err = GetPvc(cluster, "stateless", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a nil PVC", func() {
				Expect(pvc).To(BeNil())
			})
		})

		Context("with an instance ID prefix", func() {
			BeforeEach(func() {
				cluster.Spec.InstanceIDPrefix = "dc1"
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should include the prefix in the instance IDs", func() {
				Expect(pvc.Name).To(Equal(fmt.Sprintf("%s-storage-1-data", cluster.Name)))
				Expect(pvc.ObjectMeta.Labels).To(Equal(map[string]string{
					"fdb-cluster-name":  cluster.Name,
					"fdb-process-class": "storage",
					"fdb-instance-id":   "dc1-storage-1",
				}))
			})
		})

		Context("with custom name in the suffix", func() {
			BeforeEach(func() {
				cluster.Spec.Processes = map[string]fdbtypes.ProcessSettings{"general": {VolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc1"},
				}}}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should include claim name with custom suffix", func() {
				Expect(pvc.Name).To(Equal(fmt.Sprintf("%s-storage-1-pvc1", cluster.Name)))
			})
		})

		Context("with custom name in the suffix from the Spec.VolumeClaim field", func() {
			BeforeEach(func() {
				cluster.Spec.VolumeClaim = &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc1"},
				}
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should include claim name with custom suffix", func() {
				Expect(pvc.Name).To(Equal(fmt.Sprintf("%s-storage-1-pvc1", cluster.Name)))
			})
		})

		Context("with default name in the suffix", func() {
			BeforeEach(func() {
				pvc, err = GetPvc(cluster, "storage", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should include claim name with default suffix", func() {
				Expect(pvc.Name).To(Equal(fmt.Sprintf("%s-storage-1-data", cluster.Name)))
			})
		})
	})

	Describe("GetBackupDeployment", func() {
		var backup *fdbtypes.FoundationDBBackup
		var deployment *appsv1.Deployment

		BeforeEach(func() {
			backup = createDefaultBackup(cluster)
			Expect(backup.Name).To(Equal("operator-test-1"))
		})

		Context("with a basic deployment", func() {
			BeforeEach(func() {
				deployment, err = GetBackupDeployment(context.TODO(), backup, k8sClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(deployment).NotTo(BeNil())
			})

			It("should set the metadata for the deployment", func() {
				Expect(deployment.ObjectMeta.Name).To(Equal("operator-test-1-backup-agents"))
				Expect(len(deployment.ObjectMeta.OwnerReferences)).To(Equal(1))
				Expect(deployment.ObjectMeta.OwnerReferences[0].UID).To(Equal(cluster.ObjectMeta.UID))
				Expect(deployment.ObjectMeta.Labels).To(Equal(map[string]string{
					"foundationdb.org/backup-for": string(cluster.ObjectMeta.UID),
				}))
				Expect(deployment.ObjectMeta.Annotations).To(Equal(map[string]string{
					"foundationdb.org/last-applied-spec": "53bf93c896578af51723c0db12e884751be4ee702c7487a1a57108fa111a23d6",
				}))
			})

			It("should set the replication factor to the specified agent count", func() {
				Expect(deployment.Spec.Replicas).NotTo(BeNil())
				Expect(*deployment.Spec.Replicas).To(Equal(int32(3)))
			})

			It("should set the labels for the pod selector", func() {
				Expect(*deployment.Spec.Selector).To(Equal(metav1.LabelSelector{MatchLabels: map[string]string{
					"foundationdb.org/deployment-name": "operator-test-1-backup-agents",
				}}))
				Expect(deployment.Spec.Template.ObjectMeta.Labels).To(Equal(map[string]string{
					"foundationdb.org/deployment-name": "operator-test-1-backup-agents",
				}))
			})

			It("should have one container and one init container", func() {
				Expect(len(deployment.Spec.Template.Spec.Containers)).To(Equal(1))
				Expect(len(deployment.Spec.Template.Spec.InitContainers)).To(Equal(1))
			})

			It("should have the default volumes", func() {
				Expect(deployment.Spec.Template.Spec.Volumes).To(Equal([]corev1.Volume{
					{Name: "logs", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					{Name: "dynamic-conf", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					{
						Name: "config-map",
						VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
							Items: []corev1.KeyToPath{
								corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
							},
						}},
					},
				}))
			})

			Describe("the main container", func() {
				var container corev1.Container

				BeforeEach(func() {
					container = deployment.Spec.Template.Spec.Containers[0]
				})

				It("should set the container name", func() {
					Expect(container.Name).To(Equal("foundationdb"))
				})

				It("should set the image and command for the backup agent", func() {
					Expect(container.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb:%s", cluster.Spec.Version)))
					Expect(container.Command).To(Equal([]string{"backup_agent"}))
					Expect(container.Args).To(Equal([]string{
						"--log",
						"--logdir",
						"/var/log/fdb-trace-logs",
					}))
				})

				It("should set the basic environment", func() {
					Expect(container.Env).To(Equal([]corev1.EnvVar{
						{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
					}))
				})

				It("should set the default volume mounts", func() {
					Expect(container.VolumeMounts).To(Equal([]corev1.VolumeMount{
						{Name: "logs", MountPath: "/var/log/fdb-trace-logs"},
						{Name: "dynamic-conf", MountPath: "/var/dynamic-conf"},
					}))
				})

				It("should set default resource limits", func() {
					Expect(*container.Resources.Limits.Cpu()).To(Equal(resource.MustParse("1")))
					Expect(*container.Resources.Limits.Memory()).To(Equal(resource.MustParse("1Gi")))
					Expect(*container.Resources.Requests.Cpu()).To(Equal(resource.MustParse("1")))
					Expect(*container.Resources.Requests.Memory()).To(Equal(resource.MustParse("1Gi")))
				})
			})

			Describe("the init container", func() {
				var container corev1.Container

				BeforeEach(func() {
					container = deployment.Spec.Template.Spec.InitContainers[0]
				})

				It("should set the container name", func() {
					Expect(container.Name).To(Equal("foundationdb-kubernetes-init"))
				})

				It("should set the image and command for the container", func() {
					Expect(container.Image).To(Equal(fmt.Sprintf("foundationdb/foundationdb-kubernetes-sidecar:%s-1", cluster.Spec.Version)))
					Expect(container.Args).To(Equal([]string{
						"--copy-file",
						"fdb.cluster",
						"--require-not-empty",
						"fdb.cluster",
						"--init-mode",
					}))
				})

				It("should set the default volume mounts", func() {
					Expect(container.VolumeMounts).To(Equal([]corev1.VolumeMount{
						{Name: "config-map", MountPath: "/var/input-files"},
						{Name: "dynamic-conf", MountPath: "/var/output-files"},
					}))
				})
			})
		})

		Context("with a custom secret for the backup credentials", func() {
			BeforeEach(func() {
				backup.Spec.PodTemplateSpec = &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{Name: "secrets", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "backup-secrets"}}},
						},
						Containers: []corev1.Container{
							{
								Name: "foundationdb",
								Env: []corev1.EnvVar{
									{Name: "FDB_BLOB_CREDENTIALS", Value: "/var/secrets/blob_credentials.json"},
								},
								VolumeMounts: []corev1.VolumeMount{
									{Name: "secrets", MountPath: "/var/secrets"},
								},
							},
						},
					},
				}
				deployment, err = GetBackupDeployment(context.TODO(), backup, k8sClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(deployment).NotTo(BeNil())
			})

			It("should customize the volumes", func() {
				Expect(deployment.Spec.Template.Spec.Volumes).To(Equal([]corev1.Volume{
					{Name: "secrets", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "backup-secrets"}}},
					{Name: "logs", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					{Name: "dynamic-conf", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					{
						Name: "config-map",
						VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", cluster.Name)},
							Items: []corev1.KeyToPath{
								corev1.KeyToPath{Key: "cluster-file", Path: "fdb.cluster"},
							},
						}},
					},
				}))
			})

			Describe("the main container", func() {
				var container corev1.Container

				BeforeEach(func() {
					container = deployment.Spec.Template.Spec.Containers[0]
				})

				It("should set the container name", func() {
					Expect(container.Name).To(Equal("foundationdb"))
				})

				It("should customize the environment", func() {
					Expect(container.Env).To(Equal([]corev1.EnvVar{
						{Name: "FDB_BLOB_CREDENTIALS", Value: "/var/secrets/blob_credentials.json"},
						{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
					}))
				})

				It("should customize the volume mounts", func() {
					Expect(container.VolumeMounts).To(Equal([]corev1.VolumeMount{
						{Name: "secrets", MountPath: "/var/secrets"},
						{Name: "logs", MountPath: "/var/log/fdb-trace-logs"},
						{Name: "dynamic-conf", MountPath: "/var/dynamic-conf"},
					}))
				})
			})
		})

		Context("with a nil agent count", func() {
			BeforeEach(func() {
				backup.Spec.AgentCount = nil
				deployment, err = GetBackupDeployment(context.TODO(), backup, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have 2 replicas", func() {
				Expect(deployment).NotTo(BeNil())
				Expect(*deployment.Spec.Replicas).To(Equal(int32(2)))
			})
		})

		Context("with an agent count of 0", func() {
			BeforeEach(func() {
				agentCount := 0
				backup.Spec.AgentCount = &agentCount
				deployment, err = GetBackupDeployment(context.TODO(), backup, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should be nil", func() {
				Expect(deployment).To(BeNil())
			})
		})

		Context("with a custom TLS CA file", func() {
			BeforeEach(func() {
				backup.Spec.PodTemplateSpec = &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: "foundationdb",
							Env: []corev1.EnvVar{{
								Name:  "FDB_TLS_CA_FILE",
								Value: "/tmp/ca.pem",
							}},
						}},
					},
				}
				deployment, err = GetBackupDeployment(context.TODO(), backup, k8sClient)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should respect the custom CA", func() {
				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Env).To(Equal([]corev1.EnvVar{
					{Name: "FDB_TLS_CA_FILE", Value: "/tmp/ca.pem"},
					{Name: "FDB_CLUSTER_FILE", Value: "/var/dynamic-conf/fdb.cluster"},
				}))
			})
		})

		Context("with the sidecar require-not-empty field", func() {
			BeforeEach(func() {
				backup.Spec.Version = Versions.WithSidecarCrashOnEmpty.String()
				deployment, err = GetBackupDeployment(context.TODO(), backup, k8sClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(deployment).NotTo(BeNil())
			})

			Describe("the init container", func() {
				var container corev1.Container

				BeforeEach(func() {
					container = deployment.Spec.Template.Spec.InitContainers[0]
				})

				It("should have a flag to require the cluster file is present", func() {
					Expect(container.Args).To(Equal([]string{
						"--copy-file",
						"fdb.cluster",
						"--require-not-empty",
						"fdb.cluster",
						"--init-mode",
					}))
				})
			})
		})

		Context("without the sidecar require-not-empty field", func() {
			BeforeEach(func() {
				backup.Spec.Version = Versions.WithoutSidecarCrashOnEmpty.String()
				deployment, err = GetBackupDeployment(context.TODO(), backup, k8sClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(deployment).NotTo(BeNil())
			})

			Describe("the init container", func() {
				var container corev1.Container

				BeforeEach(func() {
					container = deployment.Spec.Template.Spec.InitContainers[0]
				})

				It("should not have a flag to require the cluster file is present", func() {
					Expect(container.Args).To(Equal([]string{
						"--copy-file",
						"fdb.cluster",
						"--init-mode",
					}))
				})
			})
		})
	})
})
