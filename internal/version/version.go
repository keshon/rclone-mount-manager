package version

import "runtime"

const (
	AppName        = "Rclone Mount Manager"
	AppFullName    = "Rclone Mount Manager - GUI tool to manage Rclone mounts on Windows"
	AppDescription = "GUI tool to manage Rclone mounts on Windows"
	AppRepo        = "github.com/keshon/rclone-mount-manager"
	AppAuthor      = "Innokentiy Sokolov"
)

var (
	BuildDate = ""
	GoVersion = runtime.Version()
)
