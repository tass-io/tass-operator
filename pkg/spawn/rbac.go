package spawn

import (
	"context"

	"github.com/go-logr/logr"
	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultClusterRole = "tass-operator"
)

func ReconcileRBAC(
	cli client.Client, req ctrl.Request,
	l logr.Logger, s *runtime.Scheme,
	i *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string) (string, error) {
	sa, err := ReconcileServiceAccount(cli, req, l, s, i, labels)
	if err != nil {
		return "", err
	}
	if err := ReconcileRoleBinding(cli, req, l, s, i, labels, sa); err != nil {
		return "", err
	}
	return sa.Name, nil
}

func ReconcileServiceAccount(
	cli client.Client, req ctrl.Request,
	l logr.Logger, s *runtime.Scheme,
	i *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string) (*corev1.ServiceAccount, error) {

	ctx := context.Background()
	log := l.WithValues("serviceaccount", req.NamespacedName)

	desired := DesiredServiceAccount(req.Namespace, req.Name, labels)
	if err := ctrl.SetControllerReference(i, desired, s); err != nil {
		return nil, err
	}

	actual := &corev1.ServiceAccount{}
	err := cli.Get(ctx, types.NamespacedName{
		Name:      desired.GetName(),
		Namespace: desired.GetNamespace()},
		actual)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating serviceaccount", "namespace", desired.Namespace, "name", desired.Name)

		if err := cli.Create(ctx, desired); err != nil {
			log.Error(err, "Failed to create the serviceaccount",
				"serviceaccount", desired.Name)
			return nil, err
		}
	} else if err != nil {
		log.Error(err, "failed to get the expected serviceaccount",
			"serviceaccount", desired.Name)
		return nil, err
	}
	// When the sa is created, actual is nil. Thus actual cannot be used to build rolebinding.
	return desired, nil
}

// DesiredServiceAccount returns a ServiceAccount without owner
func DesiredServiceAccount(
	namespace, name string, labels map[string]string) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind: "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
	}
	return sa
}

func ReconcileRoleBinding(
	cli client.Client, req ctrl.Request,
	l logr.Logger, s *runtime.Scheme,
	i *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string,
	sa *corev1.ServiceAccount) error {

	ctx := context.Background()
	log := l.WithValues("rolebinding", req.NamespacedName)

	desired := DesiredRoleBinding(sa)

	if err := ctrl.SetControllerReference(i, desired, s); err != nil {
		log.Error(err, "Set controller reference error, requeuing the request")
		return err
	}

	actual := &rbacv1.RoleBinding{}
	err := cli.Get(ctx, types.NamespacedName{
		Name:      desired.GetName(),
		Namespace: desired.GetNamespace()},
		actual)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating rolebinding", "namespace", desired.Namespace, "name", desired.Name)

		if err := cli.Create(ctx, desired); err != nil {
			log.Error(err, "Failed to create the rolebinding", "rolebinding", desired.Name)
			return err
		}
	} else if err != nil {
		log.Error(err, "failed to get the expected rolebinding", "rolebinding", desired.Name)
		return err
	}
	return nil
}

// DesiredRoleBinding binds a ServiceAccount and a ClusterRole
// Each Deployment has a ServiceAccount
// ClusterRole has pre-defined in `hack/prepare.yaml`
func DesiredRoleBinding(sa *corev1.ServiceAccount) *rbacv1.ClusterRoleBinding {
	// ClusterRoleBinding and ServiceAccount use same Namespace and Name naming
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sa.Namespace,
			Name:      sa.Name,
			Labels:    sa.Labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: sa.Namespace,
				Name:      sa.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     defaultClusterRole,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	return crb
}
