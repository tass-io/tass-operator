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
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
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
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		log.Error(err, "unable to fetch Function")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, nil
}

// NOTE: This is a sample of finalizer, put here as a ref in case of need
// DoFinal is a sampel of usage of Finalizer
func (r *FunctionReconciler) DoFinal(instance serverlessv1alpha1.Function) error {
	fnFinalizerName := "function.finalizers.tass.io"
	ctx := context.Background()

	// Examine DeletionTimestamp to determine if object is under deletion
	// If an object is under deletion, gc controller will create a field called `meta.DeletionTimestamp`
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.ObjectMeta.Finalizers, fnFinalizerName) {
			r.Log.V(1).Info("Add Function finalizer...")
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, fnFinalizerName)
			if err := r.Update(ctx, &instance); err != nil {
				return err
			}
			r.Log.V(1).Info("Add Function finalizer successfully")
		}
	} else {
		r.Log.V(1).Info("Remove Function finalizer...")
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, fnFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			r.Log.V(1).Info("Do finalizer...")
			// Mock take some time to remove a Resource
			time.Sleep(time.Second * 5)
			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, fnFinalizerName)
			if err := r.Update(ctx, &instance); err != nil {
				return err
			}
			r.Log.V(1).Info("Remove Function finalizer successfully")
		}
	}
	// Stop finalizer as the item is being deleted
	return nil
}

// Helper functions to check and remove string from a slice of strings.
//
// containString checks the finalizer field
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes the finalizer field
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
