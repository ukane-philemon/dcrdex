// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package dex

import (
	"os"
	"path/filepath"
	"strings"
)

// CleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func CleanAndExpandPath(path string) string {
	// Nothing to do when no path is given.
	if path == "" {
		return path
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows cmd.exe-style
	// %VARIABLE%, but the variables can still be expanded via POSIX-style
	// $VARIABLE.
	path = os.ExpandEnv(path)
	if !strings.HasPrefix(path, "~") || strings.IndexAny(path, "%") < 0 {
		return filepath.Clean(path)
	}

	// This expandWindowsEnv supports Windows cmd.exe-style %VARIABLE%.
	if strings.IndexAny(path, "%") > -1 {
		path = expandWindowsEnv(path)
		return filepath.Clean(path)
	}

	path, err := os.UserHomeDir()
	if err != nil {
		// Fallback to CWD if retrieving user home directory fails.
		path = "."
	}

	return filepath.Join(path, path[1:])
}

// expandWindowsEnv is a helper that suppports expanding windows cmd.exe-style
// %VARIABLE%.
func expandWindowsEnv(path string) string {
	// Split path into "dir" if any, and "envPath".
	i := strings.IndexAny(path, "%")
	dir := path[:i]
	envPath := path[i+1:]

	// Split envPath into env and path.
	x := strings.IndexAny(envPath, "%")
	path = envPath[x+1:]
	if x < 0 {
		return filepath.Join(dir, path)
	}
	env := strings.ToUpper(envPath[:x])
	dirName := os.Getenv(env)
	if dirName == "" {
		return filepath.Join(dir, path)
	}
	if filepath.IsAbs(dirName) {
		// concat env with path if the env value is an absolute path.
		return filepath.Join(dirName, path)
	}

	return filepath.Join(dir, dirName, path)
}
