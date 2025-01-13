package tracker

import (
	"context"
	"ctx.sh/seaway/pkg/apis/seaway.ctx.sh/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sync"
)

type TrackingInfo struct {
	Status    string
	Stage     string
	LastStage string
}

type Tracker struct {
	envs map[types.NamespacedName]*TrackingInfo

	sync.Mutex
}

func New() *Tracker {
	return &Tracker{
		envs: make(map[types.NamespacedName]*TrackingInfo),
	}
}

func (t *Tracker) Track(ctx context.Context, env *v1beta1.Environment) {
	t.Lock()
	defer t.Unlock()

	logger := ctrl.LoggerFrom(ctx, "component", "tracker")

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

		logger.V(6).Info("created tracker", "tracker", t.envs[nn])
	}

	t.envs[nn].Status = env.GetStatusString()
	t.envs[nn].LastStage = t.envs[nn].Stage
	t.envs[nn].Stage = env.GetStageString()

	logger.V(6).Info("updated tracker", "tracker", t.envs[nn])
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
