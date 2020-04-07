// +build !ignore_autogenerated

/*
Copyright 2019 Pivotal.

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

// Code generated by controller-gen. DO NOT EDIT.

package status

import (
	"k8s.io/api/core/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAvailableConditionManager) DeepCopyInto(out *ClusterAvailableConditionManager) {
	*out = *in
	in.condition.DeepCopyInto(&out.condition)
	if in.endpoints != nil {
		in, out := &in.endpoints, &out.endpoints
		*out = new(v1.Endpoints)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAvailableConditionManager.
func (in *ClusterAvailableConditionManager) DeepCopy() *ClusterAvailableConditionManager {
	if in == nil {
		return nil
	}
	out := new(ClusterAvailableConditionManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqClusterCondition) DeepCopyInto(out *RabbitmqClusterCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqClusterCondition.
func (in *RabbitmqClusterCondition) DeepCopy() *RabbitmqClusterCondition {
	if in == nil {
		return nil
	}
	out := new(RabbitmqClusterCondition)
	in.DeepCopyInto(out)
	return out
}