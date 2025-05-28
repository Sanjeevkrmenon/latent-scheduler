**Latency-Aware Kubernetes Scheduler** 🌐

A customizable, production-ready solution for network-aware scheduling in Kubernetes clusters.
Automatically places pods on nodes with the lowest real-time network latency—perfect for performance-critical, distributed workloads.



✨ Features





Cluster-wide latency mesh: Continuously measures node-to-node network latency.



Custom scheduler: Schedules only pods marked with schedulerName: my-latency-scheduler, placing them optimally based on network and resource metrics.



Simple deployment and config: End-to-end solution via Helm and static NFS manifests.



Open architecture: Easily tunable, auditable, and extendable.



📦 What’s Deployed





NFS-backed PV and PVC (/NFS dir): For global sharing of latency data.



Latency Probe DaemonSet (daemonset/): Python-based, pings all nodes, writes JSON per node.



Aggregator CronJob (aggregator/): Merges DaemonSet results every minute.



Custom Go Scheduler (scheduler/): Consumes the aggregated mesh, implements intelligent binding.



All RBAC and service accounts required (Helm and YAML).



🚀 Quickstart: Cluster Setup

1. Deploy the NFS Volume

This creates a test NFS server, exposes it, and pre-provisions the required PV and PVC:

kubectl apply -f NFS/nfs-server.yaml
kubectl apply -f NFS/nfs-pv-pvc.yaml

Wait for the nfs-server pod (namespace: nfs-test) to be running and for the PVC to show as "Bound".

2. Install the Scheduler Stack Using Helm

cd helm
helm install latency-aware-scheduler . -n latency-scheduler --create-namespace



🛠️ How It Works





Probing: Every node runs a DaemonSet probe that pings all other nodes, writing latency numbers to the shared NFS volume as NODE_NAME.json.



Aggregation: Every minute, the aggregator merges all node results into a single cluster-latency.json file.



Scheduling: The custom scheduler watches for pending pods with spec.schedulerName: my-latency-scheduler, and binds each to the node with the best (lowest-latency and available resources) per mesh.



Resilience: All data is retained in an NFS-backed volume. RBAC ensures only required permissions are granted.



📋 How to Use the Latency Scheduler

To use the custom scheduler for your workload, add the following to your pod spec:

spec:
  schedulerName: "my-latency-scheduler"

Only these pods will be processed by the latency-aware logic.



⚙️ Configuration and Customization

Container Images

To fully control and secure your deployment, you may wish to build and push your own images:





Probe: sanjeevkrmenon/latency-measure:latest



Aggregator: sanjeevkrmenon/latency-aggregator:latest



Scheduler: sanjeevkrmenon/latent-scheduler:scheduler1

Override these in helm/values.yaml as needed.

NFS Storage

Pre-provisioned via /NFS/nfs-server.yaml and /NFS/nfs-pv-pvc.yaml.





PVC name: latency-pvc



PV name: latency-nfs-pv



Namespace: latency-scheduler



StorageClass: (empty, static)

If running on a different storage provider, edit accordingly for RWX support.

Resource Requests/Limits

Customizable in helm/values.yaml for all components to suit your cluster capacity.

RBAC and Namespaces

Default manifests use reasonable privileges and work cross-namespace. You can tune them further as required.



🛡️ Troubleshooting





PVC Pending: Ensure NFS pod is running and service is reachable (nfs-server in namespace nfs-test, port 2049).



Probe pods crashloop or logs show kubectl not found: Container should have kubectl baked in; rebuild if you changed the Dockerfile base.



Pods not scheduled: Check the custom scheduler logs:

kubectl logs deploy/my-latency-scheduler -n latency-scheduler



Aggregator not merging: Check CronJob pod logs and ensure /latency has all per-node JSONs.



📂 Project Structure

latent-controller/
├── NFS/          # NFS server, PV, and PVC manifests
├── aggregator/   # Aggregator script (Python), Dockerfile, CronJob
├── daemonset/    # Node probe (Python), Dockerfile, DaemonSet, RBAC
├── scheduler/    # Custom scheduler (Go), Dockerfile, Deployment, RBAC
└── helm/         # Helm chart for easier install



🔒 Security & Best Practices

For production:





Build and push your own container images to a private registry.



Use specific image tags, not :latest.



Tune NFS and PVC sizes; consider using a managed NFS if available.



Restrict RBAC further if possible.



📜 License

Specify your license here (MIT/Apache-2.0/etc).

For questions, feature requests, or bug reports:
[GitHub Issues](GitHub Issues)



This project jump-starts latency-aware pod placement. Try it, adapt it to your cluster, and contribute improvements! 🚀
