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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ReconcileNewDeployment(
	cli client.Client, req ctrl.Request,
	l logr.Logger, s *runtime.Scheme,
	i *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string) error {

	ctx := context.Background()
	log := l.WithValues("new deployment", req.NamespacedName)

	deploy := DesiredDeployment(
		req.NamespacedName.Namespace, req.NamespacedName.Name, labels)
	if err := ctrl.SetControllerReference(i, deploy, s); err != nil {
		return err
	}

	// try to see if the Deployment is already exists
	if err := cli.Get(ctx, req.NamespacedName, &appsv1.Deployment{}); errors.IsNotFound(err) {
		log.V(1).Info("Creating Deployment...")
		if err := cli.Create(ctx, deploy); err != nil {
			return err
		}
		// Successfully created a Deployment
		log.V(1).Info("Deployment Created successfully", "name", deploy.Name)
		return nil
	} else if err != nil {
		log.Error(err, "Cannot create Deployment")
		return err
	} else {
		log.V(1).Info("Deployment exists, no need to create a Deployment.")
	}
	return nil
}

func DesiredDeployment(namespace, name string, labels map[string]string) *appsv1.Deployment {
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}
	replicas := int32(2)
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
