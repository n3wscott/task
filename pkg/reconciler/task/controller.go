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

package task

import (
	"context"

	"github.com/n3wscott/task/pkg/apis/n3wscott/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/n3wscott/task/pkg/client/injection/client"
	taskinformer "github.com/n3wscott/task/pkg/client/injection/informers/n3wscott/v1alpha1/task"
	"knative.dev/pkg/injection/clients/kubeclient"
	jobinformer "knative.dev/pkg/injection/informers/kubeinformers/batchv1/job"
	svcinformer "knative.dev/pkg/injection/informers/kubeinformers/corev1/service"
)

const (
	controllerAgentName = "task-controller"
)

// NewController returns a new HPA reconcile controller.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	taskInformer := taskinformer.Get(ctx)
	svcInformer := svcinformer.Get(ctx)
	jobInformer := jobinformer.Get(ctx)

	c := &Reconciler{
		KubeClientSet: kubeclient.Get(ctx),
		Client:        client.Get(ctx),
		Lister:        taskInformer.Lister(),
		ServiceLister: svcInformer.Lister(),
		Recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
	}
	impl := controller.NewImpl(c, logger, "Tasks")

	logger.Info("Setting up event handlers")

	taskInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	svcInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Task")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	jobInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Task")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	return impl
}
