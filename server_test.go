package main

import "testing"

func TestHTTP_Starts(t *testing.T) {
	h := NewHTTP()
	err := h.Start(":8080")
	if err != nil {
		t.Fail()
	}
}
