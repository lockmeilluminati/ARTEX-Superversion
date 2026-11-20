package server

import "testing"

func TestValidateRegistryReviewRequest(t *testing.T) {
	tests := []struct {
		name string
		req registryReviewRequest
		wantErr bool
	}{
		{name: "approve", req: registryReviewRequest{Status: "approved"}},
		{name: "reset", req: registryReviewRequest{Status: "pending"}},
		{name: "reject with reason", req: registryReviewRequest{Status: "rejected", Note: "License review required"}},
		{name: "reject without reason", req: registryReviewRequest{Status: "rejected"}, wantErr: true},
		{name: "unknown state", req: registryReviewRequest{Status: "enabled"}, wantErr: true},
		{name: "oversized note", req: registryReviewRequest{Status: "approved", Note: string(make([]byte, 2001))}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegistryReviewRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateRegistryReviewRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
