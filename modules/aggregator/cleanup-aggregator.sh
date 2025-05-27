#!/bin/bash

set -e

NAMESPACE="kube-system"
MATCH="aggregator"

echo "Deleting CronJobs related to aggregator..."
kubectl -n $NAMESPACE get cronjobs --no-headers | grep $MATCH | awk '{print $1}' | xargs -r kubectl -n $NAMESPACE delete cronjob

echo "Deleting Jobs related to aggregator..."
kubectl -n $NAMESPACE get jobs --no-headers | grep $MATCH | awk '{print $1}' | xargs -r kubectl -n $NAMESPACE delete job

echo "Deleting Deployments related to aggregator..."
kubectl -n $NAMESPACE get deployments --no-headers | grep $MATCH | awk '{print $1}' | xargs -r kubectl -n $NAMESPACE delete deployment

echo "Deleting ReplicaSets related to aggregator..."
kubectl -n $NAMESPACE get rs --no-headers | grep $MATCH | awk '{print $1}' | xargs -r kubectl -n $NAMESPACE delete rs

echo "Deleting Pods related to aggregator..."
kubectl -n $NAMESPACE get pods --no-headers | grep $MATCH | awk '{print $1}' | xargs -r kubectl -n $NAMESPACE delete pod

echo "Cleanup complete."