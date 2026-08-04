// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Orthograph is a Pencil-first collaborative sketch surface for technical
// diagrams, shared between a human on iPad and coding agents on the Mac.
package main

import (
	"fmt"
	"os"
	"strings"
)

// version is set at link time via -X main.version=...
var version = "dev"

const agentGuide = `Orthograph — agent guide (CLI)

What: collaborative technical sketch surface (Mac daemon + MCP + iPad/Pencil over pigeon).
Docs: README.md, agents-guide.md, docs/design.md

Build:
  make bullseye

CLI (scaffold):
  orthograph --version
  orthograph --help
  orthograph --help-agent

Planned subcommands:
  orthograph daemon     # MCP + pigeon backend
  orthograph pair       # QR / instance id for iPad
  orthograph status     # clients + active document
  orthograph open       # create or select document
  orthograph export     # svg|png|pdf|json

MCP tools (planned; server name "orthograph"):
  orthograph_status, orthograph_scene, orthograph_render,
  orthograph_create, orthograph_update, orthograph_delete,
  orthograph_push_image, …

Observe vector first (orthograph_scene); use render only when pixels matter.
Standalone product — not a Jevons feature; any MCP client may attach.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(os.Stdout)
		return nil
	case "--help-agent", "help-agent":
		fmt.Fprint(os.Stdout, agentGuide)
		return nil
	case "-v", "--version", "version":
		fmt.Println(version)
		return nil
	case "daemon", "pair", "status", "open", "export":
		return fmt.Errorf("%s: not implemented yet (scaffold); see docs/design.md", args[0])
	default:
		if strings.HasPrefix(args[0], "-") {
			return fmt.Errorf("unknown flag %q\n\n%s", args[0], usageText())
		}
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText())
	}
}

func usageText() string {
	return `Usage: orthograph <command> [flags]

Commands (planned):
  daemon    Run Mac daemon (MCP + pigeon)
  pair      Show pairing QR / instance id
  status    Connection and document status
  open      Create or select a document
  export    Export active document

Flags:
  --version       Print version
  --help          Show this help
  --help-agent    Help + agent guide

Design: docs/design.md
`
}

func printUsage(w *os.File) {
	fmt.Fprint(w, usageText())
}
