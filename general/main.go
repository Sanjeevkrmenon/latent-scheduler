package main

import (
    "context"
    "time"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/klog/v2"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

const schedulerName = "my-latency-scheduler"

func main() {
    klog.Info("Starting custom scheduler...")

    // Use in-cluster configuration
    config, err := rest.InClusterConfig()
    if err != nil {
        klog.Fatalf("Error creating in-cluster config: %v", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        klog.Fatalf("Error creating Kubernetes client: %v", err)
    }

    // Main scheduling loop
    for {
        // List all pending pods
        pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
            FieldSelector: "status.phase=Pending",
        })
        if err != nil {
            klog.ErrorS(err, "Failed to list pods")
            time.Sleep(2 * time.Second)
            continue
        }

        // Process pending pods
        for _, pod := range pods.Items {
            if pod.Spec.SchedulerName != schedulerName || pod.Spec.NodeName != "" {
                continue
            }

            // List all available nodes
            nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
            if err != nil || len(nodes.Items) == 0 {
                klog.ErrorS(err, "Failed to list nodes")
                continue
            }

            // Choose the first node
            chosenNode := nodes.Items[0].Name
            binding := &corev1.Binding{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      pod.Name,
                    Namespace: pod.Namespace,
                },
                Target: corev1.ObjectReference{
                    Kind: "Node",
                    Name: chosenNode,
                },
            }

            // Bind the pod to the chosen node
            err = clientset.CoreV1().Pods(pod.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
            if err != nil {
                klog.ErrorS(err, "Failed to bind pod", "pod", pod.Name, "node", chosenNode)
            } else {
                klog.InfoS("Pod bound", "pod", pod.Name, "node", chosenNode)
            }
        }

        // Wait before the next iteration
        time.Sleep(2 * time.Second)
    }
}