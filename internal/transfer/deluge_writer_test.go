package transfer

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeebo/bencode"
)

func TestWriteDelugeFastresumeMerge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "torrents.fastresume")
	orig := map[string][]byte{
		"old": []byte("olddata"),
	}
	encoded, err := bencode.EncodeBytes(orig)
	if err != nil {
		t.Fatalf("encode original: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0600); err != nil {
		t.Fatalf("write original: %v", err)
	}

	newData := map[string][]byte{
		"new": []byte("newdata"),
	}
	if err := WriteDelugeFastresume(dir, newData); err != nil {
		t.Fatalf("WriteDelugeFastresume error: %v", err)
	}

	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected backup torrents.fastresume.bak, got error: %v", err)
	}

	merged := map[string][]byte{}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read merged: %v", err)
	}
	if err := bencode.DecodeBytes(data, &merged); err != nil {
		t.Fatalf("decode merged: %v", err)
	}
	if string(merged["old"]) != "olddata" || string(merged["new"]) != "newdata" {
		t.Fatalf("unexpected merged data: %v", merged)
	}
}

func TestWriteDelugeState(t *testing.T) {
	python, err := findPython()
	if err != nil {
		t.Skip("python not available")
	}

	dir := t.TempDir()
	state := []DelugeTorrentState{
		{
			TorrentID: "abc",
			Filename:  "a.torrent",
			Queue:     1,
		},
	}
	if err := WriteDelugeState(dir, state); err != nil {
		t.Fatalf("WriteDelugeState error: %v", err)
	}

	statePath := filepath.Join(dir, "torrents.state")
	checkScript := strings.Join([]string{
		"import pickle, sys",
		"TorrentState = type('TorrentState', (), {})",
		"TorrentManagerState = type('TorrentManagerState', (), {})",
		"TorrentState.__module__ = 'deluge.core.torrentmanager'",
		"TorrentManagerState.__module__ = 'deluge.core.torrentmanager'",
		"class U(pickle.Unpickler):",
		"    def find_class(self, module, name):",
		"        if module == 'deluge.core.torrentmanager':",
		"            if name == 'TorrentState':",
		"                return TorrentState",
		"            if name == 'TorrentManagerState':",
		"                return TorrentManagerState",
		"        return pickle.Unpickler.find_class(self, module, name)",
		"with open(sys.argv[1], 'rb') as f:",
		"    s = U(f).load()",
		"print(len(getattr(s, 'torrents', [])))",
		"print(getattr(s.torrents[0], 'torrent_id', ''))",
	}, "\n")

	cmdArgs := append(append([]string{}, python.Args...), "-c", checkScript, statePath)
	cmd := exec.Command(python.Path, cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("python check failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected python output: %q", string(output))
	}
	if lines[0] != "1" || lines[1] != "abc" {
		t.Fatalf("unexpected python output: %q", string(output))
	}
}
