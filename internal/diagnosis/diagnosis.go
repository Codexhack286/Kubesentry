package diagnosis

// Finding represents an actionable diagnostic report for a Pod.
type Finding struct {
	PodName      string
	Namespace    string
	Severity     string
	Reason       string
	Message      string
	RestartCount int32
}

// Classify takes raw status strings and returns a Finding if the Pod is unhealthy.
// If the Pod is healthy or the state is unrecognized, it returns nil.
func Classify(namespace string, podName string, waitingReason string, restartCount int32) *Finding {
	// If there is no waiting reason, we assume it's healthy (or at least not failing here).
	if waitingReason == "" {
		return nil
	}

	// Base finding with the metadata we already know
	finding := &Finding{
		Namespace:    namespace,
		PodName:      podName,
		Reason:       waitingReason,
		RestartCount: restartCount,
		Severity:     "WARNING", // Default severity for failures
	}

	// Classify the specific failure
	switch waitingReason {
	case "CrashLoopBackOff":
		finding.Message = "inspect previous container logs and termination state"
		return finding
	case "ImagePullBackOff", "ErrImagePull":
		finding.Message = "inspect image name, registry access and pull secrets"
		return finding
	default:
		// For unrecognized reasons, we return nil for now as per v0.1 specs.
		return nil
	}
}
