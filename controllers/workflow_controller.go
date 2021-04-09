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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	"github.com/tass-io/tass-operator/pkg/workflow"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// WorkflowReconciler reconciles a Workflow object
type WorkflowReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=serverless.tass.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.tass.io,resources=workflows/status,verbs=get;update;patch

func (r *WorkflowReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("workflow", req.NamespacedName)

	var instance serverlessv1alpha1.Workflow
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		log.Error(err, "unable to fetch Workflow")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.V(1).Info("The Workflow environment is", "environment", instance.Spec.Environment)
	log.V(1).Info("The Workflow Spec is", "spec", instance.Spec.Spec)

	// For example, we want to get the Function list...
	var functionList serverlessv1alpha1.FunctionList
	if err := r.List(ctx, &functionList, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list child Functions")
		return ctrl.Result{}, err
	}

	// TODO: This kind of check should be placed in the admission webhook
	// Put here temporarily
	if err := workflow.ValidateFuncExist(&instance, &functionList); err != nil {
		log.Error(err, "Workflow validation error")
		return ctrl.Result{}, nil
	}
	if err := workflow.ValidateFlows(&instance); err != nil {
		log.Error(err, "Workflow validation error")
		return ctrl.Result{}, nil
	}

	// TODO: This is just an example status
	instance.Status.Status = "Running"
	if err := r.Status().Update(ctx, &instance); err != nil {
		log.Error(err, "unable to update status")
	}

	// FIXME: A sample of create a Function by controller-runtime
	// Only show the case of creating a Function, should be deleted manually
	var sample serverlessv1alpha1.Function
	if err := r.Get(ctx, types.NamespacedName{Namespace: "default", Name: "create-function-by-controller-runtime"}, &sample); errors.IsNotFound(err) {
		// do something
		sample = serverlessv1alpha1.Function{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "create-function-by-controller-runtime",
			},
			Spec: serverlessv1alpha1.FunctionSpec{
				Environment: "JavaScript",
			},
		}
		if err := r.Create(context.Background(), &sample); err != nil {
			log.Error(err, "Cannot create Function")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha1.Workflow{}).
		Complete(r)
}
