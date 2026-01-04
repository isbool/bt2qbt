package transfer

import (
	"testing"

	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
)

func TestTransferStructure_IsPrivate(t *testing.T) {
	transferStructure := CreateEmptyNewTransferStructure()
	transferStructure.TorrentFile = &torrentStructures.Torrent{
		Info: &torrentStructures.TorrentInfo{Private: 1},
	}
	if !transferStructure.IsPrivate() {
		t.Fatalf("expected private torrent")
	}

	transferStructure.TorrentFile.Info.Private = 0
	if transferStructure.IsPrivate() {
		t.Fatalf("expected non-private torrent")
	}
}
