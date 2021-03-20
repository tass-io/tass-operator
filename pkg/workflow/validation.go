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
		if pre.Spec.Domain == wf.Spec.Domain {
			domainFunctionMap[pre.Name] = true
		}
	}
	for _, flow := range wf.Spec.Spec {
		if !domainFunctionMap[flow.Function] {
			return errors.New("Function " + flow.Function + " not defined in [" + wf.Spec.Domain + "]")
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
			return errors.New("Flow " + flow.Name + " has defined more than once")
		}
		flowMap[flow.Name] = &wf.Spec.Spec[i]
		// find Flow entrance
		if len(flow.Inputs) == 0 {
			if entrance != nil {
				return errors.New("Flows has multi entrance")
			}
			entrance = &wf.Spec.Spec[i]
		}
		if len(flow.Outputs) == 0 {
			hasExit = true
		}
	}

	if entrance == nil {
		return errors.New("Flows has no entrance")
	}
	if !hasExit {
		return errors.New("Flows has no exit")
	}

	for _, flow := range wf.Spec.Spec {
		// check inputs
		for _, input := range flow.Inputs {
			if _, ok := flowMap[input]; !ok {
				return errors.New("Input " + input + " in Flow " + flow.Name + " has not define")
			}
		}
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

		// check destination in condition
		if flow.Condition == nil {
			return errors.New("Condition should be defined when the Statement is not 'direct'")
		}
		for _, item := range flow.Condition.Destination.IsTrue {
			if _, ok := outputMap[item]; !ok {
				return errors.New("Destination " + item + " in Flow " + flow.Name + " has not define in its output")
			}
		}
		for _, item := range flow.Condition.Destination.IsFalse {
			if _, ok := outputMap[item]; !ok {
				return errors.New("Destination " + item + " in Flow " + flow.Name + " has not define in its output")
			}
		}
	}
	return nil
}
