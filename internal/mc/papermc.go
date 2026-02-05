package mc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type PaperMCApi struct {
	BaseUrl string
	Client  http.Client
}

type PaperMCBuild struct {
	ID        int                    `json:"id"`
	Time      string                 `json:"time"`
	Channel   string                 `json:"channel"`
	Commits   []PaperMCCommit        `json:"commits"`
	Downloads map[string]PaperMCFile `json:"downloads"`
}

type PaperMCCommit struct {
	SHA     string `json:"sha"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

type PaperMCFile struct {
	Name      string           `json:"name"`
	Checksums PaperMCChecksums `json:"checksums"`
	Size      int64            `json:"size"`
	URL       string           `json:"url"`
}

type PaperMCChecksums struct {
	SHA256 string `json:"sha256"`
}

func (api *PaperMCApi) GetBuilds(ctx context.Context, project string, version string) ([]*PaperMCBuild, error) {
	endpoint := api.BaseUrl + "/v3/projects/" + url.PathEscape(project) + "/versions/" + url.PathEscape(version) + "/builds"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := api.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out []*PaperMCBuild
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
