package http

import "testing"

func TestIsLoopbackAddr(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{"127.0.0.1:8081", true},
		{"localhost:8081", true},
		{"[::1]:8081", true},
		{":8081", false},
		{"0.0.0.0:8081", false},
		{"10.0.0.5:8081", false},
	}
	for _, c := range cases {
		if got := IsLoopbackAddr(c.addr); got != c.want {
			t.Fatalf("IsLoopbackAddr(%q) = %v, want %v", c.addr, got, c.want)
		}
	}
}
