package delugeStructures

// https://www.libtorrent.org/manual-ref.html
type DelugeFastresume struct {
	ActiveTime         int64          `bencode:"active_time"`
	AddedTime          int64          `bencode:"added_time"`
	Allocation         string         `bencode:"allocation"`
	ApplyIpFilter      int64          `bencode:"apply_ip_filter"`
	AutoManaged        int64          `bencode:"auto_managed"`
	BannedPeers        []byte         `bencode:"banned_peers"`
	BannedPeers6       []byte         `bencode:"banned_peers6"`
	CompletedTime      int64          `bencode:"completed_time"`
	DisableDht         int64          `bencode:"disable_dht"`
	DisableLsd         int64          `bencode:"disable_lsd"`
	DisablePex         int64          `bencode:"disable_pex"`
	DownloadRateLimit  int64          `bencode:"download_rate_limit"`
	FileFormat         string         `bencode:"file-format"`
	FilePriority       []int64        `bencode:"file_priority"`
	FileVersion        int64          `bencode:"file-version"`
	FinishedTime       int64          `bencode:"finished_time"`
	HttpSeeds          []string       `bencode:"httpseeds"`
	Info               interface{}    `bencode:"info,omitempty"`
	InfoHash           string         `bencode:"info-hash"`
	InfoHash2          string         `bencode:"info-hash2"`
	LastDownload       int64          `bencode:"last_download"`
	LastSeenComplete   int64          `bencode:"last_seen_complete"`
	LastUpload         int64          `bencode:"last_upload"`
	LibTorrentVersion  string         `bencode:"libtorrent-version"`
	MappedFiles        []string       `bencode:"mapped_files,omitempty"`
	MaxConnections     int64          `bencode:"max_connections"`
	MaxUploads         int64          `bencode:"max_uploads"`
	Name               string         `bencode:"name"`
	NumComplete        int64          `bencode:"num_complete"`
	NumDownloaded      int64          `bencode:"num_downloaded"`
	NumIncomplete      int64          `bencode:"num_incomplete"`
	Paused             int64          `bencode:"paused"`
	Peers              string         `bencode:"peers"`
	Peers6             string         `bencode:"peers6"`
	Pieces             []byte         `bencode:"pieces"`
	SavePath           string         `bencode:"save_path"`
	SeedMode           int64          `bencode:"seed_mode"`
	SeedingTime        int64          `bencode:"seeding_time"`
	SequentialDownload int64          `bencode:"sequential_download"`
	ShareMode          int64          `bencode:"share_mode"`
	StopWhenReady      int64          `bencode:"stop_when_ready"`
	SuperSeeding       int64          `bencode:"super_seeding"`
	TotalDownloaded    int64          `bencode:"total_downloaded"`
	TotalUploaded      int64          `bencode:"total_uploaded"`
	Trackers           [][]string     `bencode:"trackers"`
	Unfinished         *[]interface{} `bencode:"unfinished,omitempty"`
	UploadMode         int64          `bencode:"upload_mode"`
	UploadRateLimit    int64          `bencode:"upload_rate_limit"`
	UrlList            []string       `bencode:"url-list"`
}
