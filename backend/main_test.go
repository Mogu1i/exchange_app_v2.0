package main_test

import (
	"testing"
)

func TestBasic(t *testing.T) {
	if 1+1 != 2 {
		t.Fatal("basic arithmetic failed")
	}
}
