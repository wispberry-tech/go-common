package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newRequest(query string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/?"+query, nil)
	return r
}

func TestParseQueryInt(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue int
		want         int
	}{
		{"valid", "page=5", "page", 1, 5},
		{"negative", "page=-3", "page", 1, -3},
		{"missing key", "", "page", 1, 1},
		{"invalid value", "page=abc", "page", 1, 1},
		{"empty value", "page=", "page", 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQueryInt(newRequest(tt.query), tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("ParseQueryInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseQueryInt64(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue int64
		want         int64
	}{
		{"valid", "id=9223372036854775807", "id", 0, 9223372036854775807},
		{"missing", "", "id", 42, 42},
		{"invalid", "id=notanumber", "id", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQueryInt64(newRequest(tt.query), tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("ParseQueryInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseQueryFloat64(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue float64
		want         float64
	}{
		{"valid", "price=19.99", "price", 0, 19.99},
		{"integer", "price=10", "price", 0, 10},
		{"missing", "", "price", 5.5, 5.5},
		{"invalid", "price=abc", "price", 5.5, 5.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQueryFloat64(newRequest(tt.query), tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("ParseQueryFloat64() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestParseQueryBool(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue bool
		want         bool
	}{
		{"true", "active=true", "active", false, true},
		{"TRUE", "active=TRUE", "active", false, true},
		{"1", "active=1", "active", false, true},
		{"yes", "active=yes", "active", false, true},
		{"false", "active=false", "active", true, false},
		{"0", "active=0", "active", true, false},
		{"no", "active=no", "active", true, false},
		{"missing", "", "active", true, true},
		{"invalid", "active=maybe", "active", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQueryBool(newRequest(tt.query), tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("ParseQueryBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseQueryTime(t *testing.T) {
	validTime := "2024-01-15T10:30:00Z"
	parsed, _ := time.Parse(time.RFC3339, validTime)

	tests := []struct {
		name  string
		query string
		key   string
		want  time.Time
	}{
		{"valid RFC3339", "ts=" + validTime, "ts", parsed},
		{"missing", "", "ts", time.Time{}},
		{"invalid format", "ts=not-a-time", "ts", time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQueryTime(newRequest(tt.query), tt.key)
			if !got.Equal(tt.want) {
				t.Errorf("ParseQueryTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseQueryStringPtr(t *testing.T) {
	t.Run("non-empty returns pointer", func(t *testing.T) {
		got := ParseQueryStringPtr(newRequest("name=alice"), "name")
		if got == nil || *got != "alice" {
			t.Errorf("ParseQueryStringPtr() = %v, want pointer to 'alice'", got)
		}
	})
	t.Run("missing returns nil", func(t *testing.T) {
		got := ParseQueryStringPtr(newRequest(""), "name")
		if got != nil {
			t.Errorf("ParseQueryStringPtr() = %v, want nil", got)
		}
	})
}

func TestParseQueryStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		query string
		key   string
		want  []string
	}{
		{"single value", "tag=go", "tag", []string{"go"}},
		{"multiple values", "tag=go&tag=rust", "tag", []string{"go", "rust"}},
		{"missing", "", "tag", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQueryStringSlice(newRequest(tt.query), tt.key)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ParseQueryStringSlice() = %v, want nil", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ParseQueryStringSlice() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseQueryStringSlice()[%d] = %s, want %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParsePaginationParams(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		defaultLimit int
		maxLimit     int
		wantLimit    int
		wantOffset   int
	}{
		{"defaults", "", 20, 100, 20, 0},
		{"custom values", "limit=50&offset=10", 20, 100, 50, 10},
		{"clamp over max", "limit=200", 20, 100, 100, 0},
		{"clamp under min", "limit=0", 20, 100, 1, 0},
		{"negative offset", "offset=-5", 20, 100, 20, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePaginationParams(newRequest(tt.query), tt.defaultLimit, tt.maxLimit)
			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", got.Limit, tt.wantLimit)
			}
			if got.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", got.Offset, tt.wantOffset)
			}
		})
	}
}

func TestWriteJSONResponse(t *testing.T) {
	t.Run("with data", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteJSONResponse(w, http.StatusOK, map[string]string{"key": "value"})

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", ct)
		}

		var env ResponseEnvelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if env.Error != nil {
			t.Errorf("unexpected error in envelope: %+v", env.Error)
		}
	})

	t.Run("nil data becomes empty object", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteJSONResponse(w, http.StatusOK, nil)

		var raw map[string]json.RawMessage
		if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if string(raw["data"]) != "{}" {
			t.Errorf("data = %s, want {}", string(raw["data"]))
		}
	})
}

func TestWriteJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSONError(w, http.StatusBadRequest, "INVALID_INPUT", "bad request", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var env ResponseEnvelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error in envelope")
	}
	if env.Error.Code != "INVALID_INPUT" {
		t.Errorf("error code = %s, want INVALID_INPUT", env.Error.Code)
	}
	if env.Error.Message != "bad request" {
		t.Errorf("error message = %s, want 'bad request'", env.Error.Message)
	}
}

func TestReadJSONBody(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		body := strings.NewReader(`{"name":"alice"}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)

		var v struct {
			Name string `json:"name"`
		}
		if err := ReadJSONBody(r, &v); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Name != "alice" {
			t.Errorf("Name = %s, want alice", v.Name)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		body := strings.NewReader(`{bad}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)

		var v struct{}
		if err := ReadJSONBody(r, &v); err == nil {
			t.Error("expected error for malformed JSON")
		}
	})
}
