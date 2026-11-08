# KubeSentry 🛡️

A lightweight, systems-level Kubernetes diagnostic controller written in Go. 

KubeSentry watches Kubernetes workloads, detects unhealthy container states (`CrashLoopBackOff`, `ImagePullBackOff`, `ErrImagePull`), and produces actionable diagnostic guidance for cluster operators.

Built using pure systems engineering principles with `client-go` and Hexagonal Architecture (no heavyweight operator frameworks, no dashboards, no AI).

---

## Architecture & Design Principles

KubeSentry follows **Hexagonal Architecture (Ports and Adapters)**:
- **Domain Logic (`internal/diagnosis`)**: Pure Go functions that classify failure reasons into actionable findings. Completely decoupled from the Kubernetes API and 100% unit-tested without requiring a cluster.
- **Infrastructure Layer (`main.go`)**: Uses `client-go` to connect via `~/.kube/config` (out-of-cluster) or `ServiceAccount` (in-cluster). Translates Kubernetes `corev1.Pod` objects into domain data types.

---

## Features

- **Dual-Mode Operation**:
  - **One-Shot Report**: Inspects a namespace, formats a tabular summary, and lists findings.
  - **Live Watch Mode (`--watch`)**: Uses Kubernetes HTTP chunked streaming (Watch API) to catch real-time state transitions and pod deletions with automatic connection recovery.
- **Dual-Environment Execution**:
  - Runs locally on host machine via `kubeconfig`.
  - Runs in-cluster as an autonomous Pod under **Least-Privilege RBAC**.

---

## Getting Started

### Prerequisites
- Go 1.22+
- Docker
- Local Kubernetes Cluster (Minikube or Kind)

### 1. Run Unit Tests (No cluster required)
```bash
go test -v ./...