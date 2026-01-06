package transfer

import "testing"

func TestBuildDelugeTorrentState(t *testing.T) {
	transfer := CreateEmptyNewTransferStructure()
	transfer.Fastresume.Paused = 1
	transfer.Fastresume.AutoManaged = 0
	transfer.Fastresume.SequentialDownload = 1
	transfer.Fastresume.SuperSeeding = 1
	transfer.Fastresume.FilePriority = []int64{1, 0}
	transfer.Fastresume.Trackers = [][]string{
		{"http://tracker1.example/announce"},
		{"udp://tracker2.example:80/announce", ""},
	}
	transfer.Fastresume.SavePath = "/data/"
	transfer.Fastresume.Allocation = "allocate"
	transfer.Fastresume.QbtName = "My Name"
	transfer.ResumeItem.CompletedOn = 123
	transfer.TorrentFileName = "file.torrent"

	state := BuildDelugeTorrentState(&transfer, "hash", 5, "magnet:?x")
	if state.TorrentID != "hash" {
		t.Fatalf("torrent id mismatch: %v", state.TorrentID)
	}
	if state.Filename != "file.torrent" {
		t.Fatalf("filename mismatch: %v", state.Filename)
	}
	if state.StorageMode != "allocate" {
		t.Fatalf("storage mode mismatch: %v", state.StorageMode)
	}
	if !state.Paused {
		t.Fatalf("expected paused to be true")
	}
	if state.AutoManaged {
		t.Fatalf("expected auto_managed to be false")
	}
	if !state.SequentialDownload {
		t.Fatalf("expected sequential_download to be true")
	}
	if !state.SuperSeeding {
		t.Fatalf("expected super_seeding to be true")
	}
	if state.Queue != 5 {
		t.Fatalf("queue mismatch: %v", state.Queue)
	}
	if !state.IsFinished {
		t.Fatalf("expected is_finished to be true")
	}
	if state.Name != "My Name" {
		t.Fatalf("name mismatch: %v", state.Name)
	}
	if state.Magnet != "magnet:?x" {
		t.Fatalf("magnet mismatch: %v", state.Magnet)
	}
	if len(state.Trackers) != 2 {
		t.Fatalf("expected 2 trackers, got %d", len(state.Trackers))
	}
	if state.Trackers[0].Tier != 0 || state.Trackers[1].Tier != 1 {
		t.Fatalf("unexpected tracker tiers: %v", state.Trackers)
	}
}

func TestShouldIncludeDelugeTorrent(t *testing.T) {
	if !ShouldIncludeDelugeTorrent(true, "private") {
		t.Fatalf("expected private to include private torrents")
	}
	if ShouldIncludeDelugeTorrent(true, "public") {
		t.Fatalf("expected public to exclude private torrents")
	}
	if !ShouldIncludeDelugeTorrent(false, "public") {
		t.Fatalf("expected public to include public torrents")
	}
	if !ShouldIncludeDelugeTorrent(false, "all") {
		t.Fatalf("expected all to include public torrents")
	}
	if !ShouldIncludeDelugeTorrent(true, "") {
		t.Fatalf("expected empty to include private torrents")
	}
}
