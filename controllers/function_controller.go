/*


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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	"github.com/tass-io/tass-operator/pkg/spawn"
	corev1 "k8s.io/api/core/v1"
)

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=serverless.tass.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.tass.io,resources=functions/status,verbs=get;update;patch

func (r *FunctionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("function", req.NamespacedName)

	var instance serverlessv1alpha1.Function
	// TODO: Call storage center to store the function code
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		log.Error(err, "unable to fetch Function")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: This is a sample of finalizer
	fnFinalizerName := "function.finalizers.tass.io"

	// Examine DeletionTimestamp to determine if object is under deletion
	// If an object is under deletion, gc controller will create a field called `meta.DeletionTimestamp`
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.ObjectMeta.Finalizers, fnFinalizerName) {
			log.V(1).Info("Add Function finalizer...")
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, fnFinalizerName)
			if err := r.Update(ctx, &instance); err != nil {
				return ctrl.Result{}, err
			}
			log.V(1).Info("Add Function finalizer successfully")
		}
	} else {
		log.V(1).Info("Remove Function finalizer...")
		// FIXME: Mock long time to remove a resource, which will be easier to debug
		// Remove it in the future.
		time.Sleep(time.Second * 5)
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, fnFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(&instance); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, fnFinalizerName)
			if err := r.Update(ctx, &instance); err != nil {
				return ctrl.Result{}, err
			}
			log.V(1).Info("Remove Function finalizer successfully")
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// FIXME: This is a sample of creating a Pod
	pod := spawn.NewPodForCR(instance)
	if err := ctrl.SetControllerReference(&instance, pod, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	podInstance := &corev1.Pod{}
	// try to see if the pod already exists
	if err := r.Get(ctx, req.NamespacedName, podInstance); errors.IsNotFound(err) {
		log.V(1).Info("Creating Pod...")

		// does not exist, create a pod
		if err = r.Create(ctx, pod); err != nil {
			return ctrl.Result{}, err
		}
		// Successfully created a Pod
		log.V(1).Info("Pod Created successfully", "name", pod.Name)

		return ctrl.Result{}, nil
	} else if err != nil {
		// requeue with err
		log.Error(err, "Cannot create pod")
		return ctrl.Result{}, err
	} else {
		log.V(1).Info("Pod exists, no need to create a pod.")
	}
	return ctrl.Result{}, nil
}

func (r *FunctionReconciler) deleteExternalResources(fn *serverlessv1alpha1.Function) error {
	// Actually, due to the help of garbage collector in k8s, we don't need to delete the cascade resource manually
	// Here I just show a case of deleteing a Pod
	//
	// delete external Pod resources associated with the Function
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple types for same object.
	pList := &corev1.PodList{}
	err := r.List(context.Background(), pList,
		client.InNamespace(fn.Namespace), client.MatchingLabels(map[string]string{"app": fn.Name}))
	if err != nil {
		return fmt.Errorf("Cannot find Pod match Function, %s", err.Error())
	}
	for _, pod := range pList.Items {
		err = r.Delete(context.Background(), &pod)
		if err != nil {
			return fmt.Errorf("Cannot delete Pod %s/%s, %s", pod.Namespace, pod.Name, err.Error())
		}
	}
	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha1.Function{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
