/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1beta1"
)

var condSet = apis.NewBatchConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (t *Task) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Task")
}

func (ts *TaskStatus) InitializeConditions() {
	condSet.Manage(ts).InitializeConditions()
}

func (ts *TaskStatus) MarkAddress(url *apis.URL) {
	if ts.Address == nil {
		ts.Address = &v1beta1.Addressable{}
	}
	if url != nil {
		ts.Address.URL = url
		condSet.Manage(ts).MarkTrue(TaskConditionAddressable)
	} else {
		ts.Address.URL = nil
		condSet.Manage(ts).MarkFalse(TaskConditionAddressable, "ServiceUnavailable", "Service was not created.")
	}
}

func (ts *TaskStatus) MarkJobSucceeded() {
	condSet.Manage(ts).MarkTrue(TaskConditionResult)
}

func (ts *TaskStatus) MarkJobRunning(messageFormat string, messageA ...interface{}) {
	condSet.Manage(ts).MarkUnknown(TaskConditionResult, "Active", messageFormat, messageA...)
}

func (ts *TaskStatus) MarkJobFailed(reason, messageFormat string, messageA ...interface{}) {
	condSet.Manage(ts).MarkFalse(TaskConditionResult, reason, messageFormat, messageA...)
}
