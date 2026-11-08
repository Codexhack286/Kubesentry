package diagnosis

import (
	"testing"
)

func TestClassify(t *testing.T) {
	// 1. Define our test cases (The "Table")
	tests := []struct {
		name          string
		namespace     string
		podName       string
		waitingReason string
		restartCount  int32
		wantNil       bool
		wantReason    string
	}{
		{
			name:          "Healthy Pod returns nil",
			namespace:     "demo",
			podName:       "api-healthy",
			waitingReason: "",
			restartCount:  0,
			wantNil:       true,
		},
		{
			name:          "CrashLoopBackOff produces finding",
			namespace:     "demo",
			podName:       "worker-crash",
			waitingReason: "CrashLoopBackOff",
			restartCount:  5,
			wantNil:       false,
			wantReason:    "CrashLoopBackOff",
		},
		{
			name:          "ImagePullBackOff produces finding",
			namespace:     "demo",
			podName:       "web-pull-fail",
			waitingReason: "ImagePullBackOff",
			restartCount:  0,
			wantNil:       false,
			wantReason:    "ImagePullBackOff",
		},
		{
			name:          "Unknown reason returns nil",
			namespace:     "demo",
			podName:       "unknown-pod",
			waitingReason: "SomeRandomReason",
			restartCount:  1,
			wantNil:       true,
		},
	}

	// 2. Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.namespace, tt.podName, tt.waitingReason, tt.restartCount)

			if tt.wantNil && got != nil {
				t.Errorf("expected nil finding, but got %v", got)
			}

			if !tt.wantNil {
				if got == nil {
					t.Fatalf("expected a finding, but got nil")
				}
				if got.Reason != tt.wantReason {
					t.Errorf("expected reason %s, got %s", tt.wantReason, got.Reason)
				}
				if got.Namespace != tt.namespace {
					t.Errorf("expected namespace %s, got %s", tt.namespace, got.Namespace)
				}
				if got.PodName != tt.podName {
					t.Errorf("expected podName %s, got %s", tt.podName, got.PodName)
				}
				if got.RestartCount != tt.restartCount {
					t.Errorf("expected restartCount %d, got %d", tt.restartCount, got.RestartCount)
				}
			}
		})
	}
}
