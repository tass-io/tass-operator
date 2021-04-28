package workflow

import (
	"errors"

	serverlessv1alpha1 "github.com/tass-io/tass-operator/api/v1alpha1"
)

// ValidateFuncExist validates that each Function declared in the workflow
// has been defined in Function CRD, or it will return error
func ValidateFuncExist(wf *serverlessv1alpha1.Workflow, fl *serverlessv1alpha1.FunctionList) error {
	domainFunctionMap := map[string]bool{}
	for _, pre := range fl.Items {
		domainFunctionMap[pre.Name] = true
	}
	for _, flow := range wf.Spec.Spec {
		if !domainFunctionMap[flow.Function] {
			return errors.New("function " + flow.Function + " not defined in namespace [" + wf.Namespace + "]")
		}
	}
	return nil
}

// ValidateFlows validates wether the graph of Flows is legal or not
// For every Flow, it should obey the following rules:
// - Has one and only one entrance
// - Has one and at least one exit
// - Every Flow in Inputs & Outputs should have been defined in []Flow
// - If a Flow has a Condition, every Flow in Condition.Destination
//   should have been defined in Outputs
func ValidateFlows(wf *serverlessv1alpha1.Workflow) error {
	flowMap := map[string]*serverlessv1alpha1.Flow{}
	var entrance *serverlessv1alpha1.Flow
	var hasExit bool
	for i, flow := range wf.Spec.Spec {
		if _, ok := flowMap[flow.Name]; ok {
			return errors.New("flow " + flow.Name + " has defined more than once")
		}
		flowMap[flow.Name] = &wf.Spec.Spec[i]
		// TODO: find Flow entrance
		if len(flow.Outputs) == 0 {
			hasExit = true
		}
	}

	if entrance == nil {
		return errors.New("flows has no entrance")
	}
	if !hasExit {
		return errors.New("flows has no exit")
	}

	for _, flow := range wf.Spec.Spec {
		// check outputs
		outputMap := map[string]bool{}
		for _, output := range flow.Outputs {
			if _, ok := flowMap[output]; !ok {
				return errors.New("Output " + output + " in Flow " + flow.Name + " has not define")
			}
			outputMap[output] = true
		}
		if flow.Statement == "direct" {
			return nil
		}

		// TODO: More Conditions check
		if flow.Conditions == nil {
			return errors.New("condition should be defined when the Statement is not 'direct'")
		}
	}
	return nil
}
