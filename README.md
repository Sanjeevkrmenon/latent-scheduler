# 📡 Latency-Based Kubernetes Scheduler

A custom Kubernetes scheduler that intelligently places pods on nodes based on real-time network latency. This is designed for **latency-sensitive**, **real-time**, and **edge computing** workloads, where pod placement can significantly impact performance.

---

## 🚀 Features

- 📶 **Latency-Aware Scheduling**  
  Prioritizes nodes with the lowest measured latency during pod placement.

- 🔁 **Dynamic Updates**  
  Periodically measures node-to-node latency to make up-to-date scheduling decisions.

- ⚙️ **Plug-and-Play Design**  
  Can run alongside the default Kubernetes scheduler.

- 🔒 **RBAC Support**  
  Minimal and secure permissions using custom ServiceAccounts and ClusterRoles.
