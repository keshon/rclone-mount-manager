package rclone

type Remote struct {
	Name      string
	Type      string
	Upstreams []string // only for union remotes
	Config    map[string]string
}

type Mount struct {
	Name       string
	MountPoint string
	CacheMode  string
	CacheSize  string
	CacheAge   string
	CacheDir   string
	Status     string
	LogPath    string
}
