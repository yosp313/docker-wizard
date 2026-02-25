package cli

import "testing"

func TestParseIndexSelection(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		max     int
		want    []int
		wantErr bool
	}{
		{name: "single", input: "2", max: 3, want: []int{2}},
		{name: "multiple", input: "1,3", max: 3, want: []int{1, 3}},
		{name: "dedupe", input: "2,2,1", max: 3, want: []int{1, 2}},
		{name: "with spaces", input: " 1, 3 ", max: 3, want: []int{1, 3}},
		{name: "empty", input: "", max: 3, wantErr: true},
		{name: "not number", input: "x", max: 3, wantErr: true},
		{name: "out of range", input: "4", max: 3, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIndexSelection(tt.input, tt.max)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("expected %v, got %v", tt.want, got)
				}
			}
		})
	}
}
