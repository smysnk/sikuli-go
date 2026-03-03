//go:build opencv

package cv

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/smysnk/sikuligo/internal/core"
)

func TestORBMatcherTransformScenarios(t *testing.T) {
	matcher := NewORBMatcher()
	needle := makeORBTransformPattern(160)

	cases := []struct {
		name       string
		transform  func(*image.Gray) *image.Gray
		minOverlap float64
		maxArea    float64
	}{
		{
			name:       "resize",
			transform:  func(src *image.Gray) *image.Gray { return scaleGrayNearest(src, 1.15) },
			minOverlap: 0.35,
			maxArea:    1.80,
		},
		{
			name:       "rotate",
			transform:  func(src *image.Gray) *image.Gray { return rotateGrayBilinear(src, 12, 128) },
			minOverlap: 0.30,
			maxArea:    2.00,
		},
		{
			name:       "perspective",
			transform:  func(src *image.Gray) *image.Gray { return perspectiveKeystone(src, 0.90, 1.08, 0.05, 128) },
			minOverlap: 0.27,
			maxArea:    2.20,
		},
		{
			name:       "skew",
			transform:  func(src *image.Gray) *image.Gray { return skewGrayX(src, 0.10, 128) },
			minOverlap: 0.27,
			maxArea:    2.10,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			target := tc.transform(needle)
			haystack, expected := makeORBScene(target, stableSeed(tc.name))

			matches, err := matcher.Find(core.SearchRequest{
				Haystack:     haystack,
				Needle:       needle,
				Threshold:    0.0,
				ResizeFactor: 1.0,
				MaxResults:   25,
			})
			if err != nil {
				t.Fatalf("orb find failed: %v", err)
			}
			if len(matches) == 0 {
				t.Fatalf("expected at least one match")
			}

			best, overlap, areaRatio, ok := bestCandidateByOverlap(matches, expected)
			if !ok {
				t.Fatalf("no overlap candidate expected=%v got=%d", expected, len(matches))
			}
			if overlap < tc.minOverlap {
				t.Fatalf("overlap too low got=%.3f want>=%.3f best=%+v expected=%v", overlap, tc.minOverlap, best, expected)
			}
			if areaRatio > tc.maxArea {
				t.Fatalf("candidate too large got=%.2fx want<=%.2fx best=%+v expected=%v", areaRatio, tc.maxArea, best, expected)
			}
		})
	}
}

func makeORBTransformPattern(size int) *image.Gray {
	if size < 96 {
		size = 96
	}
	img := image.NewGray(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			v := 90 + ((x*17 + y*19 + (x*y)%131) % 120)
			if ((x+y)%11 == 0) || ((x*3+y*5)%13 == 0) {
				v = 255 - v
			}
			img.SetGray(x, y, color.Gray{Y: uint8(v)})
		}
	}
	for i := 3; i < size-3; i += 5 {
		fillRectGray(img, i, i, 2, 2, 0)
		fillRectGray(img, size-1-i, i, 2, 2, 255)
	}
	fillRectGray(img, size/6, size/6, size/8, size/2, 18)
	fillRectGray(img, size/2, size/4, size/5, size/12+2, 232)
	fillRectGray(img, size/3, size/2, size/10, size/3, 40)
	return img
}

func makeORBScene(target *image.Gray, seed int) (*image.Gray, image.Rectangle) {
	w, h := 960, 540
	scene := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := 104 + ((x*13 + y*17 + seed + (x*y)%173) % 48)
			scene.SetGray(x, y, color.Gray{Y: uint8(v)})
		}
	}
	applySceneTexture(scene, seed)

	tb := target.Bounds()
	tw, th := tb.Dx(), tb.Dy()
	tx := w/3 + (seed % 19)
	ty := h/2 - th/2 + ((seed / 7) % 13)
	tx = clampIntLocal(tx, 4, w-tw-4)
	ty = clampIntLocal(ty, 4, h-th-4)
	hardBlit(scene, target, tx, ty)

	return scene, image.Rect(tx, ty, tx+tw, ty+th)
}

func bestCandidateByOverlap(cands []core.MatchCandidate, expected image.Rectangle) (core.MatchCandidate, float64, float64, bool) {
	bestIdx := -1
	bestOverlap := 0.0
	bestScore := -1.0
	for i, c := range cands {
		overlap := overlapRatioCandidate(c, expected)
		if overlap > bestOverlap || (math.Abs(overlap-bestOverlap) < 1e-9 && c.Score > bestScore) {
			bestIdx = i
			bestOverlap = overlap
			bestScore = c.Score
		}
	}
	if bestIdx < 0 || bestOverlap <= 0 {
		return core.MatchCandidate{}, 0, 0, false
	}
	best := cands[bestIdx]
	return best, bestOverlap, areaRatioCandidate(best, expected), true
}

func overlapRatioCandidate(c core.MatchCandidate, expected image.Rectangle) float64 {
	r := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)
	i := r.Intersect(expected)
	if i.Empty() {
		return 0
	}
	interArea := float64(i.Dx() * i.Dy())
	expectedArea := float64(maxIntLocal(1, expected.Dx()*expected.Dy()))
	return interArea / expectedArea
}

func areaRatioCandidate(c core.MatchCandidate, expected image.Rectangle) float64 {
	foundArea := float64(maxIntLocal(1, c.W*c.H))
	expectedArea := float64(maxIntLocal(1, expected.Dx()*expected.Dy()))
	return foundArea / expectedArea
}

func fillRectGray(img *image.Gray, x, y, w, h int, v uint8) {
	b := img.Bounds()
	for yy := y; yy < y+h; yy++ {
		if yy < b.Min.Y || yy >= b.Max.Y {
			continue
		}
		for xx := x; xx < x+w; xx++ {
			if xx < b.Min.X || xx >= b.Max.X {
				continue
			}
			img.SetGray(xx, yy, color.Gray{Y: v})
		}
	}
}

func applySceneTexture(img *image.Gray, seed int) {
	b := img.Bounds()
	for y := b.Min.Y + 1; y < b.Max.Y-1; y++ {
		for x := b.Min.X + 1; x < b.Max.X-1; x++ {
			base := int(img.GrayAt(x, y).Y)
			hash := (x*31 + y*37 + seed*11 + (x*y)%197) & 0xFF
			delta := (hash % 5) - 2
			img.SetGray(x, y, color.Gray{Y: uint8(clampU8Int(base + delta))})
		}
	}
}

func stableSeed(s string) int {
	seed := 17
	for i := 0; i < len(s); i++ {
		seed = seed*31 + int(s[i])
	}
	if seed < 0 {
		seed = -seed
	}
	return seed
}

func scaleGrayNearest(src *image.Gray, factor float64) *image.Gray {
	if factor <= 0 {
		factor = 1.0
	}
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dw := maxIntLocal(1, int(math.Round(float64(sw)*factor)))
	dh := maxIntLocal(1, int(math.Round(float64(sh)*factor)))
	dst := image.NewGray(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		sy := sb.Min.Y + minIntLocal(sh-1, int(float64(y)/factor))
		for x := 0; x < dw; x++ {
			sx := sb.Min.X + minIntLocal(sw-1, int(float64(x)/factor))
			dst.SetGray(x, y, src.GrayAt(sx, sy))
		}
	}
	return dst
}

func rotateGrayBilinear(src *image.Gray, degrees float64, bg uint8) *image.Gray {
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
	dw := maxIntLocal(1, int(math.Ceil(maxX-minX))+1)
	dh := maxIntLocal(1, int(math.Ceil(maxY-minY))+1)
	dst := image.NewGray(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			dxr := float64(x) + minX
			dyr := float64(y) + minY
			sxr := cosT*dxr + sinT*dyr
			syr := -sinT*dxr + cosT*dyr
			sx := sxr + cx
			sy := syr + cy
			dst.SetGray(x, y, color.Gray{Y: sampleGrayBilinear(src, sx, sy, bg)})
		}
	}
	return dst
}

func skewGrayX(src *image.Gray, skew float64, bg uint8) *image.Gray {
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
	dw := maxIntLocal(1, int(math.Ceil(maxX-minX))+1)
	dh := sh
	dst := image.NewGray(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		yr := float64(y) - cy
		for x := 0; x < dw; x++ {
			xr := float64(x) + minX
			sxr := xr - skew*yr
			sx := sxr + cx
			sy := yr + cy
			dst.SetGray(x, y, color.Gray{Y: sampleGrayBilinear(src, sx, sy, bg)})
		}
	}
	return dst
}

func perspectiveKeystone(src *image.Gray, topScale, bottomScale, shift float64, bg uint8) *image.Gray {
	if topScale <= 0 {
		topScale = 0.9
	}
	if bottomScale <= 0 {
		bottomScale = 1.08
	}
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
			dst.SetGray(x, y, color.Gray{Y: sampleGrayBilinear(src, sx, sy, bg)})
		}
	}
	return dst
}

func hardBlit(dst *image.Gray, src *image.Gray, atX, atY int) {
	db := dst.Bounds()
	sb := src.Bounds()
	for y := 0; y < sb.Dy(); y++ {
		for x := 0; x < sb.Dx(); x++ {
			dx := atX + x
			dy := atY + y
			if dx < db.Min.X || dy < db.Min.Y || dx >= db.Max.X || dy >= db.Max.Y {
				continue
			}
			dst.SetGray(dx, dy, src.GrayAt(sb.Min.X+x, sb.Min.Y+y))
		}
	}
}

func sampleGrayBilinear(src *image.Gray, fx, fy float64, bg uint8) uint8 {
	sb := src.Bounds()
	if fx < float64(sb.Min.X) || fy < float64(sb.Min.Y) || fx > float64(sb.Max.X-1) || fy > float64(sb.Max.Y-1) {
		return bg
	}
	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	x1 := minIntLocal(sb.Max.X-1, x0+1)
	y1 := minIntLocal(sb.Max.Y-1, y0+1)
	dx := fx - float64(x0)
	dy := fy - float64(y0)
	p00 := float64(src.GrayAt(x0, y0).Y)
	p10 := float64(src.GrayAt(x1, y0).Y)
	p01 := float64(src.GrayAt(x0, y1).Y)
	p11 := float64(src.GrayAt(x1, y1).Y)
	top := p00*(1-dx) + p10*dx
	bot := p01*(1-dx) + p11*dx
	return uint8(clampU8Int(int(math.Round(top*(1-dy) + bot*dy))))
}

func clampU8Int(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

func clampIntLocal(v, minV, maxV int) int {
	if maxV < minV {
		return minV
	}
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func maxIntLocal(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minIntLocal(a, b int) int {
	if a < b {
		return a
	}
	return b
}
