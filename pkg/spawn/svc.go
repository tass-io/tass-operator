package spawn

import (
	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// FIXME: This is an example of creating a Pod using 'busybox`
// Put the real content in the future
func NewSvcForCR(cr serverlessv1alpha1.Function) *corev1.Service {
	labels := map[string]string{
		"function": cr.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeNodePort,
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
