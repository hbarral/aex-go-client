package aex

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTFBoolUnmarshal(t *testing.T) {
	tests := []struct {
		input   string
		want    TFBool
		wantErr bool
	}{
		{`"t"`, true, false},
		{`"f"`, false, false},
		{`true`, true, false},
		{`false`, false, false},
		{`null`, false, false},
		{`""`, false, false},
		{`"x"`, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var got TFBool
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		})
	}
}

func TestTFBoolMarshal(t *testing.T) {
	tests := []struct {
		value TFBool
		want  string
	}{
		{true, `"t"`},
		{false, `"f"`},
	}

	for _, tt := range tests {
		got, err := json.Marshal(tt.value)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(got) != tt.want {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}

func TestFlexIntUnmarshal(t *testing.T) {
	tests := []struct {
		input   string
		want    FlexInt
		wantErr bool
	}{
		{`17`, 17, false},
		{`"17"`, 17, false},
		{`null`, 0, false},
		{`""`, 0, false},
		{`"abc"`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var got FlexInt
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFlexIntMarshal(t *testing.T) {
	got, err := json.Marshal(FlexInt(17))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(got) != "17" {
		t.Errorf("got %s, want 17", got)
	}
}

func TestDateTimeUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"api layout", `"2025-06-06 12:34:56"`, time.Date(2025, time.June, 6, 12, 34, 56, 0, time.UTC), false},
		{"rfc3339", `"2025-06-06T12:34:56Z"`, time.Date(2025, time.June, 6, 12, 34, 56, 0, time.UTC), false},
		{"null", `null`, time.Time{}, false},
		{"empty", `""`, time.Time{}, false},
		{"invalid", `"06/06/2025"`, time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DateTime
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !got.Time.Equal(tt.want) {
				t.Errorf("got %v, want %v", got.Time, tt.want)
			}
		})
	}
}
