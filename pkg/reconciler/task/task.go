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
	"fmt"
	"reflect"

	"github.com/n3wscott/task/pkg/apis/n3wscott/v1alpha1"
	clientset "github.com/n3wscott/task/pkg/client/clientset/versioned"
	listers "github.com/n3wscott/task/pkg/client/listers/n3wscott/v1alpha1"
	"github.com/n3wscott/task/pkg/reconciler/task/resources"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

// Reconciler implements controller.Reconciler for Task resources.
type Reconciler struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	KubeClientSet kubernetes.Interface

	// Client is used to write back status updates.
	Client clientset.Interface

	// Listers index properties about resources
	Lister        listers.TaskLister
	ServiceLister corev1listers.ServiceLister

	// Recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	Recorder record.EventRecorder
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile implements controller.Reconciler
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	// If our controller has configuration state, we'd "freeze" it and
	// attach the frozen configuration to the context.
	//    ctx = r.configStore.ToContext(ctx)

	// Get the resource with this namespace/name.
	original, err := r.Lister.Tasks(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logger.Errorf("resource %q no longer exists", key)
		return nil
	} else if err != nil {
		return err
	}
	// Don't modify the informers copy.
	resource := original.DeepCopy()

	// Reconcile this copy of the resource and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcile(ctx, resource)
	if equality.Semantic.DeepEqual(original.Status, resource.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if _, err = r.updateStatus(resource); err != nil {
		logger.Warnw("Failed to update resource status", zap.Error(err))
		r.Recorder.Eventf(resource, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for %q: %v", resource.Name, err)
		return err
	}
	if reconcileErr != nil {
		r.Recorder.Event(resource, corev1.EventTypeWarning, "InternalError", reconcileErr.Error())
	}
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, task *v1alpha1.Task) error {
	if task.GetDeletionTimestamp() != nil {
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	task.Status.InitializeConditions()

	if err := r.reconcileJob(ctx, task); err != nil {
		return err
	}

	if err := r.reconcileService(ctx, task); err != nil {
		return err
	}

	task.Status.ObservedGeneration = task.Generation
	return nil
}

func (r *Reconciler) reconcileJob(ctx context.Context, task *v1alpha1.Task) error {
	job, err := r.getJob(ctx, task, labels.SelectorFromSet(resources.Labels(task)))

	// TODO: This should be an option. Comment out for now.
	//if task.Status.IsDone() {
	//	task.Status.ClearAddress()
	//	if job != nil {
	//		_ = r.KubeClientSet.BatchV1().Jobs(task.Namespace).Delete(job.Name, &metav1.DeleteOptions{})
	//	}
	//	return nil
	//}

	// If the resource doesn't exist, we'll create it
	if apierrs.IsNotFound(err) {
		job = resources.MakeJob(resources.Arguments{
			Owner:     task,
			Namespace: task.Namespace,
			Template:  task.Spec.Template,
		})

		job, err := r.KubeClientSet.BatchV1().Jobs(task.Namespace).Create(job)
		if err != nil || job == nil {
			msg := "Failed to make Job."
			if err != nil {
				msg = msg + " " + err.Error()
			}
			task.Status.MarkJobFailed("FailedCreate", msg)
			return fmt.Errorf("failed to create Job: %s", err)
		}
		task.Status.MarkJobRunning("Created Job %q.", job.Name)
		return nil
	} else if err != nil {
		task.Status.MarkJobFailed("FailedGet", err.Error())
		return fmt.Errorf("failed to get Job: %s", err)
	}

	if isJobComplete(job) {
		if isJobSucceeded(job) {
			task.Status.MarkJobSucceeded()
		} else if isJobFailed(job) {
			task.Status.MarkJobFailed(jobFailedReasonMessage(job))
		}
	}
	return nil
}

func (r *Reconciler) getJob(ctx context.Context, owner metav1.Object, ls labels.Selector) (*batchv1.Job, error) {
	list, err := r.KubeClientSet.BatchV1().Jobs(owner.GetNamespace()).List(metav1.ListOptions{
		LabelSelector: ls.String(),
	})
	if err != nil {
		return nil, err
	}

	for _, i := range list.Items {
		if metav1.IsControlledBy(&i, owner) {
			return &i, nil
		}
	}

	return nil, apierrs.NewNotFound(schema.GroupResource{}, "")
}

func (r *Reconciler) reconcileService(ctx context.Context, task *v1alpha1.Task) error {
	logger := logging.FromContext(ctx)
	svc, err := r.getService(ctx, task, labels.SelectorFromSet(resources.Labels(task)))

	if task.Status.IsDone() {
		logger.Info("Task IsDone, Clearing address.")
		task.Status.ClearAddress()
		if svc != nil {
			_ = r.KubeClientSet.CoreV1().Services(task.Namespace).Delete(svc.Name, &metav1.DeleteOptions{})
		}
		return nil
	}

	if apierrs.IsNotFound(err) {
		svc = resources.MakeService(resources.Arguments{
			Owner:     task,
			Namespace: task.Namespace,
			Template:  task.Spec.Template,
		})

		var err error
		svc, err = r.KubeClientSet.CoreV1().Services(task.Namespace).Create(svc)
		if err != nil || svc == nil {
			msg := "Failed to make Service."
			if err != nil {
				msg = msg + " " + err.Error()
			}
			task.Status.MarkAddress(nil)
			return fmt.Errorf("failed to create Job: %s", err)
		}
	} else if err != nil {
		task.Status.MarkAddress(nil)
		return err
	}

	url := &apis.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace),
	}

	task.Status.MarkAddress(url)
	return nil
}

func (r *Reconciler) getService(ctx context.Context, owner metav1.Object, ls labels.Selector) (*corev1.Service, error) {
	list, err := r.KubeClientSet.CoreV1().Services(owner.GetNamespace()).List(metav1.ListOptions{
		LabelSelector: ls.String(),
	})
	if err != nil {
		return nil, err
	}

	for _, i := range list.Items {
		if metav1.IsControlledBy(&i, owner) {
			return &i, nil
		}
	}

	return nil, apierrs.NewNotFound(schema.GroupResource{}, "")
}

// Update the Status of the resource.  Caller is responsible for checking
// for semantic differences before calling.
func (r *Reconciler) updateStatus(desired *v1alpha1.Task) (*v1alpha1.Task, error) {
	actual, err := r.Lister.Tasks(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status, desired.Status) {
		return actual, nil
	}
	// Don't modify the informers copy
	existing := actual.DeepCopy()
	existing.Status = desired.Status
	return r.Client.N3wscottV1alpha1().Tasks(desired.Namespace).UpdateStatus(existing)
}

func isJobComplete(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isJobSucceeded(job *batchv1.Job) bool {
	return !isJobFailed(job)
}

func isJobFailed(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func jobFailedReasonMessage(job *batchv1.Job) (string, string) { // returns (reason, message)
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return c.Reason, c.Message
		}
	}
	return "", ""
}

func getFirstTerminationMessage(pod *corev1.Pod) string {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Terminated != nil && cs.State.Terminated.Message != "" {
			return cs.State.Terminated.Message
		}
	}
	return ""
}
