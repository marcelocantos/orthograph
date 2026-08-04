// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	// run writes to stdout; we only check it does not error.
	if err := run([]string{"--version"}); err != nil {
		t.Fatalf("--version: %v", err)
	}
}

func TestRunHelp(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatalf("--help: %v", err)
	}
}

func TestRunHelpAgent(t *testing.T) {
	if err := run([]string{"--help-agent"}); err != nil {
		t.Fatalf("--help-agent: %v", err)
	}
	if !strings.Contains(agentGuide, "orthograph_scene") {
		t.Fatal("agent guide missing orthograph_scene mention")
	}
}

func TestRunUnknown(t *testing.T) {
	err := run([]string{"nope"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunNotImplemented(t *testing.T) {
	err := run([]string{"daemon"})
	if err == nil {
		t.Fatal("expected not-implemented error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("unexpected error: %v", err)
	}
}
