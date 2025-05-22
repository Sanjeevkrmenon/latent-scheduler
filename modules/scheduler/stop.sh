#!/bin/bash

# Delete deployment
kubectl delete deployment my-latency-scheduler -n kube-system --ignore-not-found

# Delete service account
kubectl delete serviceaccount my-latency-scheduler-sa -n kube-system --ignore-not-found

# Delete cluster role
kubectl delete clusterrole my-latency-scheduler-role --ignore-not-found

# Delete cluster role binding
kubectl delete clusterrolebinding my-latency-scheduler-rolebinding --ignore-not-found

echo "Resources cleaned up."