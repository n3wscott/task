/*
Copyright 2019 The Knative Authors

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

package resources

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeService(args Arguments) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: svcName(args.Owner),
			Namespace:    args.Namespace,
			Labels:       Labels(args.Owner),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Owner),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       80,
				Protocol:   "TCP",
				TargetPort: intstr.FromInt(8080),
			}},
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector: map[string]string{
				labelKey: args.Owner.GetObjectMeta().GetName(),
			},
		},
	}
	return service
}

func svcName(owner kmeta.OwnerRefable) string {
	return strings.ToLower(
		strings.Join(append([]string{owner.GetObjectMeta().GetName(), "task"}), "-") + "-")
}
