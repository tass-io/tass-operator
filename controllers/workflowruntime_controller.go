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
	"github.com/tass-io/tass-operator/pkg/endpointslice"
	"github.com/tass-io/tass-operator/pkg/workflowruntime"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// WorkflowRuntimeReconciler reconciles a WorkflowRuntime object
type WorkflowRuntimeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// nolint
// +kubebuilder:rbac:groups=serverless.tass.io,resources=workflowruntimes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.tass.io,resources=workflowruntimes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;list;watch;create;update;patch;delete

func (r *WorkflowRuntimeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("workflowruntime", req.NamespacedName)

	var eps discoveryv1beta1.EndpointSlice
	if err := r.Get(ctx, req.NamespacedName, &eps); err == nil {
		log.Info("fetch endpointslice successfully")
		neweps := eps.DeepCopy()
		epsr, _ := endpointslice.NewReconciler(r.Client, log, r.Scheme, neweps)
		if err := epsr.Reconcile(); err != nil {
			return ctrl.Result{}, err
		}
	}

	var original serverlessv1alpha1.WorkflowRuntime
	if err := r.Get(ctx, req.NamespacedName, &original); err != nil {
		log.Error(err, "unable to fetch WorkflowRuntime")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// A WorkflowRuntime has its Service and Deployment which
	// run local scheduler and Flows in Workflow when a request comes
	labels := map[string]string{
		"type": "workflowRuntime",
		"name": req.NamespacedName.Name,
	}
	instance := original.DeepCopy()
	wfrtr, err := workflowruntime.NewReconciler(r.Client, log, r.Scheme, instance, labels)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := wfrtr.Reconcile(); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil

}

func (r *WorkflowRuntimeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha1.WorkflowRuntime{}).
		Watches(
			&source.Kind{Type: &discoveryv1beta1.EndpointSlice{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(r.findObjsForEndpointSlice),
			},
		).
		Complete(r)
}

// findObjsForEndpointSlice is used to find an endpointslice for workflowruntime.
// This func is the implementation of `ToRequestsFunc func(MapObject) []reconcile.Request`.
// When the controller manager starts, it iterates all endpointslices and chooses suitable resluts
// For the chosen object, they will be replaced in `reconcile.Request` slice,
// and then this request will be reconciled by `WorkflowRuntimeReconciler.Reconcile` func.
func (r *WorkflowRuntimeReconciler) findObjsForEndpointSlice(endpointsliceMap handler.MapObject) []reconcile.Request {
	ns := endpointsliceMap.Meta.GetNamespace()
	endpointsliceLabels := endpointsliceMap.Meta.GetLabels()
	// this name is the name of the Service name and is the same as Workflow name
	// so we can get WorkflowRuntime by namespace & name
	name, ok := endpointsliceLabels["kubernetes.io/service-name"]
	if !ok {
		return []reconcile.Request{}
	}

	wfrt := serverlessv1alpha1.WorkflowRuntime{}
	if err := r.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, &wfrt); err == nil {
		// such wfrt exists, then send the request to the `WorkflowRuntimeReconciler.Reconcile` function
		// TODO: note that here we use an naive way to send the request, so the Reconcile func must send a
		// request to the apiserver to check the resource type first. We should make the process more efficient.
		// for example, we can modify the name here like:
		//
		// {
		// 	NamespacedName: types.NamespacedName{
		// 		Name:      endpointsliceMap.Meta.GetName() + "-endpointslice",
		// 		Namespace: ns,
		// 	},
		// }
		// and then we can use the `WorkflowRuntimeReconciler.Reconcile` func to parse the resource
		//
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      endpointsliceMap.Meta.GetName(),
					Namespace: ns,
				},
			},
		}
	}

	return []reconcile.Request{}
}
