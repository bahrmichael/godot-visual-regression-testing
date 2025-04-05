package lib

import "strings"

func WithFolderSuffix(dir string) string {
	if strings.HasSuffix(dir, "/") {
		return dir
	}
	return dir + "/"
}
