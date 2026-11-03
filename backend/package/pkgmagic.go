// Package ffxpackage validates package magic identifiers.
//
// The magic is the compatibility marker between a package and its consumer
// (installer): a reader only accepts a package whose magic matches its own.
// Both v1 and v2 headers reserve exactly 8 bytes for it, so the maximum the
// code accepts is 8 printable ASCII bytes — anything else is rejected.
package ffxpackage

import "fmt"

// Len is the exact number of bytes the header reserves for the magic.
const Len = 8

// Validate checks that s is a usable package magic: exactly Len printable
// ASCII bytes. It returns a descriptive error otherwise.
func Validate(s string) error {
	if len(s) != Len {
		return fmt.Errorf("magic must be exactly %d bytes, got %d %q", Len, len(s), s)
	}
	for i := 0; i < Len; i++ {
		c := s[i]
		if c < 0x20 || c > 0x7e {
			return fmt.Errorf("magic must be printable ASCII (0x20-0x7e), byte %d is 0x%02x", i, c)
		}
	}
	return nil
}