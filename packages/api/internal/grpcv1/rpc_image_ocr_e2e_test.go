package grpcv1

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/smysnk/sikuligo/internal/core"
	pb "github.com/smysnk/sikuligo/internal/grpcv1/pb"
	"github.com/smysnk/sikuligo/pkg/sikuli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type e2eOCRBackend struct {
	result core.OCRResult
	err    error
}

func (b *e2eOCRBackend) Read(req core.OCRRequest) (core.OCRResult, error) {
	if err := req.Validate(); err != nil {
		return core.OCRResult{}, err
	}
	if b.err != nil {
		return core.OCRResult{}, b.err
	}
	return b.result, nil
}

func TestRPCImageAndOCRE2EViaBufconn(t *testing.T) {
	t.Helper()

	screen := sikuliImageFromRows(t, "screen", [][]uint8{
		{10, 10, 10, 10, 10, 10, 10, 10},
		{10, 10, 10, 0, 255, 10, 10, 10},
		{10, 10, 10, 255, 0, 10, 10, 10},
		{10, 10, 10, 10, 10, 10, 10, 10},
	})

	ocr := &e2eOCRBackend{result: core.OCRResult{
		Text: "Sikuli OCR Ready",
		Words: []core.OCRWord{
			{Text: "Sikuli", X: 1, Y: 0, W: 5, H: 1, Confidence: 0.99},
			{Text: "OCR", X: 7, Y: 0, W: 3, H: 1, Confidence: 0.98},
			{Text: "Ready", X: 11, Y: 0, W: 5, H: 1, Confidence: 0.97},
		},
	}}

	svc := NewServer(
		WithCaptureScreen(func(_ context.Context, _ string) (*sikuli.Image, error) {
			return screen, nil
		}),
		WithFinderFactory(func(source *sikuli.Image) (*sikuli.Finder, error) {
			f, err := sikuli.NewFinder(source)
			if err != nil {
				return nil, err
			}
			f.SetOCRBackend(ocr)
			return f, nil
		}),
	)

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(UnaryInterceptors("", nil, NewMetricsRegistry(), nil)...),
	)
	pb.RegisterSikuliServiceServer(srv, svc)

	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := pb.NewSikuliServiceClient(conn)

	findRes, err := client.FindOnScreen(context.Background(), &pb.FindOnScreenRequest{
		Pattern: &pb.Pattern{
			Image: grayImage("needle", [][]uint8{
				{0, 255},
				{255, 0},
			}),
			Exact: boolPtr(true),
		},
	})
	if err != nil {
		t.Fatalf("find_on_screen failed: %v", err)
	}
	if findRes.GetMatch() == nil {
		t.Fatalf("expected find_on_screen match")
	}
	if gotX, gotY := findRes.GetMatch().GetRect().GetX(), findRes.GetMatch().GetRect().GetY(); gotX != 3 || gotY != 1 {
		t.Fatalf("unexpected find_on_screen match location: x=%d y=%d", gotX, gotY)
	}

	source := grayImage("ocr-source", [][]uint8{
		{5, 5, 5, 5},
		{5, 5, 5, 5},
	})
	readRes, err := client.ReadText(context.Background(), &pb.ReadTextRequest{Source: source})
	if err != nil {
		t.Fatalf("read_text failed: %v", err)
	}
	if readRes.GetText() != "Sikuli OCR Ready" {
		t.Fatalf("unexpected read_text output: %q", readRes.GetText())
	}

	findTextRes, err := client.FindText(context.Background(), &pb.FindTextRequest{Source: source, Query: "ocr"})
	if err != nil {
		t.Fatalf("find_text failed: %v", err)
	}
	if len(findTextRes.GetMatches()) != 1 {
		t.Fatalf("expected 1 OCR text match, got=%d", len(findTextRes.GetMatches()))
	}
	if got := findTextRes.GetMatches()[0].GetText(); got != "OCR" {
		t.Fatalf("unexpected OCR match text: %q", got)
	}
}
