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
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		log.Error(err, "unable to fetch Function")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: Call storage center to store the function code

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

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha1.Function{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
