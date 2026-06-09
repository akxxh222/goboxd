package runner

import (
    "log"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

var nsjailLineRE = regexp.MustCompile(`(?m)^\[[A-Z]\].*$`)

// extractNsjailLogs finds lines that look like nsjail diagnostics
// (e.g. "[I][...]") and returns (cleaned, extracted).
func extractNsjailLogs(s string) (string, string) {
    if s == "" {
        return "", ""
    }

    matches := nsjailLineRE.FindAllString(s, -1)
    if len(matches) == 0 {
        return s, ""
    }

    cleaned := nsjailLineRE.ReplaceAllString(s, "")
    cleaned = strings.TrimSpace(cleaned)
    extracted := strings.Join(matches, "\n")
    return cleaned, extracted
}

func processStderr(raw string, tag string) string {
    cleaned, extracted := extractNsjailLogs(raw)
    if extracted != "" {
        log.Printf("nsjail diagnostics (%s): %s", tag, extracted)
    }
    return cleaned
}

func logNsjailFile(tempDir string, tag string) {
    if data, err := os.ReadFile(filepath.Join(tempDir, "nsjail.log")); err == nil && len(data) > 0 {
        log.Printf("nsjail log (%s): %s", tag, string(data))
        _ = os.Remove(filepath.Join(tempDir, "nsjail.log"))
    }
}
