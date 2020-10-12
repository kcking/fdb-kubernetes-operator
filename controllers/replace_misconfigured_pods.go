/*
 * replace_misconfigured_pods.go
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
	ctx "context"
	"time"

	fdbtypes "github.com/FoundationDB/fdb-kubernetes-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// ReplaceMisconfiguredPods identifies processes that need to be replaced in
// order to bring up new processes with different configuration.
type ReplaceMisconfiguredPods struct{}

// Reconcile runs the reconciler's work.
func (c ReplaceMisconfiguredPods) Reconcile(r *FoundationDBClusterReconciler, context ctx.Context, cluster *fdbtypes.FoundationDBCluster) (bool, error) {
	hasNewRemovals := false

	var removals = cluster.Status.PendingRemovals

	if removals == nil {
		removals = make(map[string]fdbtypes.PendingRemovalState)
	}

	pvcs := &corev1.PersistentVolumeClaimList{}
	err := r.List(context, pvcs, getPodListOptions(cluster, "", "")...)
	if err != nil {
		return false, err
	}

	for _, pvc := range pvcs.Items {
		ownedByCluster := false
		for _, ownerReference := range pvc.OwnerReferences {
			if ownerReference.UID == cluster.UID {
				ownedByCluster = true
				break
			}
		}

		if !ownedByCluster {
			continue
		}

		instanceID := GetInstanceIDFromMeta(pvc.ObjectMeta)
		_, pendingRemoval := removals[instanceID]
		if pendingRemoval {
			continue
		}
		_, idNum, err := ParseInstanceID(instanceID)
		if err != nil {
			return false, err
		}

		processClass := GetProcessClassFromMeta(pvc.ObjectMeta)
		desiredPVC, err := GetPvc(cluster, processClass, idNum)
		pvcHash, err := GetJSONHash(desiredPVC.Spec)
		if pvc.Annotations[LastSpecKey] != pvcHash && false { // disable for now
			instances, err := r.PodLifecycleManager.GetInstances(r, cluster, context, getSinglePodListOptions(cluster, instanceID)...)
			if err != nil {
				return false, err
			}
			if len(instances) > 0 {
				removalState := r.getPendingRemovalState(instances[0])
				removals[instanceID] = removalState
				hasNewRemovals = true
			}
		}
		if err != nil {
			return false, err
		}
	}

	instances, err := r.PodLifecycleManager.GetInstances(r, cluster, context, getPodListOptions(cluster, "", "")...)
	if err != nil {
		return false, err
	}

	for _, instance := range instances {
		if instance.Pod == nil {
			continue
		}

		instanceID := instance.GetInstanceID()
		_, pendingRemoval := removals[instanceID]
		if pendingRemoval {
			continue
		}

		_, idNum, err := ParseInstanceID(instanceID)
		if err != nil {
			return false, err
		}

		needsRemoval := false

		_, desiredInstanceID := getInstanceID(cluster, instance.GetProcessClass(), idNum)
		if err != nil {
			return false, err
		}

		if instanceID != desiredInstanceID {
			needsRemoval = true
		}

		if cluster.Spec.UpdatePodsByReplacement {
			specHash, err := GetPodSpecHash(cluster, instance.GetProcessClass(), idNum, nil)
			if err != nil {
				return false, err
			}

			if instance.Metadata.Annotations[LastSpecKey] != specHash {
				needsRemoval = true
			}
		}

		if needsRemoval {
			removalState := r.getPendingRemovalState(instance)
			removals[instanceID] = removalState
			hasNewRemovals = true
		}
	}

	if hasNewRemovals {
		cluster.Status.PendingRemovals = removals
		err = r.Status().Update(context, cluster)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return true, nil
}

// RequeueAfter returns the delay before we should run the reconciliation
// again.
func (c ReplaceMisconfiguredPods) RequeueAfter() time.Duration {
	return 0
}
