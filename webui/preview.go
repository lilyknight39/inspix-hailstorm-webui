package webui

import (
	"bufio"
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const previewRoot = "cache/webui-preview"

type PreviewInfo struct {
	Available   bool
	Kind        string
	ContentType string
	Path        string
	Source      string
}

var loopRE = regexp.MustCompile(`(?i)\bloop (start|end)\b`)

func resolvePreview(entryLabel string, entryType string, resourceType uint32, plainPath string) PreviewInfo {
	if direct := sniffDirectPreview(plainPath); direct.Available {
		direct.Path = plainPath
		direct.Source = "plain"
		return direct
	}

	if isAcb(entryLabel) {
		if info := ensureAcbPreview(entryLabel, plainPath); info.Available {
			return info
		}
	}

	if isAssetBundle(entryLabel, resourceType) {
		if info := ensureAssetBundlePreview(entryLabel, plainPath); info.Available {
			return info
		}
	}

	return PreviewInfo{}
}

func sniffDirectPreview(path string) PreviewInfo {
	file, err := os.Open(path)
	if err != nil {
		return PreviewInfo{}
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buf := make([]byte, 512)
	n, _ := reader.Read(buf)
	if n == 0 {
		return PreviewInfo{}
	}
	buf = buf[:n]

	if kind, ctype := sniffKnown(buf); kind != "" {
		return PreviewInfo{
			Available:   true,
			Kind:        kind,
			ContentType: ctype,
		}
	}

	detected := http.DetectContentType(buf)
	switch {
	case strings.HasPrefix(detected, "image/"):
		return PreviewInfo{Available: true, Kind: "image", ContentType: detected}
	case strings.HasPrefix(detected, "audio/"):
		return PreviewInfo{Available: true, Kind: "audio", ContentType: detected}
	case strings.HasPrefix(detected, "video/"):
		return PreviewInfo{Available: true, Kind: "video", ContentType: detected}
	default:
		return PreviewInfo{}
	}
}

func ensureAcbPreview(label string, plainPath string) PreviewInfo {
	if !fileExists(plainPath) {
		return PreviewInfo{}
	}
	if _, err := exec.LookPath("vgmstream-cli"); err != nil {
		return PreviewInfo{}
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return PreviewInfo{}
	}

	outDir := filepath.Join(previewRoot, "acb")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return PreviewInfo{}
	}
	base := sanitizeLabel(label)
	outPath := filepath.Join(outDir, base+".mp3")

	if isFresh(outPath, plainPath) {
		return PreviewInfo{
			Available:   true,
			Kind:        "audio",
			ContentType: "audio/mpeg",
			Path:        outPath,
			Source:      "derived",
		}
	}

	tmpDir, err := os.MkdirTemp("", "hailstorm-acb")
	if err != nil {
		return PreviewInfo{}
	}
	defer os.RemoveAll(tmpDir)

	loopOnce := detectLoop(plainPath)
	tmpWav := filepath.Join(tmpDir, "preview.wav")
	vgmArgs := []string{}
	if loopOnce {
		vgmArgs = append(vgmArgs, "-L")
	}
	vgmArgs = append(vgmArgs, plainPath, "-o", tmpWav)

	if err := exec.Command("vgmstream-cli", vgmArgs...).Run(); err != nil {
		return PreviewInfo{}
	}

	if err := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-loglevel",
		"error",
		"-y",
		"-i",
		tmpWav,
		"-vn",
		"-c:a",
		"libmp3lame",
		"-q:a",
		"0",
		"-f",
		"mp3",
		outPath,
	).Run(); err != nil {
		return PreviewInfo{}
	}

	return PreviewInfo{
		Available:   true,
		Kind:        "audio",
		ContentType: "audio/mpeg",
		Path:        outPath,
		Source:      "derived",
	}
}

func ensureAssetBundlePreview(label string, plainPath string) PreviewInfo {
	outDir := filepath.Join(previewRoot, "assetbundle", sanitizeLabel(label))
	_ = os.MkdirAll(outDir, 0755)

	if info := findBundlePreview(outDir); info.Available {
		return info
	}

	cmdTemplate := strings.TrimSpace(os.Getenv("HAILSTORM_ASSETRIPPER_CMD"))
	if cmdTemplate == "" {
		return PreviewInfo{}
	}

	cmdLine := strings.ReplaceAll(cmdTemplate, "{input}", shellEscape(plainPath))
	cmdLine = strings.ReplaceAll(cmdLine, "{output}", shellEscape(outDir))
	_ = exec.Command("sh", "-c", cmdLine).Run()

	return findBundlePreview(outDir)
}

func findBundlePreview(dir string) PreviewInfo {
	if img := firstMatch(dir, []string{"*.png", "*.jpg", "*.jpeg", "*.webp"}); img != "" {
		ctype := "image/png"
		ext := strings.ToLower(filepath.Ext(img))
		switch ext {
		case ".jpg", ".jpeg":
			ctype = "image/jpeg"
		case ".webp":
			ctype = "image/webp"
		case ".png":
			ctype = "image/png"
		}
		return PreviewInfo{
			Available:   true,
			Kind:        "image",
			ContentType: ctype,
			Path:        img,
			Source:      "derived",
		}
	}
	if vid := firstMatch(dir, []string{"*.mp4", "*.webm"}); vid != "" {
		ctype := "video/mp4"
		if strings.HasSuffix(strings.ToLower(vid), ".webm") {
			ctype = "video/webm"
		}
		return PreviewInfo{
			Available:   true,
			Kind:        "video",
			ContentType: ctype,
			Path:        vid,
			Source:      "derived",
		}
	}
	if model := firstMatch(dir, []string{"*.glb", "*.gltf"}); model != "" {
		ctype := "model/gltf-binary"
		if strings.HasSuffix(strings.ToLower(model), ".gltf") {
			ctype = "model/gltf+json"
		}
		return PreviewInfo{
			Available:   true,
			Kind:        "model",
			ContentType: ctype,
			Path:        model,
			Source:      "derived",
		}
	}
	return PreviewInfo{}
}

func firstMatch(dir string, patterns []string) string {
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(dir, pattern))
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

func detectLoop(path string) bool {
	out, err := exec.Command("vgmstream-cli", "-m", path).Output()
	if err != nil {
		return false
	}
	return loopRE.Match(out)
}

func isFresh(outPath string, inputPath string) bool {
	outInfo, err := os.Stat(outPath)
	if err != nil {
		return false
	}
	inInfo, err := os.Stat(inputPath)
	if err != nil {
		return false
	}
	return outInfo.ModTime().After(inInfo.ModTime().Add(-1 * time.Second))
}

func isAcb(label string) bool {
	return strings.HasSuffix(strings.ToLower(label), ".acb")
}

func isAssetBundle(label string, resourceType uint32) bool {
	return resourceType == 1 || strings.HasSuffix(strings.ToLower(label), ".assetbundle")
}

func sanitizeLabel(label string) string {
	label = strings.ReplaceAll(label, "/", "_")
	label = strings.ReplaceAll(label, "\\", "_")
	return label
}

func shellEscape(value string) string {
	if value == "" {
		return "''"
	}
	value = strings.ReplaceAll(value, `'`, `'\''`)
	return "'" + value + "'"
}

func sniffKnown(buf []byte) (string, string) {
	if len(buf) >= 12 {
		if bytes.HasPrefix(buf, []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) {
			return "image", "image/png"
		}
		if bytes.HasPrefix(buf, []byte{0xff, 0xd8, 0xff}) {
			return "image", "image/jpeg"
		}
		if bytes.HasPrefix(buf, []byte("GIF87a")) || bytes.HasPrefix(buf, []byte("GIF89a")) {
			return "image", "image/gif"
		}
		if bytes.HasPrefix(buf, []byte("RIFF")) && bytes.HasPrefix(buf[8:], []byte("WEBP")) {
			return "image", "image/webp"
		}
		if bytes.HasPrefix(buf, []byte("RIFF")) && bytes.HasPrefix(buf[8:], []byte("WAVE")) {
			return "audio", "audio/wav"
		}
		if bytes.HasPrefix(buf, []byte("OggS")) {
			return "audio", "audio/ogg"
		}
		if bytes.HasPrefix(buf, []byte("fLaC")) {
			return "audio", "audio/flac"
		}
		if bytes.HasPrefix(buf, []byte("ID3")) || (buf[0] == 0xff && buf[1]&0xe0 == 0xe0) {
			return "audio", "audio/mpeg"
		}
		if bytes.HasPrefix(buf, []byte("glTF")) {
			return "model", "model/gltf-binary"
		}
		if bytes.HasPrefix(buf[4:], []byte("ftyp")) {
			return "video", "video/mp4"
		}
		if bytes.HasPrefix(buf, []byte{0x1a, 0x45, 0xdf, 0xa3}) {
			return "video", "video/webm"
		}
	}
	return "", ""
}
