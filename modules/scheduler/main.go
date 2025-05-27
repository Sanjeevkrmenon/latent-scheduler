package main

import (
    "context"
    "encoding/json"
    "io/ioutil"
    "math"
    "time"

    "k8s.io/apimachinery/pkg/api/resource"
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

// For parsing JSON latencies file
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

// resourceFitsOnNode returns true if the pod's resource requests go below or equal the node's remaining allocatable resources
func resourceFitsOnNode(pod *corev1.Pod, node *corev1.Node, podsOnNode []corev1.Pod) bool {
    // Get allocatable from node
    allocatable := node.Status.Allocatable

    totalCPU := resource.MustParse("0")
    totalMem := resource.MustParse("0")
    for _, p := range podsOnNode {
        // Only count running and pending pods (not succeeded/failed)
        phase := p.Status.Phase
        if phase == corev1.PodSucceeded || phase == corev1.PodFailed {
            continue
        }
        for _, c := range p.Spec.Containers {
            req := c.Resources.Requests
            totalCPU.Add(req[corev1.ResourceCPU])
            totalMem.Add(req[corev1.ResourceMemory])
        }
    }

    // Compute requested by new pod
    newCPU := resource.MustParse("0")
    newMem := resource.MustParse("0")
    for _, c := range pod.Spec.Containers {
        req := c.Resources.Requests
        newCPU.Add(req[corev1.ResourceCPU])
        newMem.Add(req[corev1.ResourceMemory])
    }

    remainCPU := allocatable[corev1.ResourceCPU]
    remainCPU.Sub(totalCPU)
    remainCPU.Sub(newCPU)
    remainMem := allocatable[corev1.ResourceMemory]
    remainMem.Sub(totalMem)
    remainMem.Sub(newMem)

    // Both must be >= 0
    return remainCPU.Sign() >= 0 && remainMem.Sign() >= 0
}

func podsAssignedToNode(nodeName string, allPods []corev1.Pod) (ret []corev1.Pod) {
    for _, p := range allPods {
        if p.Spec.NodeName == nodeName {
            ret = append(ret, p)
        }
    }
    return
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

        nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
        if err != nil || len(nodes.Items) == 0 {
            klog.ErrorS(err, "Failed to list nodes")
            time.Sleep(2 * time.Second)
            continue
        }

        // Cache of all pods for assignment checks
        allPodsList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
        if err != nil {
            klog.ErrorS(err, "Failed to list all pods")
            time.Sleep(2 * time.Second)
            continue
        }

        for _, pod := range pods.Items {
            if pod.Spec.SchedulerName != schedulerName || pod.Spec.NodeName != "" {
                continue
            }

            // Find feasible nodes
            feasibleNodes := []corev1.Node{}
            for _, node := range nodes.Items {
                podsOnNode := podsAssignedToNode(node.Name, allPodsList.Items)
                if resourceFitsOnNode(&pod, &node, podsOnNode) {
                    feasibleNodes = append(feasibleNodes, node)
                }
            }

            if len(feasibleNodes) == 0 {
                klog.Warningf("No feasible nodes (resource fit) found for pod %s/%s", pod.Namespace, pod.Name)
                continue
            }

            // Among feasible nodes, pick with lowest average latency
            bestNode := ""
            bestAvg := latencyFallbackVal

            for _, node := range feasibleNodes {
                avg := nodeAvgLatency(node.Name, latencies)
                klog.Infof("Node %s avg latency: %.3f ms", node.Name, avg)
                if avg < bestAvg {
                    bestAvg = avg
                    bestNode = node.Name
                }
            }

            // Fallback: pick first feasible node if no latency map or all maxed out
            if bestNode == "" {
                bestNode = feasibleNodes[0].Name
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
                klog.Infof("Pod %s scheduled to node %s (latency %.3f ms)", pod.Name, bestNode, bestAvg)
            }
        }
        time.Sleep(2 * time.Second)
    }
}