package spawn

import (
	"context"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
)

func ReconcileNewService(
	cli client.Client, req ctrl.Request,
	l logr.Logger, s *runtime.Scheme,
	i *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string) error {

	ctx := context.Background()
	log := l.WithValues("service", req.NamespacedName)

	svc := DesiredService(
		req.NamespacedName.Namespace, req.NamespacedName.Name, labels)
	if err := ctrl.SetControllerReference(i, svc, s); err != nil {
		return err
	}

	// try to see if the Service is already exists
	if err := cli.Get(ctx, req.NamespacedName, &corev1.Service{}); errors.IsNotFound(err) {
		log.V(1).Info("Creating Service...")
		if err := cli.Create(ctx, svc); err != nil {
			return err
		}
		// Successfully created a Service
		log.V(1).Info("Service Created successfully", "name", svc.Name)
		return nil
	} else if err != nil {
		log.Error(err, "Cannot create Service")
		return err
	} else {
		log.V(1).Info("Service exists, no need to create a Service.")
	}
	return nil
}

func DesiredService(namespace, name string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol: "TCP",
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: 80,
					},
				},
			},
		},
	}

}
