package main

import (
    "context"
    "encoding/json"
    "io/ioutil"
    "math"
    "time"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/klog/v2"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

const (
    schedulerName      = "my-latency-scheduler"
    latencyFile        = "/latency/cluster-latency.json"
    latencyFallbackVal = math.MaxFloat64
)

type LatencyMap map[string]map[string]float64

func loadLatencies(filename string) (LatencyMap, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    var lm LatencyMap
    err = json.Unmarshal(data, &lm)
    return lm, err
}

func nodeAvgLatency(nodeName string, latencies LatencyMap) float64 {
    sum := 0.0
    cnt := 0
    if m, ok := latencies[nodeName]; ok {
        for n, latency := range m {
            if n != nodeName && latency > 0 {
                sum += latency
                cnt++
            }
        }
    }
    if cnt == 0 {
        return latencyFallbackVal
    }
    return sum / float64(cnt)
}

func main() {
    klog.Infof("Starting custom latency scheduler: %s", schedulerName)

    config, err := rest.InClusterConfig()
    if err != nil {
        klog.Fatalf("Error creating in-cluster config: %v", err)
    }
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        klog.Fatalf("Error creating Kubernetes client: %v", err)
    }

    for {
        pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
            FieldSelector: "status.phase=Pending",
        })
        if err != nil {
            klog.ErrorS(err, "Failed to list pods")
            time.Sleep(2 * time.Second)
            continue
        }

        latencies, err := loadLatencies(latencyFile)
        if err != nil {
            klog.Warning("Could not load latency mesh, will fallback to first node.")
        }

        for _, pod := range pods.Items {
            if pod.Spec.SchedulerName != schedulerName || pod.Spec.NodeName != "" {
                continue
            }

            nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
            if err != nil || len(nodes.Items) == 0 {
                klog.ErrorS(err, "Failed to list nodes")
                continue
            }

            // Pick best node by average RTT
            bestNode := ""
            bestAvg := latencyFallbackVal
            for _, node := range nodes.Items {
                avg := nodeAvgLatency(node.Name, latencies)
                klog.Infof("Node %s avg latency: %.3f ms", node.Name, avg)
                if avg < bestAvg {
                    bestAvg = avg
                    bestNode = node.Name
                }
            }
            if bestNode == "" {
                bestNode = nodes.Items[0].Name // fallback
            }

            binding := &corev1.Binding{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      pod.Name,
                    Namespace: pod.Namespace,
                },
                Target: corev1.ObjectReference{
                    Kind: "Node",
                    Name: bestNode,
                },
            }
            err = clientset.CoreV1().Pods(pod.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
            if err != nil {
                klog.ErrorS(err, "Failed to bind pod", "pod", pod.Name, "node", bestNode)
            } else {
                klog.Infof("Pod %s scheduled to node %s", pod.Name, bestNode)
            }
        }
        time.Sleep(2 * time.Second)
    }
}