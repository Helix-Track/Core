package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		lang    Language
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid language",
			lang: Language{
				Code: "en",
				Name: "English",
			},
			wantErr: false,
		},
		{
			name: "missing code",
			lang: Language{
				Name: "English",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "code too long",
			lang: Language{
				Code: "en-US-EXTRA",
				Name: "English",
			},
			wantErr: true,
			errMsg:  "code must be 10 characters or less",
		},
		{
			name: "missing name",
			lang: Language{
				Code: "en",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			lang: Language{
				Code: "en",
				Name:  string(make([]byte, 101)),
			},
			wantErr: true,
			errMsg:  "name must be 100 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lang.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLanguage_BeforeCreate(t *testing.T) {
	lang := Language{
		Code: "en",
		Name: "English",
	}

	lang.BeforeCreate()

	assert.NotEmpty(t, lang.ID)
	assert.NotZero(t, lang.CreatedAt)
	assert.NotZero(t, lang.ModifiedAt)
	assert.Equal(t, lang.CreatedAt, lang.ModifiedAt)
}

func TestLanguage_BeforeUpdate(t *testing.T) {
	lang := Language{
		Code:       "en",
		Name:       "English",
		CreatedAt:  1000,
		ModifiedAt: 1000,
	}

	lang.BeforeUpdate()

	assert.NotZero(t, lang.ModifiedAt)
	assert.NotEqual(t, lang.CreatedAt, lang.ModifiedAt)
	assert.Greater(t, lang.ModifiedAt, lang.CreatedAt)
}
