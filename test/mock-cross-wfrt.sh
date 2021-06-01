#!/bin/bash

# It's not easy to observe this situation in a short time.
# Local shceduler has a option to enforce every function goes to lsds.
# This script mocks functions in a workflow are placed in different processes.
#
WORKFLOW="workflow-sample"
POD_LIST=$(kubectl get po -o wide  | grep workflow | awk '{print $1}' | sed -n '1p')
POD_ONE=$(echo ${POD_LIST} | sed -n '1p')
POD_TWO=$(echo ${POD_LIST} | sed -n '2p')

kubectl patch workflowruntime ${WORKFLOW} --type=json -p='
- op: add
  path: /spec/status/instances/'"${POD_ONE: 0-16}"'/processRuntimes
  value:
    function2:
      number: 1
- op: add
  path: /spec/status/instances/'"${POD_TWO: 0-16}"'/processRuntimes
  value:
    function1:
      number: 1
    function3:
      number: 1
'
