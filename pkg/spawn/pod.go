package spawn

import (
	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FIXME: This is an example of creating a Pod using `httpbin`
// Put the real content in the future
func NewPodForCR(cr serverlessv1alpha1.Function) *corev1.Pod {
	labels := map[string]string{
		"function": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
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
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}
