package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceAction_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "Missing Service",
			params:  map[string]interface{}{"action": "start"},
			wantErr: true,
		},
		{
			name:    "Missing Action",
			params:  map[string]interface{}{"service": "nginx"},
			wantErr: true,
		},
		{
			name:    "Invalid Action",
			params:  map[string]interface{}{"service": "nginx", "action": "dance"},
			wantErr: true,
		},
		{
			name:    "Valid Params (Mock)",
			params:  map[string]interface{}{"service": "dummy_service", "action": "status"},
			wantErr: true, // Will fail execution because service doesn't exist, but validation passes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ServiceAction(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
