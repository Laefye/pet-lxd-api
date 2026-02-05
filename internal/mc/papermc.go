package mc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type PaperMCApi struct {
	BaseUrl string
	Client  *http.Client
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

const PaperMCBaseUrl = "https://fill.papermc.io"

func parsePaperBuilds(r io.Reader) ([]*PaperMCBuild, error) {
	var out []*PaperMCBuild
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (api *PaperMCApi) requestBuilds(ctx context.Context, project string, version string) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s/v3/projects/%s/versions/%s/builds", api.BaseUrl, url.PathEscape(project), url.PathEscape(version))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	return api.Client.Do(request)
}

func (api *PaperMCApi) GetBuilds(ctx context.Context, project string, version string) ([]*PaperMCBuild, error) {
	resp, err := api.requestBuilds(ctx, project, version)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	builds, err := parsePaperBuilds(resp.Body)
	if err != nil {
		return nil, err
	}
	return builds, nil
}
