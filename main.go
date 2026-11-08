package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"

	// Replace with your module path from go.mod if different
	"github.com/Codexhack286/kubesentry/internal/diagnosis"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Parse CLI flags
	watchMode := flag.Bool("watch", false, "Continuously watch for pod events")
	namespaceFlag := flag.String("namespace", "demo", "Kubernetes namespace to monitor")
	flag.Parse()

	var config *rest.Config
	var err error

	// 1. Try In-Cluster Config (when running as a Pod)
	config, err = rest.InClusterConfig()
	if err != nil {
		// 2. Fallback to Out-of-Cluster ~/.kube/config (when running on laptop)
		homeDir, _ := os.UserHomeDir()
		kubeconfig := filepath.Join(homeDir, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to load kubeconfig: %v", err))
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Graceful shutdown handling (Ctrl+C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *watchMode {
		runWatchMode(ctx, clientset, *namespaceFlag)
	} else {
		runOneShotReport(ctx, clientset, *namespaceFlag)
	}
}

// runOneShotReport prints the static table we built in Phase 2
func runOneShotReport(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	var findings []*diagnosis.Finding
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tPOD\tSTATUS\tRESTARTS")

	for _, pod := range pods.Items {
		status, restarts, finding := analyzePod(&pod)
		if finding != nil {
			findings = append(findings, finding)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", pod.Namespace, pod.Name, status, restarts)
	}
	w.Flush()

	if len(findings) > 0 {
		fmt.Printf("\nFindings:\n")
		for _, f := range findings {
			fmt.Printf("[%s] %s\n", f.Severity, f.PodName)
			fmt.Printf("  Reason: %s\n", f.Reason)
			fmt.Printf("  Next:   %s\n\n", f.Message)
		}
	} else {
		fmt.Println("\nAll workloads in this namespace are healthy!")
	}
}

// runWatchMode streams events in real-time with reconnection logic
func runWatchMode(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	fmt.Printf("[KubeSentry] Live watch started on namespace: '%s' (Press Ctrl+C to stop)...\n\n", namespace)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n[KubeSentry] Stopping watcher...")
			return
		default:
			// Start watching
			watcher, err := clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("[ERROR] Failed to start watch: %v. Retrying in 5s...\n", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Stream events from ResultChan
			for event := range watcher.ResultChan() {
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					continue // Ignore non-pod objects
				}

				timestamp := time.Now().Format("15:04:05")
				_, restarts, finding := analyzePod(pod)

				if finding != nil {
					fmt.Printf("[%s]  ALERT: Pod '%s' entered '%s' (Restarts: %d)\n",
						timestamp, pod.Name, finding.Reason, restarts)
					fmt.Printf("           Diagnosis: %s\n\n", finding.Message)
				} else if event.Type == "DELETED" {
					fmt.Printf("[%s] ℹ️  INFO: Pod '%s' was deleted\n\n", timestamp, pod.Name)
				}
			}

			// If the loop exits, the stream was closed by the server. We reconnect automatically!
			fmt.Println("[KubeSentry] Watch stream closed by API server. Reconnecting...")
		}
	}
}

// Helper to inspect container statuses and run the classifier
func analyzePod(pod *corev1.Pod) (string, int32, *diagnosis.Finding) {
	status := "HEALTHY"
	var totalRestarts int32 = 0
	var finalFinding *diagnosis.Finding

	for _, cs := range pod.Status.ContainerStatuses {
		totalRestarts += cs.RestartCount

		if cs.State.Waiting != nil {
			f := diagnosis.Classify(pod.Namespace, pod.Name, cs.State.Waiting.Reason, cs.RestartCount)
			if f != nil {
				status = f.Reason
				finalFinding = f
			}
		}
	}
	return status, totalRestarts, finalFinding
}
