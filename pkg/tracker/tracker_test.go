package tracker

import (
	"context"
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestTracker_Track(t *testing.T) {
	tracker := New()
	assert.Empty(t, tracker.envs)

	// Track new environment
	env := &v1beta1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Status: v1beta1.EnvironmentStatus{
			Stage:            v1beta1.EnvironmentStageInitialize,
			DeployedRevision: "fake",
		},
	}

	tracker.Track(context.TODO(), env)
	assert.Len(t, tracker.envs, 1)
	assert.Equal(t, &TrackingInfo{
		Stage:     "Initializing",
		LastStage: "Initializing",
		Status:    "initializing",
	}, tracker.envs[types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}])

	// Track an environment update
	env = &v1beta1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Status: v1beta1.EnvironmentStatus{
			Stage:            v1beta1.EnvironmentStageDeploy,
			DeployedRevision: "fake",
		},
	}

	tracker.Track(context.TODO(), env)
	assert.Len(t, tracker.envs, 1)
	assert.Equal(t, &TrackingInfo{
		Stage:     "Deploying the revision",
		LastStage: "Initializing",
		Status:    "deploying",
	}, tracker.envs[types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}])

	// Track updated environment
	env = &v1beta1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1beta1.EnvironmentSpec{
			Revision: "fake",
		},
		Status: v1beta1.EnvironmentStatus{
			Stage:            v1beta1.EnvironmentStageDeployed,
			DeployedRevision: "fake",
		},
	}

	tracker.Track(context.TODO(), env)
	assert.Len(t, tracker.envs, 1)
	assert.Equal(t, &TrackingInfo{
		Stage:     "Revision deployed",
		LastStage: "Deploying the revision",
		Status:    "deployed",
	}, tracker.envs[types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}])

	env = &v1beta1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "another-test",
			Namespace: "default",
		},
		Status: v1beta1.EnvironmentStatus{
			Stage:            v1beta1.EnvironmentStageInitialize,
			DeployedRevision: "fake",
		},
	}

	tracker.Track(context.TODO(), env)
	assert.Len(t, tracker.envs, 2)
	assert.Equal(t, &TrackingInfo{
		Stage:     "Initializing",
		LastStage: "Initializing",
		Status:    "initializing",
	}, tracker.envs[types.NamespacedName{
		Name:      "another-test",
		Namespace: "default",
	}])
	assert.Equal(t, &TrackingInfo{
		Stage:     "Revision deployed",
		LastStage: "Deploying the revision",
		Status:    "deployed",
	}, tracker.envs[types.NamespacedName{
		Name:      "test",
		Namespace: "default",
	}])
}
