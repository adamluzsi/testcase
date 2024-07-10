package doc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.llib.dev/testcase/internal"
)

type Formatter interface {
	MakeDocument(context.Context, []TestingCase) (string, error)
}

type TestingCase struct {
	// ContextPath is the Testing ContextPath
	ContextPath []string
	// TestFailed tells if the test failed
	TestFailed bool
	// TestSkipped tells if the given test was skipped
	TestSkipped bool
}

type DocumentFormat struct{}

func (gen DocumentFormat) MakeDocument(ctx context.Context, tcs []TestingCase) (string, error) {
	node := newNode()
	for _, tc := range tcs {
		node.Add(tc)
	}
	var document string
	if gen.hasFailed(node) || internal.Verbose() {
		document = gen.generateDocumentString(node, "")
	}
	return document, nil
}

type colourCode string

const (
	red    colourCode = "91m"
	green  colourCode = "92m"
	yellow colourCode = "93m"
)

func colourise(code colourCode, text string) string {
	if isColourCodingSupported() {
		return fmt.Sprintf("\033[%s%s\033[0m", code, text)
	}
	return text
}

func newNode() *node {
	return &node{Nodes: make(nodes)}
}

type node struct {
	Nodes       nodes
	TestingCase TestingCase
}

type nodes map[string]*node

func (n *node) Add(tc TestingCase) {
	n.cd(tc.ContextPath).TestingCase = tc
}

func (n *node) cd(path []string) *node {
	current := n
	for _, part := range path {
		if current.Nodes == nil {
			current.Nodes = make(nodes)
		}
		if _, ok := current.Nodes[part]; !ok {
			current.Nodes[part] = newNode()
		}
		current = current.Nodes[part]
	}
	return current
}

func (gen DocumentFormat) hasFailed(n *node) bool {
	if n == nil {
		return false
	}
	if n.TestingCase.TestFailed {
		return true
	}
	for _, child := range n.Nodes {
		if child.TestingCase.TestFailed {
			return true
		}
		if gen.hasFailed(child) {
			return true
		}
	}
	return false
}

func (gen DocumentFormat) hasFailedInSubnodes(n *node) bool {
	if n == nil {
		return false
	}
	for _, child := range n.Nodes {
		if gen.hasFailed(child) {
			return true
		}
	}
	return false
}

func (gen DocumentFormat) generateDocumentString(n *node, indent string) string {
	var sb strings.Builder
	for key, child := range n.Nodes {
		var (
			line   = key
			colour = green
		)
		if child.TestingCase.TestSkipped {
			line += " [SKIP]"
			colour = yellow
		}
		if child.TestingCase.TestFailed {
			line += " [FAIL]"
			colour = red
		}
		if len(child.Nodes) == 0 {
			line = colourise(colour, line)
		}
		if internal.Verbose() || gen.hasFailed(child) {
			sb.WriteString(indent)
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		if internal.Verbose() || gen.hasFailedInSubnodes(child) {
			sb.WriteString(gen.generateDocumentString(child, indent+"  "))
		}
	}
	return sb.String()
}

var colourSupportingTerms = map[string]struct{}{
	"xterm-256color":          {},
	"xterm-88color":           {},
	"xterm-16color":           {},
	"gnome-terminal":          {},
	"screen":                  {},
	"konsole":                 {},
	"terminator":              {},
	"aterm":                   {},
	"linux":                   {}, // default terminal type for Linux systems
	"urxvt":                   {}, // popular terminal emulator for Unix-like systems
	"konsole-256color":        {}, // 256-color version of Konsole
	"gnome-terminal-256color": {}, // 256-color version of Gnome Terminal
	"xfce4-terminal":          {}, // terminal emulator for Xfce desktop environment
	"terminator-256color":     {}, // 256-color version of Terminator
	"alacritty":               {}, // modern terminal emulator with GPU acceleration
	"kitty":                   {}, // fast, feature-rich, GPU-based terminal emulator
	"hyper":                   {}, // terminal built on web technologies
	"wezterm":                 {}, // highly configurable, GPU-accelerated terminal
	"iterm2":                  {}, // popular terminal emulator for macOS
	"st":                      {}, // simple terminal from the suckless project
	"rxvt-unicode-256color":   {}, // 256-color version of rxvt-unicode
	"rxvt-256color":           {}, // 256-color version of rxvt
	"foot":                    {}, // lightweight Wayland terminal emulator
	"mlterm":                  {}, // multilingual terminal emulator
	"putty":                   {}, // popular SSH and telnet client
}

var colourlessTerms = map[string]struct{}{
	"dumb":     {},
	"vt100":    {},
	"ansi":     {},
	"ansi.sys": {},
	"vt52":     {},
}

func isColourCodingSupported() bool {
	term, ok := os.LookupEnv("TERM")
	if !ok {
		return false
	}
	if _, ok := colourlessTerms[term]; ok {
		return false
	}
	if _, ok := colourSupportingTerms[term]; ok {
		return true
	}
	return false
}
