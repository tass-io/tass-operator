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
	"errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
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

	var original serverlessv1alpha1.Workflow
	if err := r.Get(ctx, req.NamespacedName, &original); err != nil {
		log.Error(err, "unable to fetch Workflow")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.V(1).Info("The Workflow environment is", "environment", original.Spec.Environment)
	log.V(1).Info("The Workflow Spec is", "spec", original.Spec.Spec)

	// For example, we want to get the Function list...
	var functionList serverlessv1alpha1.FunctionList
	if err := r.List(ctx, &functionList, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list child Functions")
		return ctrl.Result{}, err
	}

	// Make sure every Function in a Workflow has been defined in Function CRD
	domainFunctionMap := map[string]bool{}
	for _, pre := range functionList.Items {
		if pre.Spec.Domain == original.Spec.Domain {
			domainFunctionMap[pre.Name] = true
		}
	}
	for _, flow := range original.Spec.Spec {
		if !domainFunctionMap[flow.Function] {
			err := errors.New("Function " + flow.Name + " not defined in [" + original.Spec.Domain + "]")
			log.Error(err, "Workflow validation error")
			return ctrl.Result{}, nil
		}
	}

	// TODO: This is just an example status
	original.Status.Status = "Running"
	if err := r.Status().Update(ctx, &original); err != nil {
		log.Error(err, "unable to update status")
	}

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha1.Workflow{}).
		Complete(r)
}