package options

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Opts struct {
	Client         string   `long:"client" description:"Target client: qbt (default) or deluge"`
	BitDir         string   `short:"s" long:"source" description:"Source directory that contains resume.dat and torrents files"`
	QBitDir        string   `short:"d" long:"destination" description:"Destination directory BT_backup (as default)"`
	PrivateQBitDir string   `long:"destination-private" description:"Destination directory BT_backup for private torrents (optional)"`
	Categories     string   `short:"c" long:"categories" description:"Path to qBittorrent categories.json file (for write tags)"`
	DelugeConfig   string   `long:"deluge-config" description:"Deluge config directory (default depends on OS)"`
	DelugeStateDir string   `long:"deluge-state" description:"Deluge state directory (contains torrents.state/fastresume and .torrent files)"`
	DelugeLabels   string   `long:"deluge-labels" description:"Path to Deluge label.conf file (Label plugin)"`
	DelugePrivacy  string   `long:"deluge-privacy" description:"Deluge export filter: all (default), public, or private"`
	WithoutLabels  bool     `long:"without-labels" description:"Do not export/import labels"`
	WithoutTags    bool     `long:"without-tags" description:"Do not export/import tags"`
	SearchPaths    []string `short:"t" long:"search" description:"Additional search path for torrents files\n	Example: --search='/mnt/olddisk/savedtorrents' --search='/mnt/olddisk/workstorrents'"`
	Replaces       []string `short:"r" long:"replace" description:"Replace save paths. Important: you have to use single slashes in paths\n	Delimiter for from/to is comma - ,\n	Example: -r \"D:/films,/home/user/films\" -r \"D:/music,/home/user/music\"\n"`
	PathSeparator  string   `long:"sep" description:"Default path separator that will use in all paths. You may need use this flag if you migrating from windows to linux in some cases"`
	Stats          bool     `long:"stats" description:"Show public/private/failed counts only (no conversion)"`
	Version        bool     `short:"v" long:"version" description:"Show version"`
}

func PrepareOpts() *Opts {
	opts := &Opts{PathSeparator: string(os.PathSeparator)}
	opts.Client = "qbt"
	opts.DelugePrivacy = "all"
	switch OS := runtime.GOOS; OS {
	case "windows":
		opts.BitDir = filepath.Join(os.Getenv("APPDATA"), "uTorrent")
		opts.Categories = filepath.Join(os.Getenv("APPDATA"), "qBittorrent", "categories.json")
		opts.QBitDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "qBittorrent", "BT_backup")
		opts.DelugeConfig = filepath.Join(os.Getenv("APPDATA"), "deluge")
	case "linux":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		opts.BitDir = "/mnt/uTorrent/"
		opts.Categories = filepath.Join(usr.HomeDir, ".config", "qBittorrent", "categories.json")
		opts.QBitDir = filepath.Join(usr.HomeDir, ".local", "share", "data", "qBittorrent", "BT_backup")
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			opts.DelugeConfig = filepath.Join(xdg, "deluge")
		} else {
			opts.DelugeConfig = filepath.Join(usr.HomeDir, ".config", "deluge")
		}
	case "darwin":
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		opts.BitDir = filepath.Join(usr.HomeDir, "Library", "Application Support", "uTorrent")
		opts.Categories = filepath.Join(usr.HomeDir, ".config", "qBittorrent", "categories.json")
		opts.QBitDir = filepath.Join(usr.HomeDir, "Library", "Application Support", "QBittorrent", "BT_backup")
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			opts.DelugeConfig = filepath.Join(xdg, "deluge")
		} else {
			opts.DelugeConfig = filepath.Join(usr.HomeDir, ".config", "deluge")
		}
	}
	return opts
}

func ParseOpts(opts *Opts) *Opts {
	if _, err := flags.Parse(opts); err != nil { // https://godoc.org/github.com/jessevdk/go-flags#ErrorType
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Println(err)
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
	}
	return opts
}

// HandleOpts used for enrichment opts after first creation
func HandleOpts(opts *Opts) {
	opts.Client = strings.ToLower(strings.TrimSpace(opts.Client))
	if opts.Client == "" {
		opts.Client = "qbt"
	}
	if opts.Client == "qbittorrent" {
		opts.Client = "qbt"
	}
	opts.DelugePrivacy = strings.ToLower(strings.TrimSpace(opts.DelugePrivacy))
	if opts.DelugePrivacy == "" {
		opts.DelugePrivacy = "all"
	}
	opts.SearchPaths = append(opts.SearchPaths, opts.BitDir)

	if opts.Client == "qbt" {
		qbtDir := fileHelpers.Normalize(opts.QBitDir, `/`)
		if strings.Contains(qbtDir, `profile/qBittorrent/data/BT_backup`) {
			qbtRootDir, _ := strings.CutSuffix(qbtDir, `data/BT_backup`)

			// check that user not define categories
			refOpts := PrepareOpts()
			if refOpts.Categories == opts.Categories {
				opts.Categories = fileHelpers.Join([]string{qbtRootDir, `config/categories.json`}, opts.PathSeparator)
			}
		}
	}

	if opts.DelugeConfig != "" {
		if opts.DelugeStateDir == "" {
			opts.DelugeStateDir = filepath.Join(opts.DelugeConfig, "state")
		}
		if opts.DelugeLabels == "" {
			opts.DelugeLabels = filepath.Join(opts.DelugeConfig, "label.conf")
		}
	}
}

func OptsCheck(opts *Opts) error {
	if opts.Client == "" {
		opts.Client = "qbt"
	}
	if opts.Client != "qbt" && opts.Client != "deluge" {
		return fmt.Errorf("unknown client: %v (use qbt or deluge)", opts.Client)
	}
	if opts.Client == "deluge" {
		switch opts.DelugePrivacy {
		case "all", "public", "private":
		default:
			return fmt.Errorf("unknown deluge privacy: %v (use all, public, or private)", opts.DelugePrivacy)
		}
	}
	if len(opts.Replaces) != 0 {
		for _, str := range opts.Replaces {
			patterns := strings.Split(str, ",")
			if len(patterns) != 2 {
				return fmt.Errorf("bad replace pattern")
			}
		}
	}

	if _, err := os.Stat(opts.BitDir); os.IsNotExist(err) {
		return fmt.Errorf("can't find uTorrent\\Bittorrent folder")
	}

	if !opts.Stats {
		if opts.Client == "deluge" {
			if _, err := os.Stat(opts.DelugeStateDir); os.IsNotExist(err) {
				return fmt.Errorf("can't find Deluge state folder")
			}
		} else {
			if _, err := os.Stat(opts.QBitDir); os.IsNotExist(err) {
				return fmt.Errorf("can't find qBittorrent folder")
			}
			if opts.PrivateQBitDir != "" {
				if _, err := os.Stat(opts.PrivateQBitDir); os.IsNotExist(err) {
					return fmt.Errorf("can't find qBittorrent folder for private torrents")
				}
			}
		}
	}

	if runtime.GOOS == "linux" {
		if opts.SearchPaths == nil {
			return fmt.Errorf("on linux systems you must define search path for torrents")
		}
	}
	return nil
}

func MakeOpts() *Opts {
	opts := PrepareOpts()
	ParseOpts(opts)
	HandleOpts(opts)
	err := OptsCheck(opts)
	if err != nil {
		log.Println(err)
		time.Sleep(time.Duration(30) * time.Second)
		os.Exit(1)
	}
	return opts
}
