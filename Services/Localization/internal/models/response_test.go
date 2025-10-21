package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	data := map[string]string{"test": "data"}
	response := SuccessResponse(data)

	assert.True(t, response.Success)
	assert.Equal(t, data, response.Data)
	assert.Nil(t, response.Error)
}

func TestErrorResponse(t *testing.T) {
	code := 1001
	message := "test error"
	response := ErrorResponse(code, message)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, code, response.Error.Code)
	assert.Equal(t, message, response.Error.Message)
}
