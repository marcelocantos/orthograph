// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package doc

import "testing"

func TestEmpty(t *testing.T) {
	d := Empty("doc1", "")
	if d.Title != "Untitled" {
		t.Fatalf("title: got %q", d.Title)
	}
	if len(d.Layers) != 1 || !d.Layers[0].Visible {
		t.Fatalf("expected one visible default layer, got %+v", d.Layers)
	}
	if d.Rev != 0 {
		t.Fatalf("rev: got %d", d.Rev)
	}
}
