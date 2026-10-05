package media

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ResolveInside(root, relative string) (string, error) {
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("absolute media path is forbidden")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.Abs(filepath.Join(rootAbs, filepath.Clean(relative)))
	if err != nil {
		return "", err
	}
	prefix := rootAbs + string(filepath.Separator)
	if resolved != rootAbs && !strings.HasPrefix(resolved, prefix) {
		return "", fmt.Errorf("media path escapes MEDIA_ROOT")
	}
	realRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve MEDIA_ROOT symlinks: %w", err)
	}
	realPath, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve media symlinks: %w", err)
	}
	realPrefix := realRoot + string(filepath.Separator)
	if realPath != realRoot && !strings.HasPrefix(realPath, realPrefix) {
		return "", fmt.Errorf("media symlink escapes MEDIA_ROOT")
	}
	return realPath, nil
}
