package workflow

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
)

type Reconciler struct {
	cli      client.Client
	log      logr.Logger
	scheme   *runtime.Scheme
	instance *serverlessv1alpha1.Workflow
	gen      *generator
}

func NewReconciler(cli client.Client, l logr.Logger,
	s *runtime.Scheme, i *serverlessv1alpha1.Workflow) (*Reconciler, error) {
	g, err := newGenerator(i)
	if err != nil {
		return nil, err
	}
	return &Reconciler{
		cli:      cli,
		log:      l,
		scheme:   s,
		instance: i,
		gen:      g,
	}, nil
}

func (r *Reconciler) Reconcile() error {
	ctx := context.Background()
	log := r.log
	namespacedName := types.NamespacedName{
		Namespace: r.instance.Namespace,
		Name:      r.instance.Name,
	}

	wfrt := r.gen.desiredWorkflowRuntime()
	if err := ctrl.SetControllerReference(r.instance, wfrt, r.scheme); err != nil {
		return err
	}
	// try to see if the WorkflowRuntime is already exists
	if err := r.cli.Get(
		ctx, namespacedName, &serverlessv1alpha1.WorkflowRuntime{}); errors.IsNotFound(err) {
		if err := r.cli.Create(ctx, wfrt); err != nil {
			return err
		}
		// Successfully created a WorkflowRuntime
		log.Info("WorkflowRuntime Created successfully", "wfrt", namespacedName)
		return nil
	} else if err != nil {
		log.Error(err, "cannot create WorkflowRuntime", "wfrt", namespacedName)
		return err
	}

	return nil
}
