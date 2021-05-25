package workflowruntime

import (
	"fmt"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	defaultRole = "tass-operator"
	// local scheduler image info
	imageName     = "registry.cn-hangzhou.aliyuncs.com/tass/local-scheduler"
	imageVersion  = "v0.1.1"
	containerPort = 8080
)

type generator struct {
	workflowruntime *serverlessv1alpha1.WorkflowRuntime
	labels          map[string]string
}

func newGenerator(wfrt *serverlessv1alpha1.WorkflowRuntime,
	labels map[string]string) (*generator, error) {
	if wfrt == nil {
		return nil, fmt.Errorf("got nil when initializing Generator")
	}
	g := &generator{
		workflowruntime: wfrt,
		labels:          labels,
	}
	return g, nil
}

// desiredService returns a default config of a Service
func (g generator) desiredService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: g.workflowruntime.Namespace,
			Name:      g.workflowruntime.Name,
			Labels:    g.labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: g.labels,
			Ports: []corev1.ServicePort{
				{
					Protocol: "TCP",
					Port:     8080,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: 8080,
					},
				},
			},
		},
	}
}

// desiredDeploymentWithServiceAccount returns a default config of a Deployment
func (g generator) desiredDeploymentWithServiceAccount(sa string) *appsv1.Deployment {
	trueFlag := true
	selector := &metav1.LabelSelector{
		MatchLabels: g.labels,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: g.workflowruntime.Namespace,
			Name:      g.workflowruntime.Name,
			Labels:    g.labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: selector,
			Replicas: g.workflowruntime.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: g.labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: sa,
					Containers: []corev1.Container{
						{
							Name:  "scheduler",
							Image: imageName + ":" + imageVersion,
							Ports: []corev1.ContainerPort{{
								ContainerPort: containerPort,
								Protocol:      "TCP",
							}},
							// Args: []string{"-m"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &trueFlag,
							},
						},
					},
				},
			},
		},
	}
}

// desiredServiceAccount returns a ServiceAccount without owner
func (g generator) desiredServiceAccount() *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind: "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: g.workflowruntime.Namespace,
			Name:      g.workflowruntime.Name,
			Labels:    g.labels,
		},
	}
	return sa
}

// desiredRoleBinding binds a ServiceAccount and a Role
// Each Deployment has a ServiceAccount
// Role has pre-defined in `hack/prepare.yaml`
func (g generator) desiredRoleBinding(sa *corev1.ServiceAccount) *rbacv1.RoleBinding {
	// RoleBinding and ServiceAccount use same Namespace and Name naming
	crb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: g.workflowruntime.Namespace,
			Name:      g.workflowruntime.Name,
			Labels:    g.labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: sa.Namespace,
				Name:      sa.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     defaultRole,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	return crb
}
