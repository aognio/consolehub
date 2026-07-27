package ulid_test

import (
	"testing"

	"consolehub/libraries/go/consolehub/ulid"
)

func TestULID_GenerationAndValidation(t *testing.T) {
	id1 := ulid.Make()
	id2 := ulid.Make()

	if len(id1) != 26 {
		t.Errorf("expected ULID length 26, got %d (%s)", len(id1), id1)
	}

	if !ulid.IsValid(id1) {
		t.Errorf("expected ULID %s to be valid", id1)
	}

	if id1 == id2 {
		t.Errorf("expected unique ULIDs, got duplicate %s", id1)
	}
}
