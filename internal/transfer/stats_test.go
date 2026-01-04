package transfer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"github.com/zeebo/bencode"
)

func writeTorrentFile(t *testing.T, path string, private uint8) {
	t.Helper()
	torrent := &torrentStructures.Torrent{
		Info: &torrentStructures.TorrentInfo{
			Name:        "test",
			PieceLength: 1,
			Pieces:      []byte{0},
			Private:     private,
		},
	}
	data, err := bencode.EncodeBytes(torrent)
	if err != nil {
		t.Fatalf("encode torrent: %v", err)
	}
	if err := os.WriteFile(path, data, 0o666); err != nil {
		t.Fatalf("write torrent: %v", err)
	}
}

func TestCollectStats(t *testing.T) {
	bitDir := t.TempDir()
	searchDir := t.TempDir()

	writeTorrentFile(t, filepath.Join(bitDir, "private.torrent"), 1)
	writeTorrentFile(t, filepath.Join(searchDir, "public.torrent"), 0)

	opts := &options.Opts{
		BitDir:      bitDir,
		SearchPaths: []string{searchDir},
	}
	resumeItems := map[string]*utorrentStructs.ResumeItem{
		"private.torrent":                 {},
		"public.torrent":                  {},
		"missing.torrent":                 {},
		"magnet:?xt=urn:btih:deadbeef":    {},
		"magnet:?xt=urn:btih:deadbeef&dn": {},
	}

	stats := CollectStats(opts, resumeItems)
	if stats.Total != 5 {
		t.Fatalf("unexpected total: %v", stats.Total)
	}
	if stats.Private != 1 {
		t.Fatalf("unexpected private count: %v", stats.Private)
	}
	if stats.Public != 1 {
		t.Fatalf("unexpected public count: %v", stats.Public)
	}
	if stats.Failed != 3 {
		t.Fatalf("unexpected failed count: %v", stats.Failed)
	}
}
