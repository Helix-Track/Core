package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalization_Validate(t *testing.T) {
	tests := []struct {
		name    string
		loc     Localization
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid localization",
			loc: Localization{
				KeyID:      "key-123",
				LanguageID: "lang-456",
				Value:      "Test value",
			},
			wantErr: false,
		},
		{
			name: "missing key_id",
			loc: Localization{
				LanguageID: "lang-456",
				Value:      "Test value",
			},
			wantErr: true,
			errMsg:  "key_id is required",
		},
		{
			name: "missing language_id",
			loc: Localization{
				KeyID: "key-123",
				Value: "Test value",
			},
			wantErr: true,
			errMsg:  "language_id is required",
		},
		{
			name: "missing value",
			loc: Localization{
				KeyID:      "key-123",
				LanguageID: "lang-456",
			},
			wantErr: true,
			errMsg:  "value is required",
		},
		{
			name: "version set to default",
			loc: Localization{
				KeyID:      "key-123",
				LanguageID: "lang-456",
				Value:      "Test value",
				Version:    0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loc.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				if tt.loc.Version == 0 {
					assert.Equal(t, 1, tt.loc.Version)
				}
			}
		})
	}
}

func TestLocalization_BeforeCreate(t *testing.T) {
	loc := Localization{
		KeyID:      "key-123",
		LanguageID: "lang-456",
		Value:      "Test value",
	}

	loc.BeforeCreate()

	assert.NotEmpty(t, loc.ID)
	assert.NotZero(t, loc.CreatedAt)
	assert.NotZero(t, loc.ModifiedAt)
	assert.Equal(t, loc.CreatedAt, loc.ModifiedAt)
	assert.Equal(t, 1, loc.Version)
}

func TestLocalization_BeforeUpdate(t *testing.T) {
	loc := Localization{
		KeyID:      "key-123",
		LanguageID: "lang-456",
		Value:      "Test value",
		CreatedAt:  1000,
		ModifiedAt: 1000,
	}

	loc.BeforeUpdate()

	assert.NotZero(t, loc.ModifiedAt)
	assert.Greater(t, loc.ModifiedAt, loc.CreatedAt)
}

func TestLocalization_Approve(t *testing.T) {
	loc := Localization{
		KeyID:      "key-123",
		LanguageID: "lang-456",
		Value:      "Test value",
		Approved:   false,
	}

	loc.Approve("admin_user")

	assert.True(t, loc.Approved)
	assert.Equal(t, "admin_user", loc.ApprovedBy)
	assert.NotZero(t, loc.ApprovedAt)
}
