package transfer

import (
	"strings"

	"github.com/rumanzo/bt2qbt/pkg/helpers"
)

type DelugeTracker struct {
	URL  string `json:"url"`
	Tier int    `json:"tier"`
}

type DelugeTorrentState struct {
	TorrentID           string          `json:"torrent_id"`
	Filename            string          `json:"filename"`
	Trackers            []DelugeTracker `json:"trackers"`
	StorageMode         string          `json:"storage_mode"`
	Paused              bool            `json:"paused"`
	SavePath            string          `json:"save_path"`
	MaxConnections      int64           `json:"max_connections"`
	MaxUploadSlots      int64           `json:"max_upload_slots"`
	MaxUploadSpeed      float64         `json:"max_upload_speed"`
	MaxDownloadSpeed    float64         `json:"max_download_speed"`
	PrioritizeFirstLast bool            `json:"prioritize_first_last"`
	SequentialDownload  bool            `json:"sequential_download"`
	FilePriorities      []int64         `json:"file_priorities"`
	Queue               int             `json:"queue"`
	AutoManaged         bool            `json:"auto_managed"`
	IsFinished          bool            `json:"is_finished"`
	StopRatio           float64         `json:"stop_ratio"`
	StopAtRatio         bool            `json:"stop_at_ratio"`
	RemoveAtRatio       bool            `json:"remove_at_ratio"`
	MoveCompleted       bool            `json:"move_completed"`
	MoveCompletedPath   string          `json:"move_completed_path"`
	Magnet              string          `json:"magnet"`
	Owner               string          `json:"owner"`
	Shared              bool            `json:"shared"`
	SuperSeeding        bool            `json:"super_seeding"`
	Name                string          `json:"name"`
}

func buildDelugeTrackers(trackers [][]string) []DelugeTracker {
	result := make([]DelugeTracker, 0, len(trackers))
	for tier, urls := range trackers {
		for _, url := range urls {
			url = strings.TrimSpace(url)
			if url == "" {
				continue
			}
			result = append(result, DelugeTracker{URL: url, Tier: tier})
		}
	}
	return result
}

func ShouldIncludeDelugeTorrent(isPrivate bool, privacy string) bool {
	switch privacy {
	case "private":
		return isPrivate
	case "public":
		return !isPrivate
	default:
		return true
	}
}

func BuildDelugeTorrentState(transfer *TransferStructure, torrentID string, queue int, magnetURI string) DelugeTorrentState {
	storageMode := transfer.Fastresume.Allocation
	if storageMode == "" {
		storageMode = "sparse"
	}
	name := helpers.HandleCesu8(transfer.Fastresume.QbtName)
	return DelugeTorrentState{
		TorrentID:           torrentID,
		Filename:            transfer.TorrentFileName,
		Trackers:            buildDelugeTrackers(transfer.Fastresume.Trackers),
		StorageMode:         storageMode,
		Paused:              transfer.Fastresume.Paused == 1,
		SavePath:            transfer.Fastresume.SavePath,
		MaxConnections:      -1,
		MaxUploadSlots:      -1,
		MaxUploadSpeed:      -1,
		MaxDownloadSpeed:    -1,
		PrioritizeFirstLast: false,
		SequentialDownload:  transfer.Fastresume.SequentialDownload == 1,
		FilePriorities:      transfer.Fastresume.FilePriority,
		Queue:               queue,
		AutoManaged:         transfer.Fastresume.AutoManaged == 1,
		IsFinished:          transfer.ResumeItem.CompletedOn != 0,
		StopRatio:           2.0,
		StopAtRatio:         false,
		RemoveAtRatio:       false,
		MoveCompleted:       false,
		MoveCompletedPath:   "",
		Magnet:              magnetURI,
		Owner:               "",
		Shared:              false,
		SuperSeeding:        transfer.Fastresume.SuperSeeding == 1,
		Name:                name,
	}
}
