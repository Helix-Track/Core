package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalizationKey_Validate(t *testing.T) {
	tests := []struct {
		name    string
		key     LocalizationKey
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid key",
			key: LocalizationKey{
				Key:      "error.auth.invalid_token",
				Category: "error",
			},
			wantErr: false,
		},
		{
			name: "missing key",
			key: LocalizationKey{
				Category: "error",
			},
			wantErr: true,
			errMsg:  "key is required",
		},
		{
			name: "key too long",
			key: LocalizationKey{
				Key: string(make([]byte, 256)),
			},
			wantErr: true,
			errMsg:  "key must be 255 characters or less",
		},
		{
			name: "category too long",
			key: LocalizationKey{
				Key:      "test.key",
				Category: string(make([]byte, 101)),
			},
			wantErr: true,
			errMsg:  "category must be 100 characters or less",
		},
		{
			name: "context too long",
			key: LocalizationKey{
				Key:     "test.key",
				Context: string(make([]byte, 256)),
			},
			wantErr: true,
			errMsg:  "context must be 255 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.key.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLocalizationKey_BeforeCreate(t *testing.T) {
	key := LocalizationKey{
		Key:      "test.key",
		Category: "test",
	}

	key.BeforeCreate()

	assert.NotEmpty(t, key.ID)
	assert.NotZero(t, key.CreatedAt)
	assert.NotZero(t, key.ModifiedAt)
	assert.Equal(t, key.CreatedAt, key.ModifiedAt)
}

func TestLocalizationKey_BeforeUpdate(t *testing.T) {
	key := LocalizationKey{
		Key:        "test.key",
		CreatedAt:  1000,
		ModifiedAt: 1000,
	}

	key.BeforeUpdate()

	assert.NotZero(t, key.ModifiedAt)
	assert.Greater(t, key.ModifiedAt, key.CreatedAt)
}
