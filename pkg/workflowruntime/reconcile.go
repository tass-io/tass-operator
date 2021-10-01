package workflowruntime

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
)

type Reconciler struct {
	cli      client.Client
	log      logr.Logger
	scheme   *runtime.Scheme
	instance *serverlessv1alpha1.WorkflowRuntime
	gen      *generator
}

func NewReconciler(cli client.Client, l logr.Logger,
	s *runtime.Scheme, i *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string) (*Reconciler, error) {

	g, err := newGenerator(i, labels)
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
	serviceAccountName, err := r.reconcileRBAC()
	if err != nil {
		return err
	}
	if err := r.reconcileDeployment(serviceAccountName); err != nil {
		return err
	}

	if err := r.reconcileService(); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) reconcileRBAC() (string, error) {
	sa, err := r.reconcileServiceAccount()
	if err != nil {
		return "", err
	}
	if err := r.reconcileRoleBinding(sa); err != nil {
		return "", err
	}
	return sa.Name, nil
}

func (r *Reconciler) reconcileServiceAccount() (*corev1.ServiceAccount, error) {
	ctx := context.Background()
	namespacedName := types.NamespacedName{
		Namespace: r.instance.Namespace,
		Name:      r.instance.Name,
	}
	log := r.log.WithValues("serviceaccount", namespacedName)

	desired := r.gen.desiredServiceAccount()
	if err := ctrl.SetControllerReference(r.instance, desired, r.scheme); err != nil {
		return nil, err
	}

	actual := &corev1.ServiceAccount{}
	err := r.cli.Get(ctx, types.NamespacedName{
		Name:      desired.GetName(),
		Namespace: desired.GetNamespace()},
		actual)
	if err != nil && k8serrors.IsNotFound(err) {
		if err := r.cli.Create(ctx, desired); err != nil {
			log.Error(err, "failed to create the serviceaccount")
			return nil, err
		}
	} else if err != nil {
		log.Error(err, "failed to get the expected serviceaccount")
		return nil, err
	}
	// When the sa is created, actual is nil. Thus actual cannot be used to build rolebinding.
	return desired, nil
}

func (r *Reconciler) reconcileRoleBinding(sa *corev1.ServiceAccount) error {
	ctx := context.Background()
	namespacedName := types.NamespacedName{
		Namespace: r.instance.Namespace,
		Name:      r.instance.Name,
	}
	log := r.log.WithValues("rolebinding", namespacedName)

	desired := r.gen.desiredRoleBinding(sa)

	if err := ctrl.SetControllerReference(r.instance, desired, r.scheme); err != nil {
		log.Error(err, "set controller reference error, requeuing the request")
		return err
	}

	actual := &rbacv1.RoleBinding{}
	err := r.cli.Get(ctx, types.NamespacedName{
		Name:      desired.GetName(),
		Namespace: desired.GetNamespace()},
		actual)
	if err != nil && k8serrors.IsNotFound(err) {
		if err := r.cli.Create(ctx, desired); err != nil {
			log.Error(err, "failed to create the rolebinding")
			return err
		}
	} else if err != nil {
		log.Error(err, "failed to get the expected rolebinding")
		return err
	}
	return nil
}

// reconcileDeployment creates a new Deploy resource or updates an existing Deploy
func (r *Reconciler) reconcileDeployment(serviceAccountName string) error {
	ctx := context.Background()
	namespacedName := types.NamespacedName{
		Namespace: r.instance.Namespace,
		Name:      r.instance.Name,
	}
	log := r.log.WithValues("deployment", namespacedName)

	deploy := r.gen.desiredDeploymentWithServiceAccount(serviceAccountName)
	if err := ctrl.SetControllerReference(r.instance, deploy, r.scheme); err != nil {
		return err
	}

	// deployMutateFn is called regardless of creating or updating an object.
	// If it's a `create` action, it creates a new resource, and the `replicas` is the default value
	// If it's an `update` action, it updates the resource with the new `replicas`
	deployMutateFn := func() error {
		deploy.Spec.Replicas = r.instance.Spec.Replicas
		return nil
	}

	operationResult, err := controllerutil.CreateOrUpdate(ctx, r.cli, deploy, deployMutateFn)
	if err != nil {
		log.Error(err, "cannot create/update Deployment")
		return err
	}
	log.Info("Deployment " + string(operationResult))
	return nil
}

// reconcileService creates a new Service resource,
// if the resource exists, it will ignore the request
func (r *Reconciler) reconcileService() error {
	ctx := context.Background()
	namespacedName := types.NamespacedName{
		Namespace: r.instance.Namespace,
		Name:      r.instance.Name,
	}
	log := r.log.WithValues("service", namespacedName)

	svc := r.gen.desiredService()
	if err := ctrl.SetControllerReference(r.instance, svc, r.scheme); err != nil {
		return err
	}

	operationResult, err :=
		controllerutil.CreateOrUpdate(ctx, r.cli, svc, func() error { return nil })
	if err != nil {
		log.Error(err, "cannot create/update Service")
		return err
	}
	log.Info("Service " + string(operationResult))
	return nil
}
