package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTaskParams(t *testing.T) {
	svc := &TaskService{}

	tests := []struct {
		name    string
		req     TaskRequest
		wantErr bool
	}{
		{
			name:    "Valid Exec",
			req:     TaskRequest{Type: "exec_command", Params: map[string]interface{}{"command": "echo"}},
			wantErr: false,
		},
		{
			name:    "Invalid Exec Missing Command",
			req:     TaskRequest{Type: "exec_command", Params: map[string]interface{}{"args": []string{"hi"}}},
			wantErr: true,
		},
		{
			name:    "Valid Service Action",
			req:     TaskRequest{Type: "service_action", Params: map[string]interface{}{"service": "nginx", "action": "restart"}},
			wantErr: false,
		},
		{
			name:    "Invalid Service Action",
			req:     TaskRequest{Type: "service_action", Params: map[string]interface{}{"service": "nginx", "action": "dance"}},
			wantErr: true,
		},
		{
			name:    "Unknown Type Passes",
			req:     TaskRequest{Type: "unknown_magic", Params: nil},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateTaskParams(tt.req.Type, tt.req.Params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
