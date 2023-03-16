package version

import (
	"fmt"
	"runtime"
)

//var version control
var (
	Version   = "1.68.2"
	GitCommit string
	BuildTime string
	// GoVersion system go version
	GoVersion = runtime.Version()
	// Platform info
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

//GetVersion 获取版本信息
func GetVersion() string {
	if GitCommit != "" {
		return Version + "-" + GitCommit
	}
	return Version
}
