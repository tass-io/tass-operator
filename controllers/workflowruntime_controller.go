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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
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

// +kubebuilder:rbac:groups=serverless.tass.io,resources=workflowruntimes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.tass.io,resources=workflowruntimes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;list;watch;create;update;patch;delete

func (r *WorkflowRuntimeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("workflowruntime", req.NamespacedName)

	var endpointslice discoveryv1beta1.EndpointSlice
	if err := r.Get(ctx, req.NamespacedName, &endpointslice); err == nil {
		// TODO: See below
		// 1. After getting the endpoint successfully
		// 2. Get the corresponding WFRT instance
		// 3. Judge the change of the endpoint address
		// 4. Patch the new ip and remove the old with nil
		log.Info("fetch endpointslice successfully")
		fmt.Println("---")
		for _, item := range endpointslice.Endpoints {
			fmt.Println(item.Addresses)
		}
		fmt.Println("---")
	}

	var original serverlessv1alpha1.WorkflowRuntime
	if err := r.Get(ctx, req.NamespacedName, &original); err != nil {
		log.Error(err, "unable to fetch WorkflowRuntime")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var deploy appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deploy); err == nil {
		log.Info("fetch Deployment successfully")
		fmt.Println("---")
		fmt.Println(deploy.Name)
		fmt.Println("---")
	}

	var service corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &service); err == nil {
		log.Info("fetch service successfully")
		fmt.Println("---")
		fmt.Println(service.Name)
		fmt.Println("---")
	}

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err == nil {
		log.Info("fetch Pod successfully")
		fmt.Println("---")
		fmt.Println(pod.Name)
		fmt.Println("---")
	}

	// A WorkflowRuntime has its Service and Deployment which
	// run local scheduler and Flows in Workflow when a request comes
	labels := map[string]string{
		"type": "workflowRuntime",
		"name": req.NamespacedName.Name,
	}
	instance := original.DeepCopy()
	wfrtr, err := workflowruntime.NewReconciler(r.Client, r.Log, r.Scheme, instance, labels)
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
		Owns(&appsv1.Deployment{}).
		// Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		// Watches(
		// 	&source.Kind{Type: &corev1.Pod{}},
		// 	&handler.EnqueueRequestsFromMapFunc{
		// 		ToRequests: handler.ToRequestsFunc(r.findObjectsForPod),
		// 	},
		// ).
		Watches(
			&source.Kind{Type: &discoveryv1beta1.EndpointSlice{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(r.findObjectsForEndpointSlice),
			},
		).
		Complete(r)
}

// 第一参数是指集群内的所有 pod
func (r *WorkflowRuntimeReconciler) findObjectsForPod(podMap handler.MapObject) []reconcile.Request {
	podList := &corev1.PodList{}
	if err := r.List(context.Background(), podList,
		client.InNamespace(podMap.Meta.GetNamespace()), client.MatchingLabels{
			"name": "workflow-sample",
			"type": "workflowRuntime",
			// TODO: Update this field later
			// "endpointslice.kubernetes.io/managed-by": "endpointslice-controller.k8s.io",
			// "kubernetes.io/service-name": "workflow-sample",
		}); err != nil {
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(podList.Items))
	for i, item := range podList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// 第一参数是指集群内的所有 pod
func (r *WorkflowRuntimeReconciler) findObjectsForEndpointSlice(endpointsliceMap handler.MapObject) []reconcile.Request {
	list := &discoveryv1beta1.EndpointSliceList{}
	if err := r.List(context.Background(), list,
		client.InNamespace(endpointsliceMap.Meta.GetNamespace()), client.MatchingLabels{
			// TODO: Update this field later
			"endpointslice.kubernetes.io/managed-by": "endpointslice-controller.k8s.io",
			"kubernetes.io/service-name":             "workflow-sample",
		}); err != nil {
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(list.Items))
	for i, item := range list.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}
