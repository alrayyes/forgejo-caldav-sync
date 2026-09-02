package main

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

// configFileRelPath is the config file's location relative to whichever
// XDG config directory it resolves under — $XDG_CONFIG_HOME/<this> by
// default, falling back to the spec's documented default when that's
// unset.
const configFileRelPath = "forgejo-caldav-sync/config.yaml"

// defaultConfigPath returns where a config file would live if nothing on
// the command line names one — creating the directory (not the file) if
// it doesn't exist yet, the same way xdg.ConfigFile documents.
func defaultConfigPath() (string, error) {
	path, err := xdg.ConfigFile(configFileRelPath)
	if err != nil {
		return "", fmt.Errorf("resolving default config path: %w", err)
	}

	return path, nil
}

// readConfigFile reads path into v, or the XDG default location when path
// is empty, and reports whether it actually found and read a file. A
// missing file at the auto-resolved XDG default is not an error — every
// setting can come from flags and the environment alone — but a file
// explicitly named with --config has to exist.
func readConfigFile(v *viper.Viper, path string) (bool, error) {
	explicit := path != ""
	if !explicit {
		defaultPath, err := defaultConfigPath()
		if err != nil {
			return false, err
		}
		path = defaultPath
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) && !explicit {
			return false, nil
		}

		return false, fmt.Errorf("config file %s: %w", path, err)
	}

	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return false, fmt.Errorf("reading config file %s: %w", path, err)
	}

	return true, nil
}
