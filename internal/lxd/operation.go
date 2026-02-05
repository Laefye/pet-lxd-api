package lxd

import (
	"context"
	"net/http"
)

func (r *Rest) Wait(ctx context.Context, id string) (*RestResponse, error) {
	return r.Request(ctx, http.MethodGet, id+"/wait", nil)
}
