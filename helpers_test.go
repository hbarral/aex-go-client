package aex

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// jsonDecodeStrict decodes r's body into v, failing the test on error.
func jsonDecodeStrict(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// dateFrom parses a yyyy-mm-dd string into a time.Time, failing the test
// on error.
func dateFrom(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}

// readFixture loads a JSON fixture from testdata/.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

// fixtureEndpoint returns a handler that serves the auth endpoint and
// replies to every other path with the given fixture, capturing the
// decoded request body of the endpoint call into captured (when non-nil).
func fixtureEndpoint(t *testing.T, fixtureName string, captured any) http.Handler {
	return countingAuthEndpoint(t, new(int), func(w http.ResponseWriter, r *http.Request) {
		if captured != nil {
			if err := jsonDecodeStrict(r, captured); err != nil {
				t.Errorf("decode request: %v", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(readFixture(t, fixtureName))
	})
}

// errorEndpoint returns a handler that serves the auth endpoint and
// replies to every other path with the given API error envelope.
func errorEndpoint(t *testing.T, code, message string) http.Handler {
	return countingAuthEndpoint(t, new(int), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"` + code + `","mensaje":"` + message + `","datos":null}`))
	})
}
