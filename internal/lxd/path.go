package lxd

import (
	"net/url"
	"strings"
)

type resourcePath struct {
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

func (p resourcePath) String() string {
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

func (p resourcePath) join(segment string) resourcePath {
	return resourcePath{
		Segments: append(p.Segments, segment),
		Query:    p.Query,
	}
}

func (p resourcePath) withQuery(key, value string) resourcePath {
	if p.Query == nil {
		p.Query = url.Values{}
	}
	p.Query.Set(key, value)
	return p
}

func (p resourcePath) withoutQuery(key string) resourcePath {
	if p.Query == nil {
		return p
	}
	p.Query.Del(key)
	return p
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

func ParsePath(rawPath string) (*resourcePath, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return nil, err
	}
	segments, err := unescapeSegments(u.Path)
	if err != nil {
		return nil, err
	}
	query := url.Values(u.Query())
	return &resourcePath{
		Segments: segments,
		Query:    query,
	}, nil
}

func MustParsePath(rawPath string) resourcePath {
	path, err := ParsePath(rawPath)
	if err != nil {
		panic(err)
	}
	return *path
}
