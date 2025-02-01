package seaway

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	seawayv1beta1 "ctx.sh/seaway/pkg/gen/seaway/v1beta1"
	"ctx.sh/seaway/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (s *Service) Upload(ctx context.Context, stream *connect.ClientStream[seawayv1beta1.UploadRequest]) (*connect.Response[seawayv1beta1.UploadResponse], error) {
	// TODO: semephore to ensure we aren't uploading the same artifact at the same time.

	// TODO: I think this has issues in a distributed environment.  We might end up
	// 	with a situation where we are interacting with a request that we didn't
	// 	actually start handling initially.  The right solution is to probably fail
	// 	here and make the client re-send the request.

	// TODO: require info for the first request.  Then stream the chunks.
	//   This allows validation and we use the put to block.
	logger := log.FromContext(ctx)
	logger.V(4).Info("Received a file upload request")

	store := NewStore(&StoreOptions{
		Region:   s.options.StorageRegion,
		Bucket:   s.options.StorageBucket,
		Endpoint: s.options.StorageURL,
	})

	err := store.EnsureBucket(ctx, s.options.StorageBucket)
	if err != nil {
		logger.Error(err, "failed to ensure bucket exists")
		werr := errors.New("unable to find or create bucket")
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.Join(werr, err))
	}

	for {
		if more := stream.Receive(); !more {
			store.Close()
			break
		}

		switch payload := stream.Msg().GetPayload().(type) {
		case *seawayv1beta1.UploadRequest_ArtifactInfo:
			name := payload.ArtifactInfo.Name
			namespace := payload.ArtifactInfo.Namespace
			key := util.ArchiveKey(s.options.StoragePrefix, name, namespace)
			// Start the streaming put operation.
			go store.Put(ctx, key)
		case *seawayv1beta1.UploadRequest_Chunk:
			logger.V(6).Info("received chunk", "size", len(payload.Chunk))
			err := store.Write(payload.Chunk)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("expected info or chunk, got %T", payload))
		}
	}
	if err := stream.Err(); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	logger.V(4).Info("waiting for upload to complete")
	store.Wait()

	if err := store.Err(); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	info := store.Info()
	logger.Info("file uploaded", "key", info.Key, "size", info.Size)
	return connect.NewResponse(&seawayv1beta1.UploadResponse{
		Key:     info.Key,
		Size:    info.Size,
		Etag:    info.ETag,
		Message: "ok",
	}), nil
}
