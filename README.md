# 🚀 Latency-Aware Kubernetes Scheduler

*A blazing-fast, **network-aware pod placement** system for smarter Kubernetes clusters!*

---

## **🌟 Key Features**

- **Intelligent network-locality scheduling** — Pods are placed where latency is lowest.
- **Live, cluster-wide latency mesh** — Continuously updated using a Python probe on every node.
- **Production-ready, simple deployment** — All resources, RBAC, and config delivered via Helm.
- **Open & hackable** — Easy to extend, secure, and tune for any team or platform!

---

## **🤔 What Does It Deploy?**

- **NFS**: Shared PersistentVolume for real-time latency data storage.
- **DaemonSet**: Probe pods on every node, measuring true, current network latency.
- **Aggregator CronJob**: Smart combiner for probe data into a unified mesh.
- **Custom Go Scheduler**: Binds pods with `schedulerName: my-latency-scheduler` to the lowest-latency node.
- **RBAC and ServiceAccounts**: Secure, principle-of-least-privilege by default!

---

## **⚡️ Quick Start**

### 1️⃣  Set Up (Test) NFS Storage

```sh
kubectl apply -f NFS/nfs-server.yaml
kubectl apply -f NFS/nfs-pv-pvc.yaml
```

> **Wait for the `nfs-server` pod in `nfs-test` namespace to be `Running` and your PVC to be `Bound`!**

---

### 2️⃣ Deploy Everything with Helm

```sh
cd helm
helm install latency-aware-scheduler . -n latency-scheduler --create-namespace
```

---

## **🔍 How It Works — At a Glance**

1. **Probing:**  
   Each node’s DaemonSet pod measures real-time RTT to all other nodes, saving results as `${NODE_NAME}.json` in the shared NFS PVC.
2. **Merging:**  
   The Aggregator CronJob merges these per-node JSONs into a single `cluster-latency.json` mesh file (default: every minute).
3. **Scheduling:**  
   The custom Go scheduler binds any pod with:

   ```yaml
   spec:
     schedulerName: "my-latency-scheduler"
   ```
   ...to the node with **lowest latency and available resources**.

---

## **🛠️ User Guide**

- **To use latency-aware scheduling, add this to your Pod/Job/Deployment spec:**
    ```yaml
    spec:
      schedulerName: "my-latency-scheduler"
    ```

- **Demo images for quick trial:**
    - Probe: `sanjeevkrmenon/latency-measure:latest`
    - Aggregator: `sanjeevkrmenon/latency-aggregator:latest`
    - Scheduler: `sanjeevkrmenon/latent-scheduler:scheduler1`

> ⚠️ **Production:**  
> Build and use your own images. Pin specific image tags – never use `:latest` in production.

---

## **⚙️ Configuration & Customization**

All key options are stored in [`helm/values.yaml`](helm/values.yaml):

| **Component** | **Setting(s)**                                    | **Purpose**                             |
|---------------|---------------------------------------------------|-----------------------------------------|
| Images        | `daemonset.image`, `aggregator.image`, `scheduler.image` | Use private registry or custom images   |
| PVC & NFS     | `latencyPVC.name`, `storageClassName`, `size`     | Use your RWX NFS/PVC config             |
| Resources     | `daemonset.resources`, `aggregator.resources`, `scheduler.resources` | Tune CPU and memory          |
| Cron Schedule | `aggregator.schedule`                             | Default: every 1 min                    |
| Namespace     | `namespace`                                       | Default: `latency-scheduler`            |

**Tip:** Adjust these for your infrastructure, security, and scale needs!

---

## **🏗️ Project Structure**

```
latent-controller/
├── NFS/          # NFS server, PV, PVC manifests
├── aggregator/   # Aggregator script, Dockerfile, CronJob
├── daemonset/    # Node probe, Dockerfile, DaemonSet, RBAC
├── scheduler/    # Custom scheduler, Dockerfile, Deployment, RBAC
└── helm/         # Helm chart for single-command install
```

---

## **🛡️ Security & Best Practices**

- Use your own signed, trusted images with fixed tags for production.
- Tune RBAC per your organization’s policies.
- Make your NFS storage highly available and appropriately sized.
- Monitor logs for all pods (`kubectl logs ...`) and set up alerts for failures.

---

## **🐛 Troubleshooting**

- **PVC stuck in Pending?** NFS pod may be down, or PV/PVC config may not match.  
  Check with:  
  `kubectl get pv,pvc -A`

- **Probe pod crashes?** Ensure image includes `kubectl` and `ping`.

- **Pods not scheduled?** Check `schedulerName: my-latency-scheduler` is present, and review scheduler logs:  
  `kubectl logs deploy/my-latency-scheduler -n latency-scheduler`

- **No latency mesh or data missing?** Confirm all probe pods are `Running` and aggregator CronJob is running.

---


---

## **🤝 Community & Support**

- **Questions & Issues?**  
  [Open an issue on GitHub!](https://github.com/Sanjeevkrmenon/latent-scheduler/issues)
- **Want to contribute?**  
  PRs and suggestions welcome!

---
