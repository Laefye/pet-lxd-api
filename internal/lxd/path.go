package lxd

import (
	"net/url"
	"strings"
)

type Path struct {
	Version  string
	Segments []string
	Query    Query
}

func (p Path) String() string {
	path := "/" + p.Version
	if len(p.Segments) > 0 {
		path += "/" + strings.Join(p.Segments, "/")
	}
	path += p.Query.String()
	return path
}

func (p Path) Join(segment string) Path {
	return Path{
		Version:  p.Version,
		Segments: append(p.Segments, segment),
		Query:    p.Query,
	}
}

func ParsePath(rawPath string) Path {
	u, err := url.Parse(rawPath)
	if err != nil {
		return Path{}
	}
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	version := ""
	if len(segments) > 0 {
		version = segments[0]
		segments = segments[1:]
	}
	query := ParseQuery(u.RawQuery)
	return Path{
		Version:  version,
		Segments: segments,
		Query:    query,
	}
}
