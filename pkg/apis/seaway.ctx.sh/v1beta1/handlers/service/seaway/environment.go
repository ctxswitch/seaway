package seaway

import (
	"connectrpc.com/connect"
	"context"
	seawayv1beta1 "ctx.sh/seaway/pkg/gen/seaway/v1beta1"
	seawayv1beta1connect "ctx.sh/seaway/pkg/gen/seaway/v1beta1/seawayv1beta1connect"
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// TODO: Now that we are looking at streaming status, I can consolidate some of the
// 	reconciliation stages.

// Environment returns the status of an environment.
func (s *Service) Environment(ctx context.Context, req *connect.Request[seawayv1beta1.EnvironmentRequest]) (*connect.Response[seawayv1beta1.EnvironmentResponse], error) {
	logger := log.FromContext(ctx,
		"service", seawayv1beta1connect.SeawayServiceName,
		"path", seawayv1beta1connect.SeawayServiceEnvironmentProcedure,
	)
	logger.V(4).Info("received request", "request", req.Msg)

	track := s.options.Tracker
	// Return the last environment status.  If it doesn't exist return notfound.
	info, ok := track.Get(req.Msg.Namespace, req.Msg.Name)
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("resource not found"))
	}

	return connect.NewResponse(&seawayv1beta1.EnvironmentResponse{
		Stage:  info.Stage,
		Status: info.Status,
	}), nil
}

// EnvironmentTracker streams changes in the environment deployment status back to a
// client.
func (s *Service) EnvironmentTracker(ctx context.Context, req *connect.Request[seawayv1beta1.EnvironmentRequest], stream *connect.ServerStream[seawayv1beta1.EnvironmentResponse]) error {
	logger := log.FromContext(ctx,
		"service", seawayv1beta1connect.SeawayServiceName,
		"path", seawayv1beta1connect.SeawayServiceEnvironmentTrackerProcedure,
	)

	stopCh := make(chan struct{})
	// TODO: This is a temporary solution.  We need to subscribe to a channel
	// 	or some sort of queue and send only when a new event comes in.  Currently
	// 	we may miss changes in the stages that happen in less time than the interval.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Loop until done.
	for {
		select {
		case <-ctx.Done():
			// Context has been cancelled. We're shutting down. It will be the client's
			// responsibility to reconnect.
			logger.V(1).Info("Context cancelled")
			return connect.NewError(connect.CodeCanceled, errors.New("shutting down"))
		case <-stopCh:
			logger.V(1).Info("Request finished")
			return nil
		case <-ticker.C:
			logger.V(6).Info("Sending")
			err := s.send(ctx, stopCh, req, stream)
			if err != nil {
				return connect.NewError(connect.CodeInternal, err)
			}
		}
	}
}

func (s *Service) send(
	ctx context.Context,
	stopCh chan struct{},
	req *connect.Request[seawayv1beta1.EnvironmentRequest],
	stream *connect.ServerStream[seawayv1beta1.EnvironmentResponse],
) error {
	logger := log.FromContext(ctx, "req", req)
	track := s.options.Tracker
	name := req.Msg.Name
	namespace := req.Msg.Namespace

	info, ok := track.Get(namespace, name)
	if !ok {
		// If we can't find this, we probably haven't reconciled it yet.  This assumption
		// is partially valid.  May want to expand later.
		return nil
	}
	logger.V(6).Info("info", "info", info)

	changed := track.HasChanged(namespace, name)
	deployed := track.IsDeployed(namespace, name)

	// TODO: clean up the initializing logic here.  it's not very intuitive what
	//   this is doing and why it is needed.  For reference there was a state at the
	//   beginning of a deployment stream where we would duplicate sending a status
	//   update when in initializing.
	if (changed || deployed) && info.Status != "initializing" {
		logger.V(6).Info("sending", "info", info)
		err := stream.Send(&seawayv1beta1.EnvironmentResponse{
			Stage:  info.Stage,
			Status: info.Status,
		})
		if err != nil {
			return err
		}
	}

	if info.Status == "deployed" || info.Status == "failed" {
		logger.V(6).Info("stopping", "info", info)
		close(stopCh)
	}

	return nil
}
