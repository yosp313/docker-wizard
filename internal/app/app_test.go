package app

import "testing"

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Mode
		wantErr bool
	}{
		{name: "default empty", input: "", want: ModeStyled},
		{name: "styled", input: "styled", want: ModeStyled},
		{name: "plain", input: "plain", want: ModePlain},
		{name: "cli", input: "cli", want: ModeCLI},
		{name: "batch", input: "batch", want: ModeBatch},
		{name: "case-insensitive", input: "PLAIN", want: ModePlain},
		{name: "invalid", input: "weird", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
