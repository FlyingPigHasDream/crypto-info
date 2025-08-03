package main

import (
	"context"
	v1 "crypto-info/kitex_gen/crypto/v1"
)

// HealthServiceImpl implements the last service interface defined in the IDL.
type HealthServiceImpl struct{}

// Check implements the HealthServiceImpl interface.
func (s *HealthServiceImpl) Check(ctx context.Context, req *v1.HealthCheckRequest) (resp *v1.HealthCheckResponse, err error) {
	// TODO: Your code here...
	return
}
