package spotify

import (
	"net/url"
	"path"
	"strings"
)

func extractID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	if strings.HasPrefix(s, "spotify:") {
		parts := strings.Split(s, ":")
		if len(parts) >= 3 {
			return parts[len(parts)-1]
		}
	}

	u, err := url.Parse(s)
	if err == nil {
		p := strings.Trim(u.Path, "/")
		parts := strings.Split(p, "/")
		if len(parts) >= 2 {
			return parts[1]
		}
		return path.Base(u.Path)
	}

	return ""
}
