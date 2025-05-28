# Latency-Aware Kubernetes Scheduler Helm Chart

## Overview

This Helm chart deploys a **network-latency-aware scheduling system** for Kubernetes clusters. 
It allows pods to be scheduled on nodes with the lowest network latency, using real-time latency measurements collected between all nodes.

> **Key Components Deployed:**
> - A DaemonSet of latency probe agents (per node)
> - An aggregator CronJob that summarizes/provides latency mesh data
> - A custom scheduler Deployment that places pods based on measured network latency and resource availability
> - A shared PersistentVolumeClaim for data storage
> - All required RBAC resources

---

## How it Works

1. **DaemonSet Probe**: Each node runs a probe that measures network latency to all other nodes and writes results to a shared volume.
2. **Aggregator CronJob**: Periodically parses/aggregates probe data into a cluster-wide latency mesh (often JSON).
3. **Custom Scheduler**: Watches for pending pods, reads the aggregated latency mesh, and binds pods to the node with the best (lowest-latency and resource-available) placement.
4. **Shared Volume**: All components read/write results using a shared, RWX-accessible PersistentVolumeClaim (usually NFS).

---

## When and Why Use This?

- Your workloads are sensitive to network latency (e.g., distributed databases, machine learning, streaming).
- You want **network topology awareness** in pod scheduling—something default Kubernetes scheduling lacks.
- You need **observability** into actual, dynamic, intra-cluster network performance.

---

## Prerequisites

- **Kubernetes** v1.21+ (tested on 1.24+)
- **Helm** v3+
- A **StorageClass** that supports `ReadWriteMany` (RWX) PVCs (e.g., NFS, CephFS, EFS, or similar)
- Ability to create cluster-level RBAC

---

## Quickstart Installation

1. **Clone the repository**:
    ```bash
    git clone https://github.com/YOURORG/latency-aware-scheduler
    cd latency-aware-scheduler/helmting/latency-aware-scheduler
    ```

2. **Customize values.yaml** (see [_Changes Required by Adopters_](#changes-required-by-adopters) below for the most important settings!).

3. **Install via Helm** (create the namespace if needed):
    ```bash
    helm install my-latency-scheduler . -n latency-scheduler --create-namespace
    ```

---

## Changes Required by Adopters

> ⚠️ **Most teams will need to change several settings before use!**

### 1. **Container Image References**
- Change images for DaemonSet, Aggregator, and Scheduler to your own (built and hosted in your registry), especially for PROD use and security.
    ```yaml
    daemonset:
      image: yourregistry/latency-measure:yourtag
    aggregator:
      image: yourregistry/latency-aggregator:yourtag
    scheduler:
      image: yourregistry/latent-scheduler:yourtag
    ```

### 2. **StorageClass Name & PVC Size**
- Set `.Values.latencyPVC.storageClassName` to a **working `ReadWriteMany` StorageClass on your cluster**.
    ```yaml
    latencyPVC:
      storageClassName: your-storage-class   # e.g., "nfs"
      size: 2Gi                             # Increase if needed
    ```

### 3. **Resource Requests/Limits**
- Adapt requests/limits in values.yaml for your own hardware, to avoid eviction or resource contention.

### 4. **Namespace**
- Default is `latency-scheduler`. Change in `values.yaml` as needed and ensure it's consistent.

### 5. **Advanced Scheduling (Optional)**
- Set `nodeSelector`, `tolerations`, or `affinity` in values.yaml to constrain probes/scheduler/aggregator to specific nodes/pools.

---

## Configuration (values.yaml Key Fields)

Below is a sample `values.yaml`—**edit these before installing**:

```yaml
namespace: latency-scheduler

latencyPVC:
  name: latency-pvc
  storageClassName: "nfs"
  size: 2Gi

daemonset:
  image: yourrepo/latency-measure:1.0.0
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi
  nodeSelector: {}
  tolerations: []
  affinity: {}

aggregator:
  image: yourrepo/latency-aggregator:1.0.0
  schedule: "*/2 * * * *"
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 100m
      memory: 128Mi

scheduler:
  image: yourrepo/latent-scheduler:1.0.0
  replicas: 1
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi
