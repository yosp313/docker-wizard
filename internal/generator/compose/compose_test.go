package compose

import (
	"strings"
	"testing"

	"docker-wizard/internal/generator/catalog"
)

func TestWriteServiceRendersExposeForInternalServicesWithoutPorts(t *testing.T) {
	b := &strings.Builder{}
	writeService(b, catalog.ServiceSpec{
		ID:     "internalapi",
		Name:   "internalapi",
		Image:  "busybox",
		Public: false,
		Expose: []string{"9090"},
	})

	output := b.String()
	if !strings.Contains(output, "    expose:\n") {
		t.Fatalf("expected expose block in service output: %q", output)
	}
	if !strings.Contains(output, "      - \"9090\"\n") {
		t.Fatalf("expected expose port 9090 in service output: %q", output)
	}
	if strings.Contains(output, "    ports:\n") {
		t.Fatalf("did not expect published ports for internal service: %q", output)
	}
}
