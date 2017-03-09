package main

import (
	"testing"
)

func mustNetPortFromString(t *testing.T, s string) *NetPort {
	np := &NetPort{}
	if err := np.FromString(s); err != nil {
		t.Fatal(err)
	}
	return np
}

func mustSpecMatch(t *testing.T, spec *Spec, src, dst *NetPort, expect bool) {
	result := spec.Match("test", src, dst)
	if result != expect {
		t.Fatalf("spec=%s against src=%s dst=%s result=%v expected=%v",
			spec, src.String(), dst.String(), result, expect)
	}
}

func TestSpecMatch(t *testing.T) {
	spec := &Spec{
		src: *mustNetPortFromString(t, "10.0.0.0/16:*"),
		dst: *mustNetPortFromString(t, "1.2.3.4/32:443"),
	}
	mustSpecMatch(t, spec, mustNetPortFromString(t, "10.0.3.72/32:48291"), mustNetPortFromString(t, "1.2.3.4/32:443"), true)
	mustSpecMatch(t, spec, mustNetPortFromString(t, "10.0.3.72/32:48291"), mustNetPortFromString(t, "1.2.3.4/32:80"), false)
}
