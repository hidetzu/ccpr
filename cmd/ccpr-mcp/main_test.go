package main

import "testing"

func TestNewServerRegistersTools(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("newServer() panicked: %v", r)
		}
	}()

	if got := newServer(); got == nil {
		t.Fatal("newServer() returned nil")
	}
}
