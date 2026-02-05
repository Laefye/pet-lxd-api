package lxd

import (
	"context"
	"fmt"
)

func (r *Rest) Wait(ctx context.Context, id string) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	if err := r.Get(ctx, fmt.Sprintf("/1.0/operations/%s/wait", id), &data); err != nil {
		return nil, err
	}
	return data, nil
}
