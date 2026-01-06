package transfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zeebo/bencode"
)

const delugeStateWriterScript = `#!/usr/bin/env python3
import argparse
import json
import os
import pickle
import sys
import types

TorrentState = type("TorrentState", (), {})
TorrentManagerState = type("TorrentManagerState", (), {})
TorrentState.__module__ = "deluge.core.torrentmanager"
TorrentManagerState.__module__ = "deluge.core.torrentmanager"

deluge_mod = types.ModuleType("deluge")
core_mod = types.ModuleType("deluge.core")
torrentmanager_mod = types.ModuleType("deluge.core.torrentmanager")
torrentmanager_mod.TorrentState = TorrentState
torrentmanager_mod.TorrentManagerState = TorrentManagerState
sys.modules["deluge"] = deluge_mod
sys.modules["deluge.core"] = core_mod
sys.modules["deluge.core.torrentmanager"] = torrentmanager_mod

class DelugeUnpickler(pickle.Unpickler):
    def find_class(self, module, name):
        if module == "deluge.core.torrentmanager":
            if name == "TorrentState":
                return TorrentState
            if name == "TorrentManagerState":
                return TorrentManagerState
        return pickle.Unpickler.find_class(self, module, name)

def load_state(path):
    if not os.path.exists(path):
        state = TorrentManagerState()
        state.torrents = []
        return state
    try:
        with open(path, "rb") as f:
            return DelugeUnpickler(f).load()
    except Exception:
        state = TorrentManagerState()
        state.torrents = []
        return state

def write_state(path, state):
    tmp = path + ".tmp"
    with open(tmp, "wb") as f:
        pickle.dump(state, f, protocol=2)
    if os.path.exists(path):
        try:
            os.replace(path, path + ".bak")
        except OSError:
            pass
    os.replace(tmp, path)

def merge_state(state, new_items):
    torrents = getattr(state, "torrents", None)
    if torrents is None:
        state.torrents = []
        torrents = state.torrents
    index = {}
    for i, t in enumerate(torrents):
        tid = getattr(t, "torrent_id", None)
        if tid:
            index[tid] = i
    for item in new_items:
        tid = item.get("torrent_id")
        if not tid:
            continue
        if tid in index:
            t = torrents[index[tid]]
        else:
            t = TorrentState()
            torrents.append(t)
            index[tid] = len(torrents) - 1
        for k, v in item.items():
            setattr(t, k, v)
    state.torrents = torrents

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--state", required=True)
    parser.add_argument("--input", required=True)
    args = parser.parse_args()
    with open(args.input, "r", encoding="utf-8") as f:
        payload = json.load(f)
    items = payload.get("torrents", [])
    state = load_state(args.state)
    merge_state(state, items)
    write_state(args.state, state)

if __name__ == "__main__":
    main()
`

type pythonCommand struct {
	Path string
	Args []string
}

func findPython() (pythonCommand, error) {
	candidates := []pythonCommand{}
	if env := strings.TrimSpace(os.Getenv("BT2QBT_PYTHON")); env != "" {
		candidates = append(candidates, pythonCommand{Path: env})
	}
	candidates = append(candidates, pythonCommand{Path: "python3"})
	candidates = append(candidates, pythonCommand{Path: "python"})
	candidates = append(candidates, pythonCommand{Path: "py", Args: []string{"-3"}})

	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate.Path)
		if err != nil {
			continue
		}
		if probePython(path, candidate.Args) {
			return pythonCommand{Path: path, Args: candidate.Args}, nil
		}
	}

	return pythonCommand{}, fmt.Errorf("python not found in PATH")
}

func probePython(path string, baseArgs []string) bool {
	args := append(append([]string{}, baseArgs...), "--version")
	cmd := exec.Command(path, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.Contains(output.String(), "Python")
}

func WriteDelugeState(stateDir string, torrents []DelugeTorrentState) error {
	if len(torrents) == 0 {
		return nil
	}

	python, err := findPython()
	if err != nil {
		return err
	}

	payload := struct {
		Torrents []DelugeTorrentState `json:"torrents"`
	}{
		Torrents: torrents,
	}
	inputData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	scriptFile, err := os.CreateTemp("", "bt2qbt_deluge_state_*.py")
	if err != nil {
		return err
	}
	scriptPath := scriptFile.Name()
	if _, err := scriptFile.Write([]byte(delugeStateWriterScript)); err != nil {
		scriptFile.Close()
		return err
	}
	if err := scriptFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(scriptPath, 0700); err != nil {
		return err
	}
	defer os.Remove(scriptPath)

	inputFile, err := os.CreateTemp("", "bt2qbt_deluge_state_*.json")
	if err != nil {
		return err
	}
	inputPath := inputFile.Name()
	if _, err := inputFile.Write(inputData); err != nil {
		inputFile.Close()
		return err
	}
	if err := inputFile.Close(); err != nil {
		return err
	}
	defer os.Remove(inputPath)

	statePath := filepath.Join(stateDir, "torrents.state")
	cmdArgs := append(append([]string{}, python.Args...), scriptPath, "--state", statePath, "--input", inputPath)
	cmd := exec.Command(python.Path, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("deluge state writer failed: %v: %s", err, stderr.String())
		}
		return fmt.Errorf("deluge state writer failed: %w", err)
	}
	return nil
}

func WriteDelugeFastresume(stateDir string, newData map[string][]byte) error {
	if len(newData) == 0 {
		return nil
	}
	path := filepath.Join(stateDir, "torrents.fastresume")
	merged := map[string][]byte{}
	if data, err := os.ReadFile(path); err == nil {
		if err := bencode.DecodeBytes(data, &merged); err != nil {
			return fmt.Errorf("can't decode torrents.fastresume: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	for key, value := range newData {
		merged[key] = value
	}

	out, err := bencode.EncodeBytes(merged)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		_ = os.Rename(path, path+".bak")
	}
	return os.WriteFile(path, out, 0600)
}
