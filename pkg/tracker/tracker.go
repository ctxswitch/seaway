package tracker

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	Tracking      = "Tracking"
	Transitioning = "Transitioning"
)

type TrackingInfo struct {
	Status    string
	Stage     string
	LastStage string
}

type Tracker struct {
	envs     map[types.NamespacedName]*TrackingInfo
	recorder record.EventRecorder
	sync.Mutex
}

func New(recorder record.EventRecorder) *Tracker {
	return &Tracker{
		envs:     make(map[types.NamespacedName]*TrackingInfo),
		recorder: recorder,
	}
}

func (t *Tracker) Track(ctx context.Context, env *v1beta1.Environment) {
	t.Lock()
	defer t.Unlock()

	nn := types.NamespacedName{
		Namespace: env.Namespace,
		Name:      env.Name,
	}
	if _, ok := t.envs[nn]; !ok {
		t.envs[nn] = &TrackingInfo{
			Status:    env.GetStatusString(),
			Stage:     env.GetStageString(),
			LastStage: env.GetStageString(),
		}

		t.recorder.Event(env, corev1.EventTypeNormal, Tracking, t.envs[nn].Stage)
	}

	t.envs[nn].Status = env.GetStatusString()
	t.envs[nn].LastStage = t.envs[nn].Stage
	t.envs[nn].Stage = env.GetStageString()

	if t.envs[nn].Stage != t.envs[nn].LastStage {
		t.recorder.Event(env, corev1.EventTypeNormal, Transitioning, t.envs[nn].Stage)
	}
}

func (t *Tracker) Get(namespace, name string) (TrackingInfo, bool) {
	t.Lock()
	defer t.Unlock()

	if info, ok := t.envs[types.NamespacedName{Namespace: namespace, Name: name}]; ok {
		return *info, true
	}

	return TrackingInfo{}, false
}

func (t *Tracker) HasChanged(namespace, name string) bool {
	t.Lock()
	defer t.Unlock()

	// Do I need to return a notfound error?
	info, ok := t.envs[types.NamespacedName{Namespace: namespace, Name: name}]
	if !ok {
		return false
	}

	return info.Stage != info.LastStage
}

func (t *Tracker) IsDeployed(namespace, name string) bool {
	t.Lock()
	defer t.Unlock()

	info, ok := t.envs[types.NamespacedName{Namespace: namespace, Name: name}]
	if !ok {
		return false
	}

	return info.Status == "deployed"
}
