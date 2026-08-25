package main

import "testing"

func TestEntrypointPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
}
