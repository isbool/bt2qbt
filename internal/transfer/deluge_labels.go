package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rumanzo/bt2qbt/internal/options"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/utorrentStructs"
)

func NormalizeDelugeLabel(label string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	if label == "" {
		return ""
	}
	var out strings.Builder
	lastUnderscore := false
	for _, r := range label {
		isValid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.'
		if isValid {
			out.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			out.WriteRune('_')
			lastUnderscore = true
		}
	}
	result := strings.Trim(out.String(), "_")
	return result
}

func PickDelugeLabel(resumeItem *utorrentStructs.ResumeItem, opts *options.Opts) (string, []string) {
	var all []string
	seen := map[string]struct{}{}

	addLabel := func(label string) {
		label = NormalizeDelugeLabel(helpers.HandleCesu8(label))
		if label == "" {
			return
		}
		if _, ok := seen[label]; ok {
			return
		}
		seen[label] = struct{}{}
		all = append(all, label)
	}

	if !opts.WithoutLabels && resumeItem.Label != "" {
		addLabel(resumeItem.Label)
	}

	if !opts.WithoutTags {
		for _, tag := range resumeItem.Labels {
			if tag == "" {
				continue
			}
			addLabel(tag)
		}
	}

	var primary string
	if !opts.WithoutLabels && resumeItem.Label != "" {
		primary = NormalizeDelugeLabel(helpers.HandleCesu8(resumeItem.Label))
	} else if !opts.WithoutTags && len(resumeItem.Labels) > 0 {
		primary = NormalizeDelugeLabel(helpers.HandleCesu8(resumeItem.Labels[0]))
	}

	if primary == "" {
		return "", all
	}
	return primary, all
}

func defaultDelugeLabelOptions() map[string]interface{} {
	return map[string]interface{}{
		"apply_max":             false,
		"max_download_speed":    -1,
		"max_upload_speed":      -1,
		"max_connections":       -1,
		"max_upload_slots":      -1,
		"prioritize_first_last": false,
		"apply_queue":           false,
		"is_auto_managed":       false,
		"stop_at_ratio":         false,
		"stop_ratio":            2.0,
		"remove_at_ratio":       false,
		"apply_move_completed":  false,
		"move_completed":        false,
		"move_completed_path":   "",
		"auto_add":              false,
		"auto_add_trackers":     []interface{}{},
	}
}

func readDelugeLabelConfig(path string) (map[string]interface{}, map[string]interface{}, error) {
	version := map[string]interface{}{
		"format": 1,
		"file":   1,
	}
	config := map[string]interface{}{
		"labels":         map[string]interface{}{},
		"torrent_labels": map[string]interface{}{},
	}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return version, config, nil
	}
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	var first map[string]interface{}
	if err := dec.Decode(&first); err != nil {
		return nil, nil, err
	}
	var second map[string]interface{}
	if err := dec.Decode(&second); err != nil {
		if errors.Is(err, io.EOF) {
			return version, first, nil
		}
		return nil, nil, err
	}
	return first, second, nil
}

func writeDelugeLabelConfig(path string, version map[string]interface{}, config map[string]interface{}) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Rename(path, path+".bak"); err != nil {
			return err
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(version); err != nil {
		return err
	}
	if err := enc.Encode(config); err != nil {
		return err
	}
	return nil
}

func ProcessDelugeLabels(opts *options.Opts, labelIDs []string, torrentLabels map[string]string) error {
	if len(labelIDs) == 0 && len(torrentLabels) == 0 {
		return nil
	}

	version, config, err := readDelugeLabelConfig(opts.DelugeLabels)
	if err != nil {
		return fmt.Errorf("unexpected error while open label.conf: %w", err)
	}

	labelsRaw, ok := config["labels"].(map[string]interface{})
	if !ok || labelsRaw == nil {
		labelsRaw = map[string]interface{}{}
	}
	torrentLabelsRaw, ok := config["torrent_labels"].(map[string]interface{})
	if !ok || torrentLabelsRaw == nil {
		torrentLabelsRaw = map[string]interface{}{}
	}

	for _, label := range labelIDs {
		if _, ok := labelsRaw[label]; !ok {
			labelsRaw[label] = defaultDelugeLabelOptions()
		}
	}

	for torrentID, label := range torrentLabels {
		if label == "" {
			continue
		}
		torrentLabelsRaw[torrentID] = label
		if _, ok := labelsRaw[label]; !ok {
			labelsRaw[label] = defaultDelugeLabelOptions()
		}
	}

	config["labels"] = labelsRaw
	config["torrent_labels"] = torrentLabelsRaw

	if err := writeDelugeLabelConfig(opts.DelugeLabels, version, config); err != nil {
		return fmt.Errorf("can't write label.conf: %w", err)
	}
	return nil
}
