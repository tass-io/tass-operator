package workflow

import (
	"fmt"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type generator struct {
	workflow *serverlessv1alpha1.Workflow
}

func newGenerator(wf *serverlessv1alpha1.Workflow) (*generator, error) {
	if wf == nil {
		return nil, fmt.Errorf("got nil when initializing Generator")
	}
	g := &generator{
		workflow: wf,
	}
	return g, nil
}

// desiredWorkflowRuntime returns a default config of WorkflowRuntime resource
func (g generator) desiredWorkflowRuntime() *serverlessv1alpha1.WorkflowRuntime {
	replicas := int32(2)
	podip := "localhost"
	return &serverlessv1alpha1.WorkflowRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: g.workflow.Namespace,
			Name:      g.workflow.Name,
		},
		// TODO: Provide customization future
		Spec: &serverlessv1alpha1.WorkflowRuntimeSpec{
			Replicas: &replicas,
			Status: serverlessv1alpha1.WfrtStatus{
				// NOTE: This part initializing is essential,
				// or the operator cannnot send a add json-patch action at the first time.
				Instances: serverlessv1alpha1.Instances{
					"init": serverlessv1alpha1.Instance{
						Status: &serverlessv1alpha1.InstanceStatus{
							PodIP: &podip,
						},
					},
				},
			},
		},
	}
}
