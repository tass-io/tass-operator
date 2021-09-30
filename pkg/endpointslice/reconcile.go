package endpointslice

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-logr/logr"
	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	"github.com/tass-io/tass-operator/pkg/utils/jsonpatch"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	cli      client.Client
	log      logr.Logger
	scheme   *runtime.Scheme
	instance *discoveryv1beta1.EndpointSlice
}

func NewReconciler(cli client.Client, l logr.Logger,
	s *runtime.Scheme, i *discoveryv1beta1.EndpointSlice) (*Reconciler, error) {
	return &Reconciler{
		cli:      cli,
		log:      l,
		scheme:   s,
		instance: i,
	}, nil
}

// Reconcile contains 4 steps:
// 1. Record the endpoints pod and address
// 2. Get the corresponding WorkflowRuntime instance
// 3. Check the change of the endpoint address, add the new ip and remove the deprecated with nil
// 4. Marshal to patch bytes and patch the result
func (r Reconciler) Reconcile() error {
	ctx := context.Background()
	log := r.log.WithValues("endpointslice", types.NamespacedName{
		Namespace: r.instance.Namespace,
		Name:      r.instance.Name,
	})

	// 1. Record the endpoints pod and address
	//
	// currentSvcMesh records the network status in this endpointslice
	// the key is the pod self name, and the value is ip address
	currentSvcMesh := map[string]string{}
	for _, item := range r.instance.Endpoints {
		itemName := getPodSelfName(item.TargetRef.Name)
		currentSvcMesh[itemName] = item.Addresses[0]
	}
	log.Info("get endpointslice info successfully")

	// 2. Get the corresponding WorkflowRuntime instance
	//
	wfrtNamespacedName := types.NamespacedName{
		Namespace: r.instance.Namespace,
		Name:      getWorkflowRuntimeName(r.instance.Name),
	}
	var wfrt serverlessv1alpha1.WorkflowRuntime
	if err := r.cli.Get(ctx, wfrtNamespacedName, &wfrt); err != nil {
		return err
	}

	// 3. Check the change of the endpoint address, add the new ip and remove the deprecated with nil
	//
	log.Info("check the change of the endpoint address")
	jsonPatchItems := []jsonpatch.JsonPatchItem{}
	// 3.1 check the existed info of WorkflowRuntime resource instance
	//
	for name := range wfrt.Spec.Status.Instances {
		address, ok := currentSvcMesh[name]
		if ok {
			// put address into wfrt
			newItem := jsonpatch.JsonPatchItem{
				Op:   jsonpatch.OperationReplace,
				Path: jsonpatch.SetPath(false, "spec", "status", "instances", name, "status"),
				Value: serverlessv1alpha1.InstanceStatus{
					PodIP: &address,
				},
			}
			jsonPatchItems = append(jsonPatchItems, newItem)
			delete(currentSvcMesh, name)
		} else {
			// this pod is terminated, delete info in wfrt
			newItem := jsonpatch.JsonPatchItem{
				Op:   jsonpatch.OperationRemove,
				Path: jsonpatch.SetPath(false, "spec", "status", "instances", name),
			}
			jsonPatchItems = append(jsonPatchItems, newItem)
		}
	}
	// 3.2 the rest currentSvcMesh objects are new elements
	//
	for name, address := range currentSvcMesh {
		newItem := jsonpatch.JsonPatchItem{
			Op:   jsonpatch.OperationAdd,
			Path: jsonpatch.SetPath(false, "spec", "status", "instances", name),
			Value: serverlessv1alpha1.Instance{
				Status: &serverlessv1alpha1.InstanceStatus{PodIP: &address},
			},
		}
		jsonPatchItems = append(jsonPatchItems, newItem)
	}

	// 4. Marshal to patch bytes and patch the result
	//
	patchBytes, _ := json.Marshal(jsonPatchItems)
	if err := r.cli.Patch(ctx, &serverlessv1alpha1.WorkflowRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: wfrtNamespacedName.Namespace,
			Name:      wfrtNamespacedName.Name,
		},
	}, client.RawPatch(types.JSONPatchType, patchBytes)); err != nil {
		return err
	}
	log.Info("update the Workflow runtime " + wfrtNamespacedName.String() + " successfully")

	return nil
}

// getPodSelfName returns the pod slef name from endpointslice item TargetRef
// e.g. (workflow-sample-9657bf88d-btxwt) => (9657bf88d-btxwt)
func getPodSelfName(name string) string {
	sli := strings.Split(name, "-")
	selfName := strings.Join(sli[len(sli)-2:], "-")
	return selfName
}

// getWorkflowRuntimeName return the wfrt name from the endpointslice name
// e.g. (workflow-sample-qk4ng) => (workflow-sample)
func getWorkflowRuntimeName(name string) string {
	sli := strings.Split(name, "-")
	wfrtName := strings.Join(sli[:len(sli)-1], "-")
	return wfrtName
}
