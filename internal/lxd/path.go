package lxd

import (
	"net/url"
	"strings"
)

type Path struct {
	Segments []string
	Query    url.Values
}

func (p Path) String() string {
	path := ""
	if len(p.Segments) > 0 {
		path += "/" + strings.Join(p.Segments, "/")
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

func ParsePath(rawPath string) Path {
	u, err := url.Parse(rawPath)
	if err != nil {
		return Path{}
	}
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	query := url.Values(u.Query())
	return Path{
		Segments: segments,
		Query:    query,
	}
}
