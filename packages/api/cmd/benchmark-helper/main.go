package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type regionSpecDocument struct {
	SchemaVersion string                 `json:"schema_version"`
	Images        []regionSpecImageEntry `json:"images"`
}

type regionSpecImageEntry struct {
	ID                 string                   `json:"id"`
	ScenarioTypeIDs    []string                 `json:"scenario_type_ids"`
	ImagePath          string                   `json:"image_path"`
	SourceImagePath    string                   `json:"source_image_path"`
	BenchmarkImagePath string                   `json:"benchmark_image_path"`
	PreviewImages      []regionSpecPreviewImage `json:"preview_images"`
	Targets            []regionSpecTarget       `json:"targets"`
	BenchmarkTargets   []regionSpecTarget       `json:"benchmark_targets,omitempty"`
}

type regionSpecPreviewImage struct {
	Label     string `json:"label"`
	ImagePath string `json:"image_path"`
}

type regionSpecTarget struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	W     int    `json:"w"`
	H     int    `json:"h"`
}

type manifestBootstrap struct {
	ScenarioTypes []struct {
		ID     string `json:"id"`
		Target struct {
			AssetPool []string `json:"asset_pool"`
		} `json:"target"`
	} `json:"scenario_types"`
}

type helperServer struct {
	mu       sync.RWMutex
	repoRoot string
	specPath string
	spec     regionSpecDocument
}

func main() {
	listenAddr := flag.String("listen", ":8091", "HTTP listen address")
	specPath := flag.String("spec", filepath.Join("packages", "api", "internal", "grpcv1", "testdata", "find-bench-assets", "regions.json"), "Path to region spec JSON file")
	manifestPath := flag.String("manifest", filepath.Join("docs", "bench", "find-on-screen-scenarios.example.json"), "Manifest used to bootstrap region spec when missing")
	flag.Parse()

	repoRoot := findRepoRoot(".")
	if repoRoot == "" {
		cwd, _ := os.Getwd()
		repoRoot = cwd
	}

	resolvedSpecPath, err := resolvePath(repoRoot, "", *specPath)
	if err != nil {
		log.Fatalf("resolve spec path %q: %v", *specPath, err)
	}
	if err := os.MkdirAll(filepath.Dir(resolvedSpecPath), 0o755); err != nil {
		log.Fatalf("mkdir spec dir: %v", err)
	}

	doc, initialized, err := loadOrBootstrapSpec(repoRoot, resolvedSpecPath, *manifestPath)
	if err != nil {
		log.Fatalf("load region spec: %v", err)
	}
	if initialized {
		log.Printf("initialized region spec at %s", resolvedSpecPath)
	}

	srv := &helperServer{
		repoRoot: repoRoot,
		specPath: resolvedSpecPath,
		spec:     doc,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/api/spec", srv.handleSpec)
	mux.HandleFunc("/api/image/", srv.handleImage)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	httpSrv := &http.Server{
		Addr:              *listenAddr,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("benchmark-helper listening http=%s spec=%s", *listenAddr, resolvedSpecPath)
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}
}

func loadOrBootstrapSpec(repoRoot, specPath, manifestPath string) (regionSpecDocument, bool, error) {
	if raw, err := os.ReadFile(specPath); err == nil {
		var doc regionSpecDocument
		if err := json.Unmarshal(raw, &doc); err != nil {
			return regionSpecDocument{}, false, fmt.Errorf("parse %s: %w", specPath, err)
		}
		if err := validateSpec(doc); err != nil {
			return regionSpecDocument{}, false, fmt.Errorf("validate %s: %w", specPath, err)
		}
		sortSpec(&doc)
		return doc, false, nil
	}

	doc := bootstrapSpec(repoRoot, manifestPath)
	if err := validateSpec(doc); err != nil {
		return regionSpecDocument{}, false, err
	}
	if err := writeSpecAtomic(specPath, doc); err != nil {
		return regionSpecDocument{}, false, err
	}
	return doc, true, nil
}

func bootstrapSpec(repoRoot, manifestPath string) regionSpecDocument {
	defaultImage := filepath.ToSlash(filepath.Join("docs", "bench", "assets", "photo", "4256_clutter_crop_zoom.jpg"))
	scenarioImagePath := func(id string) string {
		extensions := []string{".png", ".jpg", ".jpeg", ".webp", ".avif"}
		roots := []string{
			filepath.Join("packages", "api", "internal", "grpcv1", "testdata", "find-bench-assets", "scenario", "source"),
			filepath.Join("packages", "api", "internal", "grpcv1", "testdata", "find-bench-assets", "scenario"),
		}
		for _, root := range roots {
			for _, ext := range extensions {
				path := filepath.ToSlash(filepath.Join(root, id+ext))
				if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(path))); err == nil {
					return path
				}
			}
		}
		return defaultImage
	}
	doc := regionSpecDocument{SchemaVersion: "1.0.0", Images: []regionSpecImageEntry{}}

	resolvedManifest, err := resolvePath(repoRoot, "", manifestPath)
	if err != nil {
		return doc
	}
	raw, err := os.ReadFile(resolvedManifest)
	if err != nil {
		return doc
	}
	var mf manifestBootstrap
	if err := json.Unmarshal(raw, &mf); err != nil {
		return doc
	}

	for _, st := range mf.ScenarioTypes {
		sid := strings.TrimSpace(st.ID)
		if sid == "" {
			continue
		}
		imagePath := defaultImage
		if len(st.Target.AssetPool) > 0 {
			candidate := strings.TrimSpace(st.Target.AssetPool[0])
			if candidate != "" {
				imagePath = candidate
			}
		} else {
			imagePath = scenarioImagePath(sid)
		}
		doc.Images = append(doc.Images, regionSpecImageEntry{
			ID:                 sid,
			ScenarioTypeIDs:    []string{sid},
			ImagePath:          filepath.ToSlash(imagePath),
			SourceImagePath:    filepath.ToSlash(imagePath),
			BenchmarkImagePath: filepath.ToSlash(imagePath),
			PreviewImages:      []regionSpecPreviewImage{},
			Targets:            []regionSpecTarget{},
		})
	}
	sortSpec(&doc)
	return doc
}

func (s *helperServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, indexHTML)
}

func (s *helperServer) handleSpec(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		doc := s.spec
		s.mu.RUnlock()
		writeJSON(w, http.StatusOK, doc)
		return
	case http.MethodPut:
		body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
		if err != nil {
			http.Error(w, fmt.Sprintf("read body: %v", err), http.StatusBadRequest)
			return
		}
		var doc regionSpecDocument
		if err := json.Unmarshal(body, &doc); err != nil {
			http.Error(w, fmt.Sprintf("parse json: %v", err), http.StatusBadRequest)
			return
		}
		if err := validateSpec(doc); err != nil {
			http.Error(w, fmt.Sprintf("invalid spec: %v", err), http.StatusBadRequest)
			return
		}
		sortSpec(&doc)
		if err := writeSpecAtomic(s.specPath, doc); err != nil {
			http.Error(w, fmt.Sprintf("write spec: %v", err), http.StatusInternalServerError)
			return
		}
		s.mu.Lock()
		s.spec = doc
		s.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "saved_at": time.Now().UTC().Format(time.RFC3339Nano)})
		return
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (s *helperServer) handleImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/image/")
	id = strings.TrimSpace(id)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	s.mu.RLock()
	entry, ok := findImageByID(s.spec.Images, id)
	s.mu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	imagePath := imagePathForMode(entry, r.URL.Query().Get("mode"), r.URL.Query().Get("idx"))
	resolved, err := resolvePath(s.repoRoot, filepath.Dir(s.specPath), imagePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("resolve image %s: %v", imagePath, err), http.StatusNotFound)
		return
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		http.Error(w, fmt.Sprintf("read image: %v", err), http.StatusNotFound)
		return
	}
	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(resolved)))
	if ct == "" {
		ct = http.DetectContentType(raw)
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(raw)
}

func imagePathForMode(entry regionSpecImageEntry, modeRaw string, idxRaw string) string {
	sourcePath := strings.TrimSpace(entry.SourceImagePath)
	if sourcePath == "" {
		sourcePath = strings.TrimSpace(entry.ImagePath)
	}
	benchmarkPath := strings.TrimSpace(entry.BenchmarkImagePath)
	if benchmarkPath == "" {
		benchmarkPath = sourcePath
	}

	mode := strings.ToLower(strings.TrimSpace(modeRaw))
	switch mode {
	case "benchmark":
		return benchmarkPath
	case "preview":
		idx := 0
		if idxRaw != "" {
			parsed, err := strconv.Atoi(strings.TrimSpace(idxRaw))
			if err != nil {
				idx = 0
			} else {
				idx = parsed
			}
		}
		if idx >= 0 && idx < len(entry.PreviewImages) {
			if path := strings.TrimSpace(entry.PreviewImages[idx].ImagePath); path != "" {
				return path
			}
		}
		return benchmarkPath
	default:
		return sourcePath
	}
}

func findImageByID(images []regionSpecImageEntry, id string) (regionSpecImageEntry, bool) {
	for _, entry := range images {
		if strings.TrimSpace(entry.ID) == id {
			return entry, true
		}
	}
	return regionSpecImageEntry{}, false
}

func validateSpec(doc regionSpecDocument) error {
	if strings.TrimSpace(doc.SchemaVersion) == "" {
		return fmt.Errorf("schema_version is required")
	}
	if len(doc.Images) == 0 {
		return fmt.Errorf("images must contain at least one entry")
	}
	seenIDs := map[string]struct{}{}
	for i, img := range doc.Images {
		imgID := strings.TrimSpace(img.ID)
		if imgID == "" {
			return fmt.Errorf("images[%d].id is required", i)
		}
		if _, ok := seenIDs[imgID]; ok {
			return fmt.Errorf("images[%d].id duplicates %q", i, imgID)
		}
		seenIDs[imgID] = struct{}{}
		sourcePath := strings.TrimSpace(img.SourceImagePath)
		if sourcePath == "" {
			sourcePath = strings.TrimSpace(img.ImagePath)
		}
		if sourcePath == "" {
			return fmt.Errorf("images[%d] must include source_image_path or image_path", i)
		}
		for j, preview := range img.PreviewImages {
			if strings.TrimSpace(preview.ImagePath) == "" {
				return fmt.Errorf("images[%d].preview_images[%d].image_path is required", i, j)
			}
		}
		for j, t := range img.Targets {
			if t.W < 1 || t.H < 1 {
				return fmt.Errorf("images[%d].targets[%d] must have w/h >= 1", i, j)
			}
		}
		for j, t := range img.BenchmarkTargets {
			if t.W < 1 || t.H < 1 {
				return fmt.Errorf("images[%d].benchmark_targets[%d] must have w/h >= 1", i, j)
			}
		}
	}
	return nil
}

func sortSpec(doc *regionSpecDocument) {
	if doc == nil {
		return
	}
	sort.Slice(doc.Images, func(i, j int) bool {
		return strings.TrimSpace(doc.Images[i].ID) < strings.TrimSpace(doc.Images[j].ID)
	})
	for i := range doc.Images {
		sort.Slice(doc.Images[i].ScenarioTypeIDs, func(a, b int) bool {
			return doc.Images[i].ScenarioTypeIDs[a] < doc.Images[i].ScenarioTypeIDs[b]
		})
		sort.Slice(doc.Images[i].PreviewImages, func(a, b int) bool {
			al := strings.TrimSpace(doc.Images[i].PreviewImages[a].Label)
			bl := strings.TrimSpace(doc.Images[i].PreviewImages[b].Label)
			if al == bl {
				return strings.TrimSpace(doc.Images[i].PreviewImages[a].ImagePath) < strings.TrimSpace(doc.Images[i].PreviewImages[b].ImagePath)
			}
			return al < bl
		})
		sort.Slice(doc.Images[i].Targets, func(a, b int) bool {
			return strings.TrimSpace(doc.Images[i].Targets[a].ID) < strings.TrimSpace(doc.Images[i].Targets[b].ID)
		})
	}
}

func writeSpecAtomic(path string, doc regionSpecDocument) error {
	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".region-spec-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := io.Copy(tmp, bytes.NewReader(encoded)); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func resolvePath(repoRoot, specDir, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("path is empty")
	}
	candidates := make([]string, 0, 8)
	seen := map[string]struct{}{}
	add := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		candidates = append(candidates, clean)
	}
	if filepath.IsAbs(raw) {
		add(raw)
	} else {
		if specDir != "" {
			add(filepath.Join(specDir, raw))
		}
		if repoRoot != "" {
			add(filepath.Join(repoRoot, raw))
		}
		cwd, _ := os.Getwd()
		add(filepath.Join(cwd, raw))
		add(raw)
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		return candidate, nil
	}
	return "", fmt.Errorf("file not found: %s", raw)
}

func findRepoRoot(start string) string {
	abs, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	dir := abs
	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

//go:embed index.html
var indexHTML string
