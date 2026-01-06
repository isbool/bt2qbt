package transfer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
)

func TestNormalizeDelugeLabel(t *testing.T) {
	tests := map[string]string{
		"My Label!":      "my_label",
		"  spaced  out ": "spaced_out",
		"Mix-Ed.Case":    "mix-ed.case",
		"___":            "",
	}
	for input, expected := range tests {
		if got := NormalizeDelugeLabel(input); got != expected {
			t.Fatalf("NormalizeDelugeLabel(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestPickDelugeLabel(t *testing.T) {
	item := &utorrentStructs.ResumeItem{
		Label:  "Movies",
		Labels: []string{"Action", "Drama"},
	}
	opts := &options.Opts{}
	label, all := PickDelugeLabel(item, opts)
	if label != "movies" {
		t.Fatalf("expected primary label movies, got %q", label)
	}
	expectedAll := []string{"movies", "action", "drama"}
	if !reflect.DeepEqual(all, expectedAll) {
		t.Fatalf("expected labels %v, got %v", expectedAll, all)
	}

	opts.WithoutLabels = true
	label, all = PickDelugeLabel(item, opts)
	if label != "action" {
		t.Fatalf("expected primary label action, got %q", label)
	}
	expectedAll = []string{"action", "drama"}
	if !reflect.DeepEqual(all, expectedAll) {
		t.Fatalf("expected labels %v, got %v", expectedAll, all)
	}
}

func TestProcessDelugeLabels(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "label.conf")

	version := map[string]interface{}{
		"format": 1,
		"file":   1,
	}
	config := map[string]interface{}{
		"labels": map[string]interface{}{
			"old": map[string]interface{}{
				"apply_max": false,
			},
		},
		"torrent_labels": map[string]interface{}{
			"abc": "old",
		},
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create label.conf: %v", err)
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(version); err != nil {
		t.Fatalf("encode version: %v", err)
	}
	if err := enc.Encode(config); err != nil {
		t.Fatalf("encode config: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close label.conf: %v", err)
	}

	opts := &options.Opts{DelugeLabels: path}
	err = ProcessDelugeLabels(opts, []string{"newlabel"}, map[string]string{"deadbeef": "newlabel"})
	if err != nil {
		t.Fatalf("ProcessDelugeLabels error: %v", err)
	}

	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected backup label.conf.bak, got error: %v", err)
	}

	readFile, err := os.Open(path)
	if err != nil {
		t.Fatalf("open label.conf: %v", err)
	}
	defer readFile.Close()

	dec := json.NewDecoder(readFile)
	var gotVersion map[string]interface{}
	if err := dec.Decode(&gotVersion); err != nil {
		t.Fatalf("decode version: %v", err)
	}
	if !versionMatches(gotVersion, version) {
		t.Fatalf("version mismatch: got %v want %v", gotVersion, version)
	}
	var gotConfig map[string]interface{}
	if err := dec.Decode(&gotConfig); err != nil {
		t.Fatalf("decode config: %v", err)
	}

	labels, ok := gotConfig["labels"].(map[string]interface{})
	if !ok {
		t.Fatalf("labels is not a map")
	}
	if _, ok := labels["old"]; !ok {
		t.Fatalf("expected old label to remain")
	}
	if _, ok := labels["newlabel"]; !ok {
		t.Fatalf("expected newlabel to be added")
	}

	torrentLabels, ok := gotConfig["torrent_labels"].(map[string]interface{})
	if !ok {
		t.Fatalf("torrent_labels is not a map")
	}
	if torrentLabels["deadbeef"] != "newlabel" {
		t.Fatalf("expected torrent_labels deadbeef to be newlabel, got %v", torrentLabels["deadbeef"])
	}
}

func versionMatches(got, want map[string]interface{}) bool {
	for _, key := range []string{"format", "file"} {
		gotVal, ok := got[key]
		if !ok {
			return false
		}
		wantVal, ok := want[key]
		if !ok {
			return false
		}
		if normalizeNumber(gotVal) != normalizeNumber(wantVal) {
			return false
		}
	}
	return true
}

func normalizeNumber(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
	}
	return 0
}
