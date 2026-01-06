package transfer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/torrentStructures"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
	"github.com/zeebo/bencode"
)

type delugeItemResult struct {
	TorrentID  string
	State      DelugeTorrentState
	ResumeData []byte
	IsPrivate  bool
}

func HandleResumeItemsDeluge(opts *options.Opts, resumeItems map[string]*utorrentStructs.ResumeItem) {
	totalJobs := len(resumeItems)
	replaces := CreateReplaces(opts.Replaces)

	type resumeEntry struct {
		key  string
		item *utorrentStructs.ResumeItem
	}
	items := make([]resumeEntry, 0, totalJobs)
	for key, item := range resumeItems {
		items = append(items, resumeEntry{key: key, item: item})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].item.AddedOn == items[j].item.AddedOn {
			return items[i].key < items[j].key
		}
		return items[i].item.AddedOn < items[j].item.AddedOn
	})

	resumeData := map[string][]byte{}
	torrentStates := make([]DelugeTorrentState, 0, totalJobs)
	labelAssignments := map[string]string{}
	labelSet := map[string]struct{}{}

	numJob := 1
	var publicCount int
	var privateCount int
	var failedCount int
	var skippedCount int
	var wasErrors bool

	for queue, entry := range items {
		transferStruct := CreateEmptyNewTransferStructure()
		transferStruct.ResumeItem = entry.item
		transferStruct.Replace = replaces
		transferStruct.Opts = opts

		result, skipped, err := HandleResumeItemDeluge(helpers.HandleCesu8(entry.key), &transferStruct, queue)
		if err != nil {
			fmt.Printf("%v/%v %v \n", numJob, totalJobs, err.Error())
			wasErrors = true
			failedCount++
			numJob++
			continue
		}
		if skipped {
			reason := "filtered"
			if opts.DelugePrivacy == "private" && !transferStruct.IsPrivate() {
				reason = "public"
			} else if opts.DelugePrivacy == "public" && transferStruct.IsPrivate() {
				reason = "private"
			}
			fmt.Printf("%v/%v Skipped %v (%v) \n", numJob, totalJobs, helpers.HandleCesu8(entry.key), reason)
			skippedCount++
			numJob++
			continue
		}

		resumeData[result.TorrentID] = result.ResumeData
		torrentStates = append(torrentStates, result.State)

		if result.IsPrivate {
			privateCount++
		} else {
			publicCount++
		}

		if !opts.WithoutLabels || !opts.WithoutTags {
			primary, labels := PickDelugeLabel(entry.item, opts)
			for _, label := range labels {
				labelSet[label] = struct{}{}
			}
			if primary != "" {
				labelAssignments[result.TorrentID] = primary
			}
		}

		fmt.Printf("%v/%v Sucessfully imported %v \n", numJob, totalJobs, helpers.HandleCesu8(entry.key))
		numJob++
	}

	if err := WriteDelugeFastresume(opts.DelugeStateDir, resumeData); err != nil {
		fmt.Printf("Can't write torrents.fastresume: %v\n", err)
		wasErrors = true
	}
	if err := WriteDelugeState(opts.DelugeStateDir, torrentStates); err != nil {
		fmt.Printf("Can't write torrents.state: %v\n", err)
		wasErrors = true
	}

	if !opts.WithoutLabels || !opts.WithoutTags {
		labels := make([]string, 0, len(labelSet))
		for label := range labelSet {
			labels = append(labels, label)
		}
		if err := ProcessDelugeLabels(opts, labels, labelAssignments); err != nil {
			fmt.Printf("Can't handle labels with error:\n%v\n", err)
		}
	}

	fmt.Printf("Summary: public %v, private %v, skipped %v, failed %v of %v total\n", publicCount, privateCount, skippedCount, failedCount, totalJobs)
	fmt.Println()
	if wasErrors {
		fmt.Println("Not all torrents was processed")
	}
}

func HandleResumeItemDeluge(key string, transferStruct *TransferStructure, queue int) (*delugeItemResult, bool, error) {
	var err error

	HandleTorrentFilePath(transferStruct, key)
	if err = FindTorrentFile(transferStruct); err != nil {
		return nil, false, err
	}

	if err = helpers.DecodeTorrentFile(transferStruct.TorrentFilePath, transferStruct.TorrentFile); err != nil {
		return nil, false, fmt.Errorf("can't decode torrent file %v for torrent %v with error %v", transferStruct.TorrentFilePath, key, err)
	}

	if !strings.HasPrefix(key, "magnet:?") {
		if err = helpers.DecodeTorrentFile(transferStruct.TorrentFilePath, &transferStruct.TorrentFileRaw); err != nil {
			return nil, false, fmt.Errorf("can't decode torrent file %v for torrent %v with error %v", transferStruct.TorrentFilePath, key, err)
		}
	} else {
		transferStruct.Magnet = true
		transferStruct.TorrentFile = &torrentStructures.Torrent{
			Info: &torrentStructures.TorrentInfo{},
		}
	}

	transferStruct.HandleStructures()
	if !ShouldIncludeDelugeTorrent(transferStruct.IsPrivate(), transferStruct.Opts.DelugePrivacy) {
		return nil, true, nil
	}

	torrentID := transferStruct.GetHash()
	destDir := transferStruct.Opts.DelugeStateDir
	if err := helpers.CopyFile(transferStruct.TorrentFilePath, filepath.Join(destDir, torrentID+".torrent")); err != nil {
		return nil, false, fmt.Errorf("can't create Deluge torrent file %v", filepath.Join(destDir, torrentID+".torrent"))
	}

	delugeResume := BuildDelugeFastresume(transferStruct)
	resumeData, err := bencode.EncodeBytes(delugeResume)
	if err != nil {
		return nil, false, fmt.Errorf("can't encode Deluge fastresume for torrent %v with error %v", key, err)
	}

	magnetURI := ""
	if strings.HasPrefix(key, "magnet:?") {
		magnetURI = key
	}

	state := BuildDelugeTorrentState(transferStruct, torrentID, queue, magnetURI)
	return &delugeItemResult{
		TorrentID:  torrentID,
		State:      state,
		ResumeData: resumeData,
		IsPrivate:  transferStruct.IsPrivate(),
	}, false, nil
}
