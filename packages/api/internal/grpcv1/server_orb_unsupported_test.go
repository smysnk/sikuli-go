//go:build !opencv

package grpcv1

import (
	"context"
	"testing"

	pb "github.com/smysnk/sikuligo/internal/grpcv1/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFindFeatureMatchersWithoutOpenCVMapToUnimplemented(t *testing.T) {
	srv := NewServer()
	featureEngines := []pb.MatcherEngine{
		pb.MatcherEngine_MATCHER_ENGINE_ORB,
		pb.MatcherEngine_MATCHER_ENGINE_AKAZE,
		pb.MatcherEngine_MATCHER_ENGINE_BRISK,
		pb.MatcherEngine_MATCHER_ENGINE_KAZE,
		pb.MatcherEngine_MATCHER_ENGINE_SIFT,
	}
	for _, engine := range featureEngines {
		engine := engine
		t.Run(engine.String(), func(t *testing.T) {
			req := &pb.FindRequest{
				Source: grayImage("source", [][]uint8{
					{10, 10, 10, 10},
					{10, 0, 255, 10},
					{10, 255, 0, 10},
					{10, 10, 10, 10},
				}),
				Pattern: &pb.Pattern{
					Image: grayImage("needle", [][]uint8{
						{0, 255},
						{255, 0},
					}),
					Exact: boolPtr(true),
				},
				MatcherEngine: engine,
			}
			_, err := srv.Find(context.Background(), req)
			if err == nil {
				t.Fatalf("expected unimplemented error")
			}
			if code := status.Code(err); code != codes.Unimplemented {
				t.Fatalf("expected unimplemented code, got %s", code)
			}
		})
	}
}
