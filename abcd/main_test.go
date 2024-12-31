package main

import "testing"

func TestMain(t *testing.T) {
	got := "Hello, World!"
	want := "Hello, World!"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
