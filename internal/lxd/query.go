package lxd

import "net/url"

type Query struct {
	Project string
}

func (q Query) String() string {
	values := url.Values{}
	if q.Project != "" {
		values.Set("project", q.Project)
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func ParseQuery(rawQuery string) Query {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return Query{}
	}
	return Query{
		Project: values.Get("project"),
	}
}
