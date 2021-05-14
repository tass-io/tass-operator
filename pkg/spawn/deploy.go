package spawn

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileNewDeployment creates a new Deploy resource or updates an existing Deploy
func ReconcileNewDeployment(
	cli client.Client, req ctrl.Request,
	l logr.Logger, s *runtime.Scheme,
	i *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string, replicas int32) error {

	ctx := context.Background()
	log := l.WithValues("new deployment", req.NamespacedName)

	deploy := DesiredDeployment(
		req.NamespacedName.Namespace, req.NamespacedName.Name, labels, replicas)
	if err := ctrl.SetControllerReference(i, deploy, s); err != nil {
		return err
	}

	// deployMutateFn is called regardless of creating or updating an object.
	// If it's a `create` action, it creates a new resource, and the `replicas` is the default value
	// If it's an `update` action, it updates the resource with the new `replicas`
	deployMutateFn := func() error {
		deploy.Spec.Replicas = &replicas
		return nil
	}

	operationResult, err := controllerutil.CreateOrUpdate(ctx, cli, deploy, deployMutateFn)
	if err != nil {
		log.Error(err, "Cannot create/update Deployment")
		return err
	}
	log.V(1).Info("Deployment "+string(operationResult), "name", deploy.Name)
	return nil
}

// DesiredDeployment returns a default config of a Deployment
func DesiredDeployment(namespace, name string, labels map[string]string,
	replicas int32) *appsv1.Deployment {
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: selector,
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					// TODO: replace it with real local scheduler
					Containers: []corev1.Container{
						{
							Name:  "httpbin",
							Image: "kennethreitz/httpbin",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 80,
								Protocol:      "TCP",
							}},
						},
					},
				},
			},
		},
	}
}
