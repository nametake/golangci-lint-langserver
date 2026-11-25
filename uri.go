package main

import (
	"net/url"
	"path/filepath"
	"strings"
	"unicode"
)

func uriToPath(uri string) string {
	switch {
	case strings.HasPrefix(uri, "file:///"):
		uri = uri[len("file://"):]
	case strings.HasPrefix(uri, "file://"):
		uri = uri[len("file:/"):]
	}

	if path, err := url.PathUnescape(uri); err == nil {
		uri = path
	}

	if isWindowsDriveURIPath(uri) {
		uri = strings.ToUpper(string(uri[1])) + uri[2:]
	}

	return filepath.FromSlash(uri)
}

// pathToURI converts an absolute file system path into a file URI.
func pathToURI(path string) DocumentURI {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		// Windows drive paths like C:/foo need a leading slash: file:///C:/foo
		p = "/" + p
	}

	u := url.URL{Scheme: "file", Path: p}

	return DocumentURI(u.String())
}

func isWindowsDriveURIPath(uri string) bool {
	if len(uri) < 4 {
		return false
	}

	return uri[0] == '/' && unicode.IsLetter(rune(uri[1])) && uri[2] == ':'
}
