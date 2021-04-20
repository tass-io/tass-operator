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

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileNewService creates a new Service resource,
// if the resource exists, it will ignore the request
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

	operationResult, err :=
		controllerutil.CreateOrUpdate(ctx, cli, svc, func() error { return nil })
	if err != nil {
		log.Error(err, "Cannot create/update Service")
		return err
	}
	log.V(1).Info("Service "+string(operationResult), "name", svc.Name)
	return nil
}

// DesiredService returns a default config of a Service
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
