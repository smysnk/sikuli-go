package grpcv1

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"net"
	"testing"
	"time"

	pb "github.com/smysnk/sikuligo/internal/grpcv1/pb"
	"github.com/smysnk/sikuligo/pkg/sikuli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type findBenchEngine struct {
	name string
	enum pb.MatcherEngine
}

type findBenchScenario struct {
	name          string
	variant       string
	size          int
	rotation      int
	screenW       int
	screenH       int
	tolerance     float64
	maxAreaRatio  float64
	transformKind string
	transformA    float64
	transformB    float64
	transformC    float64
	queryFromBase bool
}

func BenchmarkFindOnScreenE2E(b *testing.B) {
	b.ReportAllocs()
	visuals := newFindBenchVisualCollectorFromEnv(b)

	engines := []findBenchEngine{
		{name: "template", enum: pb.MatcherEngine_MATCHER_ENGINE_TEMPLATE},
		{name: "orb", enum: pb.MatcherEngine_MATCHER_ENGINE_ORB},
		{name: "hybrid", enum: pb.MatcherEngine_MATCHER_ENGINE_HYBRID},
	}

	scenarios := []findBenchScenario{
		{name: "grid_small_r0_480x270", variant: "grid", size: 32, rotation: 0, screenW: 480, screenH: 270, tolerance: 0.30, maxAreaRatio: 1.50},
		{name: "grid_medium_r90_640x360", variant: "grid", size: 48, rotation: 90, screenW: 640, screenH: 360, tolerance: 0.30, maxAreaRatio: 1.50},
		{name: "grid_large_r180_800x450", variant: "grid", size: 64, rotation: 180, screenW: 800, screenH: 450, tolerance: 0.30, maxAreaRatio: 1.50},
		{name: "glyph_small_r270_480x270", variant: "glyph", size: 32, rotation: 270, screenW: 480, screenH: 270, tolerance: 0.30, maxAreaRatio: 1.50},
		{name: "glyph_medium_r0_640x360", variant: "glyph", size: 48, rotation: 0, screenW: 640, screenH: 360, tolerance: 0.30, maxAreaRatio: 1.50},
		{name: "glyph_large_r90_800x450", variant: "glyph", size: 64, rotation: 90, screenW: 800, screenH: 450, tolerance: 0.30, maxAreaRatio: 1.50},
		{name: "noise_medium_r180_800x450", variant: "noise", size: 96, rotation: 180, screenW: 800, screenH: 450, tolerance: 0.25, maxAreaRatio: 1.50},
		{name: "noise_large_r270_960x540", variant: "noise", size: 128, rotation: 270, screenW: 960, screenH: 540, tolerance: 0.25, maxAreaRatio: 1.50},
		{name: "orbtex_medium_r0_800x450", variant: "orbtex", size: 96, rotation: 0, screenW: 800, screenH: 450, tolerance: 0.25, maxAreaRatio: 1.50},
		{name: "orbtex_large_r90_960x540", variant: "orbtex", size: 128, rotation: 90, screenW: 960, screenH: 540, tolerance: 0.25, maxAreaRatio: 1.50},
		{name: "orbtex_large_r180_1024x576", variant: "orbtex", size: 128, rotation: 180, screenW: 1024, screenH: 576, tolerance: 0.25, maxAreaRatio: 1.50},
		{name: "orbx_resize_115_960x540", variant: "orbtex", size: 160, rotation: 0, screenW: 960, screenH: 540, tolerance: 0.22, maxAreaRatio: 2.20, transformKind: "scale", transformA: 1.15, queryFromBase: true},
		{name: "orbx_rotate_12deg_960x540", variant: "orbtex", size: 160, rotation: 0, screenW: 960, screenH: 540, tolerance: 0.20, maxAreaRatio: 2.40, transformKind: "rotate", transformA: 12.0, queryFromBase: true},
		{name: "orbx_perspective_960x540", variant: "orbtex", size: 160, rotation: 0, screenW: 960, screenH: 540, tolerance: 0.18, maxAreaRatio: 2.60, transformKind: "perspective", transformA: 0.90, transformB: 1.08, transformC: 0.05, queryFromBase: true},
		{name: "orbx_skewx_010_960x540", variant: "orbtex", size: 160, rotation: 0, screenW: 960, screenH: 540, tolerance: 0.18, maxAreaRatio: 2.60, transformKind: "skewx", transformA: 0.10, queryFromBase: true},
	}

	for _, engine := range engines {
		engine := engine
		b.Run("engine="+engine.name, func(b *testing.B) {
			for _, scenario := range scenarios {
				scenario := scenario
				b.Run(scenario.name, func(b *testing.B) {
					runFindOnScreenScenarioBenchmark(b, engine, scenario, visuals)
				})
			}
		})
	}

	if visuals != nil {
		if err := visuals.WriteScenarioSummaries(); err != nil {
			b.Fatalf("write benchmark visual summaries: %v", err)
		}
	}
}

func runFindOnScreenScenarioBenchmark(b *testing.B, engine findBenchEngine, scenario findBenchScenario, visuals *findBenchVisualCollector) {
	source, patternImage, expectedRect := buildFindBenchFixture(b, scenario)

	svc := NewServer(WithCaptureScreen(func(_ context.Context, _ string) (*sikuli.Image, error) {
		return source, nil
	}))

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(UnaryInterceptors("", nil, NewMetricsRegistry(), nil)...),
	)
	pb.RegisterSikuliServiceServer(srv, svc)

	go func() {
		_ = srv.Serve(lis)
	}()
	b.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(
		dialCtx,
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		b.Fatalf("dial bufconn: %v", err)
	}
	b.Cleanup(func() { _ = conn.Close() })

	client := pb.NewSikuliServiceClient(conn)
	pattern := &pb.Pattern{
		Image:      patternImage,
		Similarity: benchSimilarityPtr(engine.name),
	}
	request := &pb.FindOnScreenRequest{
		Pattern: pattern,
		Opts: &pb.ScreenQueryOptions{
			MatcherEngine: engine.enum,
		},
	}

	if visuals != nil {
		visuals.CaptureAttempts(b, client, request, source.Gray(), expectedRect, engine, scenario)
	}

	_, _ = client.FindOnScreen(context.Background(), request)

	successCount := 0
	notFoundCount := 0
	unsupportedCount := 0
	errorCount := 0
	overlapMissCount := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := client.FindOnScreen(context.Background(), request)
		if err != nil {
			code := status.Code(err)
			switch code {
			case codes.NotFound:
				notFoundCount++
			case codes.Unimplemented:
				unsupportedCount++
			default:
				errorCount++
			}
			continue
		}
		if res.GetMatch() == nil || res.GetMatch().GetRect() == nil {
			errorCount++
			continue
		}
		if !rectMatchSatisfies(res.GetMatch().GetRect(), expectedRect, scenario.tolerance, scenario.maxAreaRatio) {
			overlapMissCount++
			continue
		}
		successCount++
	}
	if b.N > 0 {
		denom := float64(b.N)
		b.ReportMetric(float64(successCount)/denom, "success/op")
		b.ReportMetric(float64(notFoundCount)/denom, "not_found/op")
		b.ReportMetric(float64(unsupportedCount)/denom, "unsupported/op")
		b.ReportMetric(float64(errorCount)/denom, "error/op")
		b.ReportMetric(float64(overlapMissCount)/denom, "overlap_miss/op")
	}
}

func buildFindBenchFixture(t testing.TB, scenario findBenchScenario) (*sikuli.Image, *pb.GrayImage, *pb.Rect) {
	t.Helper()

	basePattern := buildBenchPattern(scenario.variant, scenario.size)
	queryPattern := rotateGrayByQuarterTurns(basePattern, scenario.rotation)
	targetPattern := applyBenchTransform(queryPattern, scenario)
	haystack := buildBenchHaystack(scenario.screenW, scenario.screenH, scenario.variant)

	pbounds := targetPattern.Bounds()
	insertX := (scenario.screenW - pbounds.Dx()) / 3
	insertY := (scenario.screenH - pbounds.Dy()) / 2
	if insertX < 0 || insertY < 0 {
		t.Fatalf("pattern does not fit haystack scenario=%s pattern=%dx%d haystack=%dx%d", scenario.name, pbounds.Dx(), pbounds.Dy(), scenario.screenW, scenario.screenH)
	}
	// Fill the full frame with near-match decoys first so the target is camouflaged.
	populateNearMatchDecoys(haystack, targetPattern, insertX, insertY, scenario)
	// Blend target edges slightly into local context, then use the inserted patch as the query pattern.
	blitGrayFeather(haystack, targetPattern, insertX, insertY, maxInt(2, scenario.size/20))
	applySeamSmoothing(haystack, scenario.variant)

	source, err := sikuli.NewImageFromGray(fmt.Sprintf("bench-%s-source", scenario.name), haystack)
	if err != nil {
		t.Fatalf("new source image: %v", err)
	}

	patternGray := cropGray(haystack, image.Rect(insertX, insertY, insertX+pbounds.Dx(), insertY+pbounds.Dy()))
	if scenario.queryFromBase {
		patternGray = queryPattern
	}
	pattern := grayProtoFromGray(fmt.Sprintf("bench-%s-pattern", scenario.name), patternGray)
	expected := &pb.Rect{X: int32(insertX), Y: int32(insertY), W: int32(pbounds.Dx()), H: int32(pbounds.Dy())}
	return source, pattern, expected
}

func buildBenchPattern(variant string, size int) *image.Gray {
	if size < 16 {
		size = 16
	}
	img := image.NewGray(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Pix[img.PixOffset(x, y)] = 235
		}
	}

	for i := 0; i < size; i++ {
		img.Pix[img.PixOffset(i, i)] = uint8(70 + (i*3)%70)
		img.Pix[img.PixOffset(size-1-i, i)] = uint8(180 - (i*5)%70)
	}

	switch variant {
	case "glyph":
		step := maxInt(6, size/6)
		for y := step; y < size-step; y += step {
			for x := step; x < size-step; x += step {
				if (x+y)/step%2 == 0 {
					setRect(img, x-1, y-1, 3, 3, 20)
				} else {
					setRect(img, x-1, y-1, 3, 3, 245)
				}
			}
		}
		setRect(img, size/4, size/4, size/2, maxInt(3, size/10), 30)
		setRect(img, size/3, size/2, size/3, maxInt(3, size/10), 210)
	case "noise":
		for y := 1; y < size-1; y++ {
			for x := 1; x < size-1; x++ {
				v := (x*73 + y*151 + (x*y)%251 + size*19) & 0xFF
				img.Pix[img.PixOffset(x, y)] = uint8(v)
			}
		}
		setRect(img, size/8, size/8, size/3, size/20+2, 0)
		setRect(img, size/2, size/3, size/4, size/18+2, 255)
		setRect(img, size/3, size/2, size/20+2, size/3, 0)
	case "orbtex":
		for y := 1; y < size-1; y++ {
			for x := 1; x < size-1; x++ {
				v := (x*97 + y*193 + (x*y)%239 + size*41) & 0xFF
				if ((x+y)%7 == 0) || ((x*3+y*5)%11 == 0) {
					v = 255 - v
				}
				img.Pix[img.PixOffset(x, y)] = uint8(v)
			}
		}
		ringStep := maxInt(10, size/8)
		for r := ringStep; r < size-ringStep; r += ringStep {
			setRect(img, r, r, size-r*2, 2, uint8((r*37)&0xFF))
			setRect(img, r, r, 2, size-r*2, uint8((r*59)&0xFF))
			setRect(img, size-r-2, r, 2, size-r*2, uint8((r*83)&0xFF))
			setRect(img, r, size-r-2, size-r*2, 2, uint8((r*97)&0xFF))
		}
		for i := 0; i < size; i += 3 {
			img.Pix[img.PixOffset(i, (i*5)%size)] = 0
			img.Pix[img.PixOffset((i*7)%size, i)] = 255
		}
	default:
		for y := 2; y < size-2; y++ {
			for x := 2; x < size-2; x++ {
				if (x/4+y/4)%2 == 0 {
					img.Pix[img.PixOffset(x, y)] = 55
				}
			}
		}
		setRect(img, size/5, size/5, size/2, size/2, 180)
		setRect(img, size/2, size/3, size/4, size/4, 25)
	}

	seed := 43
	if variant == "glyph" {
		seed = 97
	} else if variant == "noise" {
		seed = 157
	} else if variant == "orbtex" {
		seed = 211
	}
	applyFeatureTexture(img, seed)

	return img
}

func buildBenchHaystack(w, h int, variant string) *image.Gray {
	if w < 64 {
		w = 64
	}
	if h < 64 {
		h = 64
	}
	img := image.NewGray(image.Rect(0, 0, w, h))
	seed := 37
	if variant == "glyph" {
		seed = 91
	} else if variant == "noise" {
		seed = 149
	} else if variant == "orbtex" {
		seed = 227
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(90 + ((x*13 + y*7 + seed) % 41))
			img.Pix[img.PixOffset(x, y)] = v
		}
	}
	applyDenseArtifactField(img, seed, variant)
	return img
}

func populateNearMatchDecoys(haystack *image.Gray, target *image.Gray, targetX, targetY int, scenario findBenchScenario) {
	if haystack == nil || target == nil {
		return
	}
	hb := haystack.Bounds()
	tb := target.Bounds()
	tw := tb.Dx()
	th := tb.Dy()
	if tw <= 0 || th <= 0 {
		return
	}

	seed := len(scenario.name)*31 + scenario.rotation*7 + scenario.screenW + scenario.screenH + tw*11 + th*13
	stepX := maxInt(6, tw/3)
	stepY := maxInt(6, th/3)
	index := 0
	for y := hb.Min.Y - th/2; y < hb.Max.Y+th/2; y += stepY {
		for x := hb.Min.X - tw/2; x < hb.Max.X+tw/2; x += stepX {
			jx := ((seed + x*13 + y*7 + index*5) % maxInt(2, stepX)) - stepX/2
			jy := ((seed + x*11 + y*17 + index*3) % maxInt(2, stepY)) - stepY/2
			px := x + jx
			py := y + jy
			// Keep exact target location unique.
			if px == targetX && py == targetY {
				continue
			}
			if (index+seed)%19 == 0 {
				index++
				continue
			}
			decoy := makeNearMatchVariant(target, seed+index*17)
			blitGrayFeather(haystack, decoy, px, py, maxInt(2, minInt(tw, th)/10))
			index++
		}
	}
}

func makeNearMatchVariant(src *image.Gray, seed int) *image.Gray {
	dst := cloneGray(src)
	b := dst.Bounds()
	w := b.Dx()
	h := b.Dy()
	total := maxInt(1, w*h)

	// Spread subtle brightness deltas across ~10-14% of pixels.
	mutationCount := total/9 + (seed % maxInt(1, total/20+1))
	for i := 0; i < mutationCount; i++ {
		x := (i*73 + seed*19 + (i*i)%137) % w
		y := (i*97 + seed*11 + (i*i)%149) % h
		delta := ((i*17 + seed*7) % 19) - 9
		off := dst.PixOffset(b.Min.X+x, b.Min.Y+y)
		v := int(dst.Pix[off]) + delta
		if v < 0 {
			v = 0
		} else if v > 255 {
			v = 255
		}
		dst.Pix[off] = uint8(v)
	}

	// Add faint line cuts that remain visually subtle but break exact identity.
	if w > 6 && h > 6 {
		lineY := 1 + ((seed*13 + w) % maxInt(1, h-2))
		lineX := 1 + ((seed*17 + h) % maxInt(1, w-2))
		for x := 1; x < w-1; x++ {
			off := dst.PixOffset(b.Min.X+x, b.Min.Y+lineY)
			v := int(dst.Pix[off]) + ((x+seed)%7 - 3)
			if v < 0 {
				v = 0
			} else if v > 255 {
				v = 255
			}
			dst.Pix[off] = uint8(v)
		}
		for y := 1; y < h-1; y++ {
			off := dst.PixOffset(b.Min.X+lineX, b.Min.Y+y)
			v := int(dst.Pix[off]) + ((y+seed)%7 - 3)
			if v < 0 {
				v = 0
			} else if v > 255 {
				v = 255
			}
			dst.Pix[off] = uint8(v)
		}
	}

	return dst
}

func applyDenseArtifactField(img *image.Gray, seed int, variant string) {
	if img == nil {
		return
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			off := img.PixOffset(x, y)
			v := int(img.Pix[off])
			hash := (x*29 + y*41 + seed*17 + (x*y)%233) & 0xFF
			delta := (hash % 9) - 4
			if variant == "noise" || variant == "orbtex" {
				delta += (hash % 7) - 3
			}
			v += delta
			if v < 0 {
				v = 0
			} else if v > 255 {
				v = 255
			}
			img.Pix[off] = uint8(v)
		}
	}

	spacing := 8
	if variant == "noise" || variant == "orbtex" {
		spacing = 6
	}
	for y := b.Min.Y + 2; y < b.Max.Y-2; y += spacing {
		for x := b.Min.X + 2; x < b.Max.X-2; x += spacing {
			base := uint8((x*11 + y*13 + seed*23) & 0xFF)
			img.Pix[img.PixOffset(x, y)] = base
			img.Pix[img.PixOffset(x+1, y)] = 255 - base/2
			img.Pix[img.PixOffset(x, y+1)] = base / 2
			if (x+y+seed)%3 == 0 {
				img.Pix[img.PixOffset(x-1, y)] = 255 - base
			}
		}
	}
}

func blitGray(dst *image.Gray, src *image.Gray, atX, atY int) {
	db := dst.Bounds()
	sb := src.Bounds()
	for y := 0; y < sb.Dy(); y++ {
		for x := 0; x < sb.Dx(); x++ {
			dx := atX + x
			dy := atY + y
			if dx < db.Min.X || dy < db.Min.Y || dx >= db.Max.X || dy >= db.Max.Y {
				continue
			}
			dst.Pix[dst.PixOffset(dx, dy)] = src.Pix[src.PixOffset(sb.Min.X+x, sb.Min.Y+y)]
		}
	}
}

func blitGrayFeather(dst *image.Gray, src *image.Gray, atX, atY, feather int) {
	if feather < 1 {
		blitGray(dst, src, atX, atY)
		return
	}
	db := dst.Bounds()
	sb := src.Bounds()
	sw := sb.Dx()
	sh := sb.Dy()
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			dx := atX + x
			dy := atY + y
			if dx < db.Min.X || dy < db.Min.Y || dx >= db.Max.X || dy >= db.Max.Y {
				continue
			}
			srcV := int(src.Pix[src.PixOffset(sb.Min.X+x, sb.Min.Y+y)])
			dstOff := dst.PixOffset(dx, dy)
			dstV := int(dst.Pix[dstOff])

			distL := x
			distR := sw - 1 - x
			distT := y
			distB := sh - 1 - y
			edgeDist := minInt(minInt(distL, distR), minInt(distT, distB))
			if edgeDist >= feather {
				dst.Pix[dstOff] = uint8(srcV)
				continue
			}
			alphaNum := edgeDist + 1
			alphaDen := feather + 1
			blended := (dstV*(alphaDen-alphaNum) + srcV*alphaNum) / alphaDen
			if blended < 0 {
				blended = 0
			} else if blended > 255 {
				blended = 255
			}
			dst.Pix[dstOff] = uint8(blended)
		}
	}
}

func applySeamSmoothing(img *image.Gray, variant string) {
	if img == nil {
		return
	}
	passes := 1
	if variant == "noise" || variant == "orbtex" {
		passes = 2
	}
	for pass := 0; pass < passes; pass++ {
		src := cloneGray(img)
		b := src.Bounds()
		for y := b.Min.Y + 1; y < b.Max.Y-1; y++ {
			for x := b.Min.X + 1; x < b.Max.X-1; x++ {
				c := int(src.Pix[src.PixOffset(x, y)])
				u := int(src.Pix[src.PixOffset(x, y-1)])
				d := int(src.Pix[src.PixOffset(x, y+1)])
				l := int(src.Pix[src.PixOffset(x-1, y)])
				r := int(src.Pix[src.PixOffset(x+1, y)])
				mix := (c*5 + u + d + l + r) / 9
				img.Pix[img.PixOffset(x, y)] = uint8(mix)
			}
		}
	}
}

func cropGray(src *image.Gray, r image.Rectangle) *image.Gray {
	b := src.Bounds()
	c := r.Intersect(b)
	if c.Empty() {
		return image.NewGray(image.Rect(0, 0, 1, 1))
	}
	out := image.NewGray(image.Rect(0, 0, c.Dx(), c.Dy()))
	for y := 0; y < c.Dy(); y++ {
		srcStart := src.PixOffset(c.Min.X, c.Min.Y+y)
		srcEnd := srcStart + c.Dx()
		dstStart := y * out.Stride
		copy(out.Pix[dstStart:dstStart+c.Dx()], src.Pix[srcStart:srcEnd])
	}
	return out
}

func rotateGrayByQuarterTurns(src *image.Gray, degrees int) *image.Gray {
	turns := ((degrees % 360) + 360) % 360 / 90
	switch turns {
	case 0:
		return cloneGray(src)
	case 1:
		return rotate90Gray(src)
	case 2:
		return rotate180Gray(src)
	case 3:
		return rotate270Gray(src)
	default:
		return cloneGray(src)
	}
}

func applyBenchTransform(src *image.Gray, scenario findBenchScenario) *image.Gray {
	switch scenario.transformKind {
	case "":
		return cloneGray(src)
	case "scale":
		factor := scenario.transformA
		if factor <= 0 {
			factor = 1.0
		}
		return scaleGrayNearestBench(src, factor)
	case "rotate":
		return rotateGrayBilinearBench(src, scenario.transformA, 128)
	case "perspective":
		topScale := scenario.transformA
		bottomScale := scenario.transformB
		shift := scenario.transformC
		if topScale <= 0 {
			topScale = 0.90
		}
		if bottomScale <= 0 {
			bottomScale = 1.08
		}
		return perspectiveKeystoneBench(src, topScale, bottomScale, shift, 128)
	case "skewx":
		return skewGrayXBench(src, scenario.transformA, 128)
	default:
		return cloneGray(src)
	}
}

func scaleGrayNearestBench(src *image.Gray, factor float64) *image.Gray {
	if factor <= 0 {
		factor = 1.0
	}
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dw := maxInt(1, int(math.Round(float64(sw)*factor)))
	dh := maxInt(1, int(math.Round(float64(sh)*factor)))
	dst := image.NewGray(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		sy := sb.Min.Y + minInt(sh-1, int(float64(y)/factor))
		for x := 0; x < dw; x++ {
			sx := sb.Min.X + minInt(sw-1, int(float64(x)/factor))
			dst.SetGray(x, y, src.GrayAt(sx, sy))
		}
	}
	return dst
}

func rotateGrayBilinearBench(src *image.Gray, degrees float64, bg uint8) *image.Gray {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	if sw <= 0 || sh <= 0 {
		return image.NewGray(image.Rect(0, 0, 1, 1))
	}
	theta := degrees * math.Pi / 180.0
	cosT := math.Cos(theta)
	sinT := math.Sin(theta)
	cx := (float64(sw) - 1) / 2.0
	cy := (float64(sh) - 1) / 2.0

	corners := [][2]float64{
		{-cx, -cy},
		{float64(sw-1) - cx, -cy},
		{float64(sw-1) - cx, float64(sh-1) - cy},
		{-cx, float64(sh-1) - cy},
	}
	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64
	for _, c := range corners {
		x := cosT*c[0] - sinT*c[1]
		y := sinT*c[0] + cosT*c[1]
		minX = math.Min(minX, x)
		maxX = math.Max(maxX, x)
		minY = math.Min(minY, y)
		maxY = math.Max(maxY, y)
	}
	dw := maxInt(1, int(math.Ceil(maxX-minX))+1)
	dh := maxInt(1, int(math.Ceil(maxY-minY))+1)
	dst := image.NewGray(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			dxr := float64(x) + minX
			dyr := float64(y) + minY
			sxr := cosT*dxr + sinT*dyr
			syr := -sinT*dxr + cosT*dyr
			sx := sxr + cx
			sy := syr + cy
			dst.SetGray(x, y, color.Gray{Y: sampleGrayBilinearBench(src, sx, sy, bg)})
		}
	}
	return dst
}

func skewGrayXBench(src *image.Gray, skew float64, bg uint8) *image.Gray {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	cx := (float64(sw) - 1) / 2.0
	cy := (float64(sh) - 1) / 2.0

	corners := [][2]float64{
		{-cx + skew*(-cy), -cy},
		{float64(sw-1) - cx + skew*(-cy), -cy},
		{float64(sw-1) - cx + skew*(float64(sh-1)-cy), float64(sh-1) - cy},
		{-cx + skew*(float64(sh-1)-cy), float64(sh-1) - cy},
	}
	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	for _, c := range corners {
		minX = math.Min(minX, c[0])
		maxX = math.Max(maxX, c[0])
	}
	dw := maxInt(1, int(math.Ceil(maxX-minX))+1)
	dh := sh
	dst := image.NewGray(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		yr := float64(y) - cy
		for x := 0; x < dw; x++ {
			xr := float64(x) + minX
			sxr := xr - skew*yr
			sx := sxr + cx
			sy := yr + cy
			dst.SetGray(x, y, color.Gray{Y: sampleGrayBilinearBench(src, sx, sy, bg)})
		}
	}
	return dst
}

func perspectiveKeystoneBench(src *image.Gray, topScale, bottomScale, shift float64, bg uint8) *image.Gray {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dst := image.NewGray(image.Rect(0, 0, sw, sh))
	cx := (float64(sw) - 1) / 2.0
	for y := 0; y < sh; y++ {
		t := 0.0
		if sh > 1 {
			t = float64(y) / float64(sh-1)
		}
		scale := topScale*(1-t) + bottomScale*t
		shiftX := shift * (0.5 - t) * float64(sw)
		for x := 0; x < sw; x++ {
			sx := (float64(x)-cx-shiftX)/scale + cx
			sy := float64(y)
			dst.SetGray(x, y, color.Gray{Y: sampleGrayBilinearBench(src, sx, sy, bg)})
		}
	}
	return dst
}

func sampleGrayBilinearBench(src *image.Gray, fx, fy float64, bg uint8) uint8 {
	sb := src.Bounds()
	if fx < float64(sb.Min.X) || fy < float64(sb.Min.Y) || fx > float64(sb.Max.X-1) || fy > float64(sb.Max.Y-1) {
		return bg
	}
	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	x1 := minInt(sb.Max.X-1, x0+1)
	y1 := minInt(sb.Max.Y-1, y0+1)
	dx := fx - float64(x0)
	dy := fy - float64(y0)
	p00 := float64(src.GrayAt(x0, y0).Y)
	p10 := float64(src.GrayAt(x1, y0).Y)
	p01 := float64(src.GrayAt(x0, y1).Y)
	p11 := float64(src.GrayAt(x1, y1).Y)
	top := p00*(1-dx) + p10*dx
	bot := p01*(1-dx) + p11*dx
	v := int(math.Round(top*(1-dy) + bot*dy))
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func rotate90Gray(src *image.Gray) *image.Gray {
	sb := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, sb.Dy(), sb.Dx()))
	for y := 0; y < sb.Dy(); y++ {
		for x := 0; x < sb.Dx(); x++ {
			dx := sb.Dy() - 1 - y
			dy := x
			dst.Pix[dst.PixOffset(dx, dy)] = src.Pix[src.PixOffset(sb.Min.X+x, sb.Min.Y+y)]
		}
	}
	return dst
}

func rotate180Gray(src *image.Gray) *image.Gray {
	sb := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, sb.Dx(), sb.Dy()))
	for y := 0; y < sb.Dy(); y++ {
		for x := 0; x < sb.Dx(); x++ {
			dx := sb.Dx() - 1 - x
			dy := sb.Dy() - 1 - y
			dst.Pix[dst.PixOffset(dx, dy)] = src.Pix[src.PixOffset(sb.Min.X+x, sb.Min.Y+y)]
		}
	}
	return dst
}

func rotate270Gray(src *image.Gray) *image.Gray {
	sb := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, sb.Dy(), sb.Dx()))
	for y := 0; y < sb.Dy(); y++ {
		for x := 0; x < sb.Dx(); x++ {
			dx := y
			dy := sb.Dx() - 1 - x
			dst.Pix[dst.PixOffset(dx, dy)] = src.Pix[src.PixOffset(sb.Min.X+x, sb.Min.Y+y)]
		}
	}
	return dst
}

func cloneGray(src *image.Gray) *image.Gray {
	sb := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, sb.Dx(), sb.Dy()))
	for y := 0; y < sb.Dy(); y++ {
		copy(dst.Pix[y*dst.Stride:y*dst.Stride+sb.Dx()], src.Pix[(sb.Min.Y+y)*src.Stride+sb.Min.X:(sb.Min.Y+y)*src.Stride+sb.Min.X+sb.Dx()])
	}
	return dst
}

func grayProtoFromGray(name string, in *image.Gray) *pb.GrayImage {
	b := in.Bounds()
	pix := make([]byte, 0, b.Dx()*b.Dy())
	for y := 0; y < b.Dy(); y++ {
		rowStart := (b.Min.Y+y)*in.Stride + b.Min.X
		pix = append(pix, in.Pix[rowStart:rowStart+b.Dx()]...)
	}
	return &pb.GrayImage{
		Name:   name,
		Width:  int32(b.Dx()),
		Height: int32(b.Dy()),
		Pix:    pix,
	}
}

func rectMatchSatisfies(got *pb.Rect, want *pb.Rect, minOverlap float64, maxAreaRatio float64) bool {
	if maxAreaRatio <= 0 {
		maxAreaRatio = 1.50
	}
	if rectAreaRatio(got, want) > maxAreaRatio {
		return false
	}
	return rectOverlapRatio(got, want) >= math.Max(0.0, math.Min(1.0, minOverlap))
}

func rectAreaRatio(got *pb.Rect, want *pb.Rect) float64 {
	gotArea := float64(max32(1, got.GetW()*got.GetH()))
	wantArea := float64(max32(1, want.GetW()*want.GetH()))
	return gotArea / wantArea
}

func rectOverlapRatio(got *pb.Rect, want *pb.Rect) float64 {
	gx0, gy0 := got.GetX(), got.GetY()
	gx1, gy1 := gx0+got.GetW(), gy0+got.GetH()
	wx0, wy0 := want.GetX(), want.GetY()
	wx1, wy1 := wx0+want.GetW(), wy0+want.GetH()

	ix0 := max32(gx0, wx0)
	iy0 := max32(gy0, wy0)
	ix1 := min32(gx1, wx1)
	iy1 := min32(gy1, wy1)
	if ix1 <= ix0 || iy1 <= iy0 {
		return 0
	}
	interArea := float64((ix1 - ix0) * (iy1 - iy0))
	wantArea := float64(max32(1, want.GetW()*want.GetH()))
	return interArea / wantArea
}

func setRect(img *image.Gray, x, y, w, h int, value uint8) {
	b := img.Bounds()
	for yy := y; yy < y+h; yy++ {
		if yy < b.Min.Y || yy >= b.Max.Y {
			continue
		}
		for xx := x; xx < x+w; xx++ {
			if xx < b.Min.X || xx >= b.Max.X {
				continue
			}
			img.Pix[img.PixOffset(xx, yy)] = value
		}
	}
}

func applyFeatureTexture(img *image.Gray, seed int) {
	b := img.Bounds()
	for y := 2; y < b.Dy()-2; y++ {
		for x := 2; x < b.Dx()-2; x++ {
			base := int(img.Pix[img.PixOffset(x, y)])
			hash := (x*131 + y*197 + seed*53 + (x*y)%251) & 0xFF
			mix := (base*3 + hash) / 4
			if ((x+y+seed)%5 == 0) || ((x*7+y*11+seed)%13 == 0) {
				mix = (base + hash) / 2
			}
			img.Pix[img.PixOffset(x, y)] = uint8(mix)
		}
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func benchSimilarityPtr(engine string) *float64 {
	v := 0.99
	if engine == "orb" {
		v = 0.10
	}
	return &v
}
