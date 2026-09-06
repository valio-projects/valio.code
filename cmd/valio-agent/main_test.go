package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestUnknownCommandFails(t *testing.T) {
	var b bytes.Buffer
	if run(context.Background(), []string{"imaginary"}, &b) == nil {
		t.Fatal("unknown command claimed success")
	}
}
func TestDoctorOnlyReportsAvailability(t *testing.T) {
	var b bytes.Buffer
	if e := run(context.Background(), []string{"doctor"}, &b); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(b.String(), "Availability only") {
		t.Fatal("doctor overstates capability")
	}
}
