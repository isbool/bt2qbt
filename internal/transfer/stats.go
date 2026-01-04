package transfer

import (
	"strings"

	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
)

type StatsResult struct {
	Total   int
	Public  int
	Private int
	Failed  int
}

func CollectStats(opts *options.Opts, resumeItems map[string]*utorrentStructs.ResumeItem) StatsResult {
	result := StatsResult{Total: len(resumeItems)}
	for key := range resumeItems {
		normalizedKey := helpers.HandleCesu8(key)
		if strings.HasPrefix(normalizedKey, "magnet:?") {
			result.Failed++
			continue
		}

		transferStructure := CreateEmptyNewTransferStructure()
		transferStructure.Opts = opts
		HandleTorrentFilePath(&transferStructure, normalizedKey)
		if err := FindTorrentFile(&transferStructure); err != nil {
			result.Failed++
			continue
		}

		torrent := &torrentStructures.Torrent{}
		if err := helpers.DecodeTorrentFile(transferStructure.TorrentFilePath, torrent); err != nil {
			result.Failed++
			continue
		}

		if torrent.Info != nil && torrent.Info.Private != 0 {
			result.Private++
		} else {
			result.Public++
		}
	}
	return result
}
