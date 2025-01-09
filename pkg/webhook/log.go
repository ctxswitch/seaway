package webhook

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

var logger = ctrl.Log.WithName("webhook") //nolint:gochecknoglobals
