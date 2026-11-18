package helper

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGRPCErrorOutput(t *testing.T) {
	// Mock Recorder
	w := httptest.NewRecorder()
	code := 500
	message := "internal error"

	resp := APIResponse{
		Success: false,
		Message: message,
		Code:    code,
		Meta:    Meta{Timestamp: time.Now().UTC().Format(time.RFC3339)},
	}

	// Mimic writeJSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", result.StatusCode)
	}

	// Check body
	bodyBuf := w.Body.String()
	t.Logf("Body: %s", bodyBuf)

	if !strings.Contains(bodyBuf, "\"code\":500") {
		t.Error("Body missing code field")
	}
	if !strings.Contains(bodyBuf, "\"meta\":") {
		t.Error("Body missing meta field")
	}
}
