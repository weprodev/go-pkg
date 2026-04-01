// Package sanitizer provides configurable HTML sanitization for Go web services.
//
// It wraps the bluemonday library and ships with safe, opinionated defaults for
// rich-text content (bold, italic, paragraphs, lists, highlight marks, semantic
// colour classes, text alignment). All defaults can be extended or replaced via
// functional options.
//
// Quick start — use the package-level defaults:
//
//	safe := sanitizer.SanitizeHTML(userInput)
//	plain := sanitizer.StripAllHTML(userInput)
//
// Custom sanitizer with additional elements:
//
//	s := sanitizer.New(
//	    sanitizer.WithElements("a", "img"),
//	    sanitizer.WithAttrs("href", "src").OnElements("a", "img"),
//	)
//	safe := s.Sanitize(userInput)
package sanitizer

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// ─── regex helpers ────────────────────────────────────────────────────────────

// colorClassRegex matches the four semantic colour classes used by common
// CSS frameworks (Bootstrap / Tailwind-compatible naming).
var colorClassRegex = regexp.MustCompile(`^(text-(primary|success|warning|danger))$`)

// hexColorRegex matches a strict 6-digit hex colour: #RRGGBB.
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// bgColorRegex matches safe background-color values only: rgb(N,N,N) or
// #RRGGBB. Rejects url(), expression(), var(), and any CSS function.
var bgColorRegex = regexp.MustCompile(`^(rgb\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*\)|#[0-9A-Fa-f]{6})$`)

// ─── options ─────────────────────────────────────────────────────────────────

// Option is a functional option that mutates a bluemonday.Policy.
type Option func(p *bluemonday.Policy)

// WithElements adds extra allowed HTML elements on top of (or instead of)
// the defaults when used with New.
func WithElements(elements ...string) Option {
	return func(p *bluemonday.Policy) {
		p.AllowElements(elements...)
	}
}

// AttrBuilder is returned by WithAttrs to allow chaining OnElements.
type AttrBuilder struct {
	attrs []string
}

// WithAttrs returns an AttrBuilder that you can then attach to elements.
//
//	sanitizer.WithAttrs("href", "title").OnElements("a")
func WithAttrs(attrs ...string) *AttrBuilder {
	return &AttrBuilder{attrs: attrs}
}

// OnElements returns an Option that allows the configured attributes on the
// given elements.
func (b *AttrBuilder) OnElements(elements ...string) Option {
	return func(p *bluemonday.Policy) {
		p.AllowAttrs(b.attrs...).OnElements(elements...)
	}
}

// WithStyleProperties adds allowed CSS property/value pairs.
//
//	sanitizer.WithStyleProperties("font-size").MatchingEnum("small","medium","large").OnElements("p")
//
// Because bluemonday's fluent API returns chained builders we accept a raw
// func(p) here for maximum flexibility.
func WithPolicyFunc(fn func(p *bluemonday.Policy)) Option {
	return fn
}

// ─── Sanitizer ────────────────────────────────────────────────────────────────

// Sanitizer wraps a bluemonday policy and exposes Sanitize / StripAll methods.
type Sanitizer struct {
	policy *bluemonday.Policy
}

// New creates a Sanitizer starting from the built-in safe defaults and
// applying any provided options on top.
//
// Options are applied in order, so later options can override earlier ones.
func New(opts ...Option) *Sanitizer {
	p := defaultPolicy()
	for _, opt := range opts {
		opt(p)
	}
	return &Sanitizer{policy: p}
}

// NewStrict creates a Sanitizer that strips ALL HTML tags (plain text only).
// Options are ignored — use this when you need guaranteed plain text.
func NewStrict() *Sanitizer {
	return &Sanitizer{policy: bluemonday.StrictPolicy()}
}

// NewBlank creates a Sanitizer with a completely empty policy.
// All allowed elements and attributes must be added via options.
//
//	s := sanitizer.NewBlank(
//	    sanitizer.WithElements("b", "i"),
//	)
func NewBlank(opts ...Option) *Sanitizer {
	p := bluemonday.NewPolicy()
	for _, opt := range opts {
		opt(p)
	}
	return &Sanitizer{policy: p}
}

// Sanitize sanitizes html using this Sanitizer's policy.
func (s *Sanitizer) Sanitize(html string) string {
	if html == "" {
		return html
	}
	return s.policy.Sanitize(html)
}

// StripAll removes every HTML tag and returns plain text.
func (s *Sanitizer) StripAll(html string) string {
	if html == "" {
		return html
	}
	return bluemonday.StrictPolicy().Sanitize(html)
}

// ─── Package-level defaults ───────────────────────────────────────────────────

var (
	defaultSanitizer = &Sanitizer{policy: defaultPolicy()}
	strictSanitizer  = NewStrict()
)

// SanitizeHTML sanitizes html using the package-level default policy.
// Thread-safe; suitable as a drop-in for the common case.
func SanitizeHTML(html string) string {
	return defaultSanitizer.Sanitize(html)
}

// StripAllHTML removes all HTML tags and returns plain text.
func StripAllHTML(html string) string {
	return strictSanitizer.StripAll(html)
}

// ─── default policy ───────────────────────────────────────────────────────────

func defaultPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	// Safe rich-text formatting and structural elements.
	p.AllowElements("b", "i", "em", "strong", "u", "br", "p", "span", "mark", "ul", "ol", "li")

	// class on <span>/<p>: only the four semantic colour tokens.
	p.AllowAttrs("class").Matching(colorClassRegex).OnElements("span", "p")

	// data-color on <mark>: strict 6-digit hex only.
	p.AllowAttrs("data-color").Matching(hexColorRegex).OnElements("mark")

	// style attribute on block and highlight elements.
	p.AllowAttrs("style").OnElements("p", "span", "mark", "ul", "ol", "li")

	// text-align: only the four standard values.
	p.AllowStyles("text-align").
		MatchingEnum("left", "center", "right", "justify").
		OnElements("p", "span", "ul", "ol", "li")

	// background-color on <mark>: rgb or hex only (no CSS functions).
	p.AllowStyles("background-color").
		MatchingHandler(func(value string) bool { return bgColorRegex.MatchString(value) }).
		OnElements("mark")

	// color on <mark>: only "inherit".
	p.AllowStyles("color").MatchingEnum("inherit").OnElements("mark")

	return p
}
