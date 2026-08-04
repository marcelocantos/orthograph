// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package doc holds the vector document model and op log.
// Phase 0 will add persistence and mutation; this file only defines
// the public types so MCP and protocol packages can compile against them.
package doc

// ID is a stable object or document identifier (ULID string in production).
type ID string

// Author records who created or last mutated an object.
type Author struct {
	Kind    string `json:"kind"` // "human" | "agent"
	Name    string `json:"name,omitempty"`
	Session string `json:"session,omitempty"`
}

// Point is a canvas coordinate in document points.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BBox is an axis-aligned bounding box.
type BBox struct {
	Min Point `json:"min"`
	Max Point `json:"max"`
}

// Object is one scene element. Type-specific fields live in Props until
// the schema stabilizes.
type Object struct {
	ID     ID             `json:"id"`
	Type   string         `json:"type"` // stroke, rect, ellipse, line, arrow, connector, text, image, group, note, frame
	Author Author         `json:"author"`
	Layer  ID             `json:"layer,omitempty"`
	BBox   BBox           `json:"bbox"`
	Tags   []string       `json:"tags,omitempty"`
	Props  map[string]any `json:"props,omitempty"`
}

// Layer groups objects for visibility and z-order.
type Layer struct {
	ID      ID     `json:"id"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
	Locked  bool   `json:"locked"`
	Z       int    `json:"z"`
}

// Document is the vector scene agents and the pad share.
type Document struct {
	ID      ID       `json:"id"`
	Title   string   `json:"title"`
	Rev     uint64   `json:"rev"`
	Layers  []Layer  `json:"layers"`
	Objects []Object `json:"objects"`
}

// Empty returns a new empty document with a default layer.
func Empty(id ID, title string) *Document {
	if title == "" {
		title = "Untitled"
	}
	return &Document{
		ID:    id,
		Title: title,
		Rev:   0,
		Layers: []Layer{{
			ID:      "layer-default",
			Name:    "Default",
			Visible: true,
			Locked:  false,
			Z:       0,
		}},
		Objects: nil,
	}
}
