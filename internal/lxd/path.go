package lxd

import (
	"net/url"
	"strings"
)

type Path struct {
	Segments []string
	Query    url.Values
}

func escapeSegments(segments []string) []string {
	escapedSegments := make([]string, len(segments))
	for i, segment := range segments {
		escapedSegments[i] = url.PathEscape(segment)
	}
	return escapedSegments
}

func (p Path) String() string {
	path := ""
	escapedSegments := escapeSegments(p.Segments)
	if len(escapedSegments) > 0 {
		path += "/" + strings.Join(escapedSegments, "/")
	}
	if len(p.Query) > 0 {
		path += "?" + p.Query.Encode()
	}
	return path
}

func (p Path) Join(segment string) Path {
	return Path{
		Segments: append(p.Segments, segment),
		Query:    p.Query,
	}
}

func (p Path) WithProject(project string) Path {
	return p.WithQuery("project", project)
}

func (p Path) WithSecret(secret string) Path {
	return p.WithQuery("secret", secret)
}

func (p Path) WithQuery(key, value string) Path {
	if p.Query == nil {
		p.Query = url.Values{}
	}
	p.Query.Set(key, value)
	return p
}

func (p Path) Last(i int) string {
	return p.Segments[len(p.Segments)-i-1]
}

func unescapeSegments(path string) ([]string, error) {
	escapedSegments := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, len(escapedSegments))
	for i, segment := range escapedSegments {
		unescapedSegment, err := url.PathUnescape(segment)
		if err != nil {
			return nil, err
		}
		segments[i] = unescapedSegment
	}
	return segments, nil
}

func ParsePath(rawPath string) (*Path, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return nil, err
	}
	segments, err := unescapeSegments(u.Path)
	if err != nil {
		return nil, err
	}
	query := url.Values(u.Query())
	return &Path{
		Segments: segments,
		Query:    query,
	}, nil
}

func MustParsePath(rawPath string) Path {
	path, err := ParsePath(rawPath)
	if err != nil {
		panic(err)
	}
	return *path
}
