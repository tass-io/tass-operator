#!/bin/bash

# Use this script when you want to send a request directly to the pod.

echo "Send a request to a pod..."
echo "The response is:"
POD_IP=$(kubectl get po -o wide  | grep workflow | awk '{print $6}' | sed -n '1p')
curl --request POST "http://${POD_IP}/v1/workflow/" --header 'Content-Type: application/json' --data-raw '{"workflowName": "workflow-sample", "flowName": "", "parameters": {}}'

sleep 2s
echo -e "\n"
echo "Now the WorkflowRuntime workflow-sample is:"
kubectl get workflowruntime workflow-sample -o=jsonpath='{.spec.status.instances}'
