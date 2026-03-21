package darktable

import "github.com/photoview/photoview/api/log"

var DarktableCLI *Darktable

func init() {
	DarktableCLI = New()

	path, version, err := DarktableCLI.PathVersion()

	log.Info(nil, "darktable cli", "path", path, "version", version)

	if err != nil {
		log.Error(nil, "darktable cli", "error", err)
	}
}
