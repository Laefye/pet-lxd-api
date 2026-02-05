package lxd

import (
	"context"
	"fmt"
)

func (r *Rest) Wait(ctx context.Context, id string) (*RestResponse, error) {
	return r.Get(ctx, fmt.Sprintf("/1.0/operations/%s/wait", id))
}
