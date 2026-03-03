package grpcv1

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/smysnk/sikuligo/internal/grpcv1/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type findBenchVisualCollector struct {
	outDir      string
	maxAttempts int
	rpcTimeout  time.Duration

	mu        sync.Mutex
	scenarios map[string]*findBenchScenarioVisual
}

type findBenchScenarioVisual struct {
	scenario findBenchScenario
	engines  map[string][]findBenchAttemptVisual
}

type findBenchScenarioSummaryImage struct {
	ScenarioName string
	Path         string
}

type findBenchAttemptVisual struct {
	Attempt   int
	Duration  time.Duration
	Retried   bool
	Status    string
	Error     string
	Overlap   float64
	AreaRatio float64
	Expected  *pb.Rect
	Found     *pb.Rect
	File      string
}

func newFindBenchVisualCollectorFromEnv(t testing.TB) *findBenchVisualCollector {
	t.Helper()
	if !envFlagTrue(os.Getenv("FIND_BENCH_VISUAL")) {
		return nil
	}

	outDir := strings.TrimSpace(os.Getenv("FIND_BENCH_VISUAL_DIR"))
	if outDir == "" {
		outDir = ".test-results/bench/visuals"
	}
	maxAttempts := parseEnvInt("FIND_BENCH_VISUAL_MAX_ATTEMPTS", 2)
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	rpcTimeout := parseEnvDuration("FIND_BENCH_VISUAL_TIMEOUT", 5*time.Second)

	if err := os.MkdirAll(filepath.Join(outDir, "attempts"), 0o755); err != nil {
		t.Fatalf("create benchmark visual attempts directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(outDir, "summaries"), 0o755); err != nil {
		t.Fatalf("create benchmark visual summary directory: %v", err)
	}
	t.Logf("benchmark visuals enabled: dir=%s max_attempts=%d timeout=%s", outDir, maxAttempts, rpcTimeout)

	return &findBenchVisualCollector{
		outDir:      outDir,
		maxAttempts: maxAttempts,
		rpcTimeout:  rpcTimeout,
		scenarios:   map[string]*findBenchScenarioVisual{},
	}
}

func (c *findBenchVisualCollector) CaptureAttempts(
	t testing.TB,
	client pb.SikuliServiceClient,
	request *pb.FindOnScreenRequest,
	source *image.Gray,
	expected *pb.Rect,
	engine findBenchEngine,
	scenario findBenchScenario,
) {
	t.Helper()
	if c == nil || client == nil || request == nil || source == nil || expected == nil {
		return
	}

	attempts := make([]findBenchAttemptVisual, 0, c.maxAttempts)
	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		rec := findBenchAttemptVisual{
			Attempt:  attempt,
			Retried:  attempt > 1,
			Status:   "error",
			Expected: cloneRect(expected),
		}

		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), c.rpcTimeout)
		res, err := client.FindOnScreen(ctx, request)
		cancel()
		rec.Duration = time.Since(start)

		if err != nil {
			rec.Status = visualStatusFromError(err)
			rec.Error = err.Error()
		} else {
			rect := res.GetMatch().GetRect()
			if rect == nil {
				rec.Status = "error"
				rec.Error = "missing match rect"
			} else {
				rec.Found = cloneRect(rect)
				rec.Overlap = rectOverlapRatio(rect, expected)
				rec.AreaRatio = rectAreaRatio(rect, expected)
				if rectMatchSatisfies(rect, expected, scenario.tolerance, scenario.maxAreaRatio) {
					rec.Status = "ok"
				} else {
					rec.Status = "overlap_miss"
				}
			}
		}

		path, writeErr := c.writeAttemptImage(source, scenario, engine, rec)
		if writeErr != nil {
			t.Logf("write benchmark attempt image scenario=%s engine=%s attempt=%d: %v", scenario.name, engine.name, attempt, writeErr)
		} else {
			rec.File = path
		}
		attempts = append(attempts, rec)

		if rec.Status == "ok" {
			break
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.scenarios[scenario.name]
	if !ok {
		entry = &findBenchScenarioVisual{
			scenario: scenario,
			engines:  map[string][]findBenchAttemptVisual{},
		}
		c.scenarios[scenario.name] = entry
	}
	entry.engines[engine.name] = append([]findBenchAttemptVisual(nil), attempts...)
}

func (c *findBenchVisualCollector) WriteScenarioSummaries() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	snapshots := make([]*findBenchScenarioVisual, 0, len(c.scenarios))
	for _, v := range c.scenarios {
		cp := &findBenchScenarioVisual{
			scenario: v.scenario,
			engines:  map[string][]findBenchAttemptVisual{},
		}
		for engine, rows := range v.engines {
			cp.engines[engine] = append([]findBenchAttemptVisual(nil), rows...)
		}
		snapshots = append(snapshots, cp)
	}
	c.mu.Unlock()

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].scenario.name < snapshots[j].scenario.name
	})
	summaryImages := make([]findBenchScenarioSummaryImage, 0, len(snapshots))
	for _, scenario := range snapshots {
		if err := c.writeScenarioSummary(scenario); err != nil {
			return err
		}
		summaryImages = append(summaryImages, findBenchScenarioSummaryImage{
			ScenarioName: scenario.scenario.name,
			Path:         c.scenarioSummaryPath(scenario.scenario.name),
		})
	}
	if err := c.writeRunMegaSummary(summaryImages); err != nil {
		return err
	}
	return nil
}

func (c *findBenchVisualCollector) writeAttemptImage(source *image.Gray, scenario findBenchScenario, engine findBenchEngine, rec findBenchAttemptVisual) (string, error) {
	canvas := grayToRGBA(source)
	if rec.Found != nil {
		boxColor := color.RGBA{R: 230, G: 80, B: 80, A: 255} // false positive/wrong place
		if rec.Status == "ok" {
			boxColor = color.RGBA{R: 40, G: 210, B: 95, A: 255} // correct match
		}
		drawRectOutline(canvas, rec.Found, boxColor, 3)
	}

	overlayHeight := 74
	if canvas.Bounds().Dy() < overlayHeight+8 {
		overlayHeight = maxInt(24, canvas.Bounds().Dy()/2)
	}
	draw.Draw(
		canvas,
		image.Rect(0, 0, canvas.Bounds().Dx(), overlayHeight),
		&image.Uniform{C: color.RGBA{R: 14, G: 19, B: 28, A: 220}},
		image.Point{},
		draw.Over,
	)

	line1 := fmt.Sprintf("ENGINE %s ATTEMPT %d", engine.name, rec.Attempt)
	line2 := fmt.Sprintf("STATUS %s DURATION %.3fMS RETRY %s", rec.Status, float64(rec.Duration.Microseconds())/1000.0, boolWord(rec.Retried))
	line3 := "FOUND NONE"
	if rec.Found != nil {
		line3 = fmt.Sprintf("FOUND X=%d Y=%d W=%d H=%d OVERLAP %.3f AREA %.2fx", rec.Found.GetX(), rec.Found.GetY(), rec.Found.GetW(), rec.Found.GetH(), rec.Overlap, rec.AreaRatio)
	}
	line4 := fmt.Sprintf("TARGET X=%d Y=%d W=%d H=%d", rec.Expected.GetX(), rec.Expected.GetY(), rec.Expected.GetW(), rec.Expected.GetH())
	if rec.Found != nil {
		if rec.Status == "ok" {
			line4 = "MATCH CLASS CORRECT (GREEN)"
		} else {
			line4 = "MATCH CLASS FALSE_POSITIVE (RED)"
		}
	} else if rec.Status == "not_found" {
		line4 = "MATCH CLASS NO_MATCH (NO BOX)"
	}
	if rec.Error != "" {
		line4 = truncateUpper(fmt.Sprintf("ERROR %s", rec.Error), 88)
	}

	textScale := 2
	y := 7
	y += drawTinyText(canvas, 8, y, line1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, textScale)
	y += 3
	y += drawTinyText(canvas, 8, y, line2, color.RGBA{R: 220, G: 231, B: 248, A: 255}, textScale)
	y += 3
	y += drawTinyText(canvas, 8, y, line3, color.RGBA{R: 160, G: 245, B: 183, A: 255}, textScale)
	y += 3
	_ = drawTinyText(canvas, 8, y, line4, color.RGBA{R: 250, G: 189, B: 189, A: 255}, textScale)

	scenarioDir := filepath.Join(c.outDir, "attempts", sanitizeFileToken(scenario.name))
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("engine-%s-attempt-%d-%s.png", sanitizeFileToken(engine.name), rec.Attempt, sanitizeFileToken(rec.Status))
	path := filepath.Join(scenarioDir, filename)
	if err := writePNG(path, canvas); err != nil {
		return "", err
	}
	return path, nil
}

func (c *findBenchVisualCollector) writeScenarioSummary(scenario *findBenchScenarioVisual) error {
	if scenario == nil {
		return nil
	}
	engines := make([]string, 0, len(scenario.engines))
	for name := range scenario.engines {
		engines = append(engines, name)
	}
	sort.Strings(engines)
	if len(engines) == 0 {
		return nil
	}

	rows := 1
	for _, engine := range engines {
		if n := len(scenario.engines[engine]); n > rows {
			rows = n
		}
	}

	srcW := maxInt(1, scenario.scenario.screenW)
	srcH := maxInt(1, scenario.scenario.screenH)
	tileW := 420
	tileH := maxInt(120, int(float64(srcH)*float64(tileW)/float64(srcW)))
	margin := 12
	headerH := 78
	captionH := 20
	cols := len(engines)

	width := margin + cols*(tileW+margin)
	height := headerH + margin + rows*(tileH+captionH+margin)
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 242, G: 246, B: 252, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(0, 0, width, headerH), &image.Uniform{C: color.RGBA{R: 16, G: 26, B: 40, A: 255}}, image.Point{}, draw.Src)

	title := fmt.Sprintf("SCENARIO %s RES %dx%d ROT %d VARIANT %s", scenario.scenario.name, scenario.scenario.screenW, scenario.scenario.screenH, scenario.scenario.rotation, scenario.scenario.variant)
	subtitle := "GREEN CORRECT MATCH   RED FALSE POSITIVE   NO BOX NO MATCH"
	drawTinyText(canvas, 10, 8, truncateUpper(title, 108), color.RGBA{R: 255, G: 255, B: 255, A: 255}, 2)
	drawTinyText(canvas, 10, 42, subtitle, color.RGBA{R: 190, G: 208, B: 236, A: 255}, 2)

	for col, engine := range engines {
		records := scenario.engines[engine]
		for row := 0; row < rows; row++ {
			x := margin + col*(tileW+margin)
			y := headerH + margin + row*(tileH+captionH+margin)
			tileRect := image.Rect(x, y, x+tileW, y+tileH)
			draw.Draw(canvas, tileRect, &image.Uniform{C: color.RGBA{R: 222, G: 228, B: 237, A: 255}}, image.Point{}, draw.Src)

			caption := fmt.Sprintf("ENGINE %s ATTEMPT %d NO RETRY", engine, row+1)
			if row < len(records) {
				rec := records[row]
				caption = fmt.Sprintf("ENGINE %s ATTEMPT %d %s %.3fMS RETRY %s", engine, rec.Attempt, rec.Status, float64(rec.Duration.Microseconds())/1000.0, boolWord(rec.Retried))
				img, err := readImage(rec.File)
				if err == nil {
					thumb := resizeNearest(img, tileW, tileH)
					draw.Draw(canvas, tileRect, thumb, image.Point{}, draw.Src)
				}
			}
			drawRectOutline(canvas, &pb.Rect{X: int32(x), Y: int32(y), W: int32(tileW), H: int32(tileH)}, color.RGBA{R: 82, G: 95, B: 116, A: 255}, 2)
			drawTinyText(canvas, x+2, y+tileH+4, truncateUpper(caption, 54), color.RGBA{R: 35, G: 46, B: 64, A: 255}, 1)
		}
	}

	path := c.scenarioSummaryPath(scenario.scenario.name)
	return writePNG(path, canvas)
}

func (c *findBenchVisualCollector) scenarioSummaryPath(scenarioName string) string {
	return filepath.Join(c.outDir, "summaries", fmt.Sprintf("summary-%s.png", sanitizeFileToken(scenarioName)))
}

func (c *findBenchVisualCollector) writeRunMegaSummary(images []findBenchScenarioSummaryImage) error {
	if len(images) == 0 {
		return nil
	}
	sort.Slice(images, func(i, j int) bool {
		return images[i].ScenarioName < images[j].ScenarioName
	})

	const (
		margin   = 16
		headerH  = 96
		tileW    = 680
		tileH    = 330
		captionH = 24
	)
	cols := int(math.Ceil(math.Sqrt(float64(len(images)))))
	if cols < 1 {
		cols = 1
	}
	rows := int(math.Ceil(float64(len(images)) / float64(cols)))

	width := margin + cols*(tileW+margin)
	height := headerH + margin + rows*(tileH+captionH+margin)
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 238, G: 243, B: 250, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(0, 0, width, headerH), &image.Uniform{C: color.RGBA{R: 14, G: 24, B: 39, A: 255}}, image.Point{}, draw.Src)

	title := "RUN-LEVEL MEGA SUMMARY   FINDONSCREEN BENCHMARK"
	meta := fmt.Sprintf("SCENARIOS %d   GRID %dX%d   GENERATED %s", len(images), cols, rows, time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	drawTinyText(canvas, 10, 12, title, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 2)
	drawTinyText(canvas, 10, 46, truncateUpper(meta, 120), color.RGBA{R: 184, G: 202, B: 232, A: 255}, 2)

	for i, summary := range images {
		col := i % cols
		row := i / cols
		x := margin + col*(tileW+margin)
		y := headerH + margin + row*(tileH+captionH+margin)

		tileRect := image.Rect(x, y, x+tileW, y+tileH)
		draw.Draw(canvas, tileRect, &image.Uniform{C: color.RGBA{R: 219, G: 226, B: 236, A: 255}}, image.Point{}, draw.Src)
		if img, err := readImage(summary.Path); err == nil {
			thumb := resizeNearest(img, tileW, tileH)
			draw.Draw(canvas, tileRect, thumb, image.Point{}, draw.Src)
		}
		drawRectOutline(canvas, &pb.Rect{X: int32(x), Y: int32(y), W: int32(tileW), H: int32(tileH)}, color.RGBA{R: 77, G: 90, B: 112, A: 255}, 2)
		drawTinyText(
			canvas,
			x+4,
			y+tileH+5,
			truncateUpper(fmt.Sprintf("SCENARIO %s", summary.ScenarioName), 72),
			color.RGBA{R: 33, G: 43, B: 60, A: 255},
			1,
		)
	}

	return writePNG(filepath.Join(c.outDir, "summaries", "summary-run-mega.png"), canvas)
}

func writePNG(path string, in image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, in)
}

func readImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func grayToRGBA(in *image.Gray) *image.RGBA {
	b := in.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			v := in.Pix[in.PixOffset(b.Min.X+x, b.Min.Y+y)]
			out.SetRGBA(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return out
}

func resizeNearest(src image.Image, dstW, dstH int) *image.RGBA {
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}
	sb := src.Bounds()
	sw := maxInt(1, sb.Dx())
	sh := maxInt(1, sb.Dy())
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		sy := sb.Min.Y + (y*sh)/dstH
		for x := 0; x < dstW; x++ {
			sx := sb.Min.X + (x*sw)/dstW
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func drawRectOutline(img *image.RGBA, rect *pb.Rect, c color.RGBA, thickness int) {
	if img == nil || rect == nil || thickness < 1 {
		return
	}
	b := img.Bounds()
	x0 := int(rect.GetX())
	y0 := int(rect.GetY())
	x1 := x0 + int(rect.GetW()) - 1
	y1 := y0 + int(rect.GetH()) - 1
	if x1 < x0 || y1 < y0 {
		return
	}
	for t := 0; t < thickness; t++ {
		yt := y0 + t
		yb := y1 - t
		for x := x0; x <= x1; x++ {
			setRGBAIfInBounds(img, b, x, yt, c)
			setRGBAIfInBounds(img, b, x, yb, c)
		}
		xl := x0 + t
		xr := x1 - t
		for y := y0; y <= y1; y++ {
			setRGBAIfInBounds(img, b, xl, y, c)
			setRGBAIfInBounds(img, b, xr, y, c)
		}
	}
}

func setRGBAIfInBounds(img *image.RGBA, bounds image.Rectangle, x, y int, c color.RGBA) {
	if x < bounds.Min.X || y < bounds.Min.Y || x >= bounds.Max.X || y >= bounds.Max.Y {
		return
	}
	img.SetRGBA(x, y, c)
}

func cloneRect(in *pb.Rect) *pb.Rect {
	if in == nil {
		return nil
	}
	return &pb.Rect{X: in.GetX(), Y: in.GetY(), W: in.GetW(), H: in.GetH()}
}

func visualStatusFromError(err error) string {
	code := status.Code(err)
	switch code {
	case codes.NotFound:
		return "not_found"
	case codes.Unimplemented:
		return "unsupported"
	case codes.DeadlineExceeded:
		return "timeout"
	default:
		return "error"
	}
}

func boolWord(v bool) string {
	if v {
		return "YES"
	}
	return "NO"
}

func parseEnvInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func parseEnvDuration(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	v, err := time.ParseDuration(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func envFlagTrue(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func sanitizeFileToken(in string) string {
	if in == "" {
		return "untitled"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range in {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "untitled"
	}
	return out
}

func truncateUpper(in string, maxLen int) string {
	upper := strings.ToUpper(in)
	if len(upper) <= maxLen || maxLen < 4 {
		return upper
	}
	return upper[:maxLen-3] + "..."
}

func drawTinyText(img *image.RGBA, x, y int, text string, c color.RGBA, scale int) int {
	if scale < 1 {
		scale = 1
	}
	penX := x
	upper := strings.ToUpper(text)
	for _, r := range upper {
		g, ok := tinyFont5x7[r]
		if !ok {
			g = tinyFont5x7['?']
		}
		drawTinyGlyph(img, penX, y, g, c, scale)
		penX += 6 * scale
	}
	return 8 * scale
}

func drawTinyGlyph(img *image.RGBA, x, y int, rows [7]uint8, c color.RGBA, scale int) {
	b := img.Bounds()
	for row := 0; row < 7; row++ {
		mask := rows[row]
		for col := 0; col < 5; col++ {
			if mask&(1<<(4-col)) == 0 {
				continue
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					setRGBAIfInBounds(img, b, x+col*scale+dx, y+row*scale+dy, c)
				}
			}
		}
	}
}

var tinyFont5x7 = map[rune][7]uint8{
	' ': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	'.': {0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x0C},
	',': {0x00, 0x00, 0x00, 0x00, 0x0C, 0x0C, 0x08},
	':': {0x00, 0x0C, 0x0C, 0x00, 0x0C, 0x0C, 0x00},
	'-': {0x00, 0x00, 0x00, 0x1F, 0x00, 0x00, 0x00},
	'_': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1F},
	'/': {0x01, 0x02, 0x04, 0x08, 0x10, 0x00, 0x00},
	'=': {0x00, 0x00, 0x1F, 0x00, 0x1F, 0x00, 0x00},
	'%': {0x19, 0x19, 0x02, 0x04, 0x08, 0x13, 0x13},
	'(': {0x02, 0x04, 0x08, 0x08, 0x08, 0x04, 0x02},
	')': {0x08, 0x04, 0x02, 0x02, 0x02, 0x04, 0x08},
	'?': {0x0E, 0x11, 0x01, 0x02, 0x04, 0x00, 0x04},

	'0': {0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E},
	'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'2': {0x0E, 0x11, 0x01, 0x06, 0x08, 0x10, 0x1F},
	'3': {0x0E, 0x11, 0x01, 0x06, 0x01, 0x11, 0x0E},
	'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
	'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
	'6': {0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E},
	'7': {0x1F, 0x11, 0x01, 0x02, 0x04, 0x04, 0x04},
	'8': {0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E},
	'9': {0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C},

	'A': {0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'B': {0x1E, 0x11, 0x11, 0x1E, 0x11, 0x11, 0x1E},
	'C': {0x0E, 0x11, 0x10, 0x10, 0x10, 0x11, 0x0E},
	'D': {0x1C, 0x12, 0x11, 0x11, 0x11, 0x12, 0x1C},
	'E': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x1F},
	'F': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x10},
	'G': {0x0E, 0x11, 0x10, 0x17, 0x11, 0x11, 0x0E},
	'H': {0x11, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'I': {0x0E, 0x04, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'J': {0x07, 0x02, 0x02, 0x02, 0x02, 0x12, 0x0C},
	'K': {0x11, 0x12, 0x14, 0x18, 0x14, 0x12, 0x11},
	'L': {0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x1F},
	'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
	'N': {0x11, 0x19, 0x19, 0x15, 0x13, 0x13, 0x11},
	'O': {0x0E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'P': {0x1E, 0x11, 0x11, 0x1E, 0x10, 0x10, 0x10},
	'Q': {0x0E, 0x11, 0x11, 0x11, 0x15, 0x12, 0x0D},
	'R': {0x1E, 0x11, 0x11, 0x1E, 0x14, 0x12, 0x11},
	'S': {0x0F, 0x10, 0x10, 0x0E, 0x01, 0x01, 0x1E},
	'T': {0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04},
	'U': {0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'V': {0x11, 0x11, 0x11, 0x11, 0x0A, 0x0A, 0x04},
	'W': {0x11, 0x11, 0x11, 0x15, 0x15, 0x1B, 0x11},
	'X': {0x11, 0x11, 0x0A, 0x04, 0x0A, 0x11, 0x11},
	'Y': {0x11, 0x11, 0x0A, 0x04, 0x04, 0x04, 0x04},
	'Z': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x10, 0x1F},
}
