package errors

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	data := map[string]string{"message": "test"}
	resp := Success(data)

	assert.True(t, resp.Success)
	assert.Equal(t, data, resp.Data)
	assert.Nil(t, resp.Error)
	assert.Nil(t, resp.Meta)
}

func TestSuccessWithMeta(t *testing.T) {
	data := []string{"item1", "item2"}
	meta := &Meta{
		Page:    1,
		PerPage: 20,
		Total:   100,
	}
	resp := SuccessWithMeta(data, meta)

	assert.True(t, resp.Success)
	assert.Equal(t, data, resp.Data)
	assert.Nil(t, resp.Error)
	assert.Equal(t, meta, resp.Meta)
}

func TestResponseStructure(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantData interface{}
		wantErr  bool
		wantMeta bool
	}{
		{
			name: "success response with data",
			response: Response{
				Success: true,
				Data:    map[string]int{"id": 1},
			},
			wantData: map[string]int{"id": 1},
			wantErr:  false,
			wantMeta: false,
		},
		{
			name: "error response",
			response: Response{
				Success: false,
				Error: &ErrorInfo{
					Code:      "TEST_ERROR",
					Message:   "test error",
					Timestamp: time.Now(),
				},
			},
			wantData: nil,
			wantErr:  true,
			wantMeta: false,
		},
		{
			name: "success with metadata",
			response: Response{
				Success: true,
				Data:    []int{1, 2, 3},
				Meta: &Meta{
					Page:  1,
					Total: 100,
				},
			},
			wantData: []int{1, 2, 3},
			wantErr:  false,
			wantMeta: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.NotNil(t, tt.response.Error)
				assert.Nil(t, tt.response.Data)
			} else {
				assert.Nil(t, tt.response.Error)
				assert.Equal(t, tt.wantData, tt.response.Data)
			}

			if tt.wantMeta {
				assert.NotNil(t, tt.response.Meta)
			} else {
				assert.Nil(t, tt.response.Meta)
			}
		})
	}
}

func TestResponseJSONSerialization(t *testing.T) {
	t.Run("success response serializes correctly", func(t *testing.T) {
		resp := Success(map[string]string{"message": "hello"})
		data, err := json.Marshal(resp)
		assert.NoError(t, err)

		var decoded Response
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)

		assert.True(t, decoded.Success)
		assert.NotNil(t, decoded.Data)
	})

	t.Run("error response serializes correctly", func(t *testing.T) {
		resp := Response{
			Success: false,
			Error: &ErrorInfo{
				Code:      "TEST_ERROR",
				Message:   "test message",
				Timestamp: time.Now(),
			},
		}
		data, err := json.Marshal(resp)
		assert.NoError(t, err)

		var decoded Response
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)

		assert.False(t, decoded.Success)
		assert.NotNil(t, decoded.Error)
	})
}
