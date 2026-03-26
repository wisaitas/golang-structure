package mask

import "testing"

func TestMaskPlainString_NM(t *testing.T) {
	if got := MaskPlainString("password", "2:2"); got != "pa****rd" {
		t.Fatalf("2:2: got %q", got)
	}
	if got := MaskPlainString("abcdefghi", "4:2"); got != "abcd***hi" {
		t.Fatalf("4:2: got %q", got)
	}
}

func TestMaskPlainString_SuffixMarker(t *testing.T) {
	// local part masked, domain @gmail.com visible
	if got := MaskPlainString("test01@gmail.com", "4:@gmail.com"); got != "test**@gmail.com" {
		t.Fatalf("suffix: got %q", got)
	}
}

func TestMaskPlainString_SuffixMarkerMiddleNoise(t *testing.T) {
	// marker in the middle → whole string N:0
	if got := MaskPlainString("me@yahoo.com", "4:@gmail.com"); got != "me@y********" {
		t.Fatalf("no suffix match: got %q", got)
	}
}

func TestMaskPlainString_PrefixMarker(t *testing.T) {
	if got := MaskPlainString("prefix-rest", "4:prefix"); got != "prefix-res*" {
		t.Fatalf("prefix: got %q", got)
	}
}

func TestMaskPlainString_EmbeddedMarker(t *testing.T) {
	// Suffix @gmail.com at end; core before it is masked with 4:0 → @gma** + @gmail.com
	if got := MaskPlainString("@gmail@gmail.com", "4:@gmail.com"); got != "@gma**@gmail.com" {
		t.Fatalf("embedded: got %q", got)
	}
}

func TestMaskPlainString_InvalidPattern(t *testing.T) {
	if got := MaskPlainString("x", "nope"); got != "nope" {
		t.Fatalf("got %q", got)
	}
}
