package spawn

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReconcileNewWorkflowRuntime creates a new Service resource,
// if the resource exists, it will ignore the request
func ReconcileNewWorkflowRuntime(
	cli client.Client, req ctrl.Request,
	l logr.Logger, s *runtime.Scheme,
	i *serverlessv1alpha1.Workflow) error {

	ctx := context.Background()
	log := l.WithValues("new workflowruntime", req.NamespacedName)

	wfrt := DesiredWorkflowRuntime(req.NamespacedName.Namespace, req.NamespacedName.Name)
	if err := ctrl.SetControllerReference(i, wfrt, s); err != nil {
		return err
	}

	// try to see if the WorkflowRuntime is already exists
	if err := cli.Get(ctx, req.NamespacedName, &serverlessv1alpha1.WorkflowRuntime{}); errors.IsNotFound(err) {
		if err := cli.Create(ctx, wfrt); err != nil {
			return err
		}
		// Successfully created a WorkflowRuntime
		log.V(1).Info("WorkflowRuntime Created successfully", "name", wfrt.Name)
		return nil
	} else if err != nil {
		log.Error(err, "Cannot create WorkflowRuntime")
		return err
	}
	return nil
}

// DesiredWorkflowRuntime returns a default config of WorkflowRuntime resource
func DesiredWorkflowRuntime(namespace, name string) *serverlessv1alpha1.WorkflowRuntime {
	return &serverlessv1alpha1.WorkflowRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		// TODO: Provide customization future
		Spec: serverlessv1alpha1.WorkflowRuntimeSpec{
			Replicas: 2,
		},
	}
}
