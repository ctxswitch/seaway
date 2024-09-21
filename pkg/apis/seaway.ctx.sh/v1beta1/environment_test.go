package v1beta1

import "testing"

func TestHasFailed(t *testing.T) {
	var tests = []struct {
		name     string
		stage    EnvironmentStage
		expected bool
	}{
		{"BuildImageFailed", EnvironmentStageBuildImageFailed, true},
		{"DeployFailed", EnvironmentStageDeployFailed, true},
		{"Failed", EnvironmentStageFailed, true},
		{"NotFailed", EnvironmentStageDeploy, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Environment{
				Status: EnvironmentStatus{
					Stage: tt.stage,
				},
			}

			if got := env.HasFailed(); got != tt.expected {
				t.Errorf("HasFailed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasDeviated(t *testing.T) {
	var tests = []struct {
		name             string
		specRevision     string
		expectedRevision string
		expected         bool
	}{
		{"No deviation", "rev1", "rev1", false},
		{"Deviated", "rev1", "rev2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Environment{
				Spec: EnvironmentSpec{
					Revision: tt.specRevision,
				},
				Status: EnvironmentStatus{
					ExpectedRevision: tt.expectedRevision,
				},
			}

			if got := env.HasDeviated(); got != tt.expected {
				t.Errorf("HasDeviated() = %v, want %v", got, tt.expected)
			}
		})
	}
}
