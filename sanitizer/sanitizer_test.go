package sanitizer_test

import (
	"testing"

	"github.com/weprodev/go-pkg/sanitizer"
)

// ─── Default SanitizeHTML ─────────────────────────────────────────────────────

func TestSanitizeHTML_RemovesDangerousTags(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{
			name:  "removes script tag",
			input: `<b>Bold</b><script>alert('XSS')</script>`,
			want:  `<b>Bold</b>`,
		},
		{
			name:  "removes iframe",
			input: `<p>Text</p><iframe src="evil.com"></iframe>`,
			want:  `<p>Text</p>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizer.SanitizeHTML(tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeHTML_PreservesSafeTags(t *testing.T) {
	input := `<b>Bold</b> <i>Italic</i> <em>Em</em> <strong>Strong</strong> <u>U</u> <br> <p>P</p> <span>Span</span> <mark>Mark</mark> <ul><li>Item</li></ul> <ol><li>Ord</li></ol>`
	got := sanitizer.SanitizeHTML(input)
	if got != input {
		t.Errorf("safe tags altered:\ngot:  %s\nwant: %s", got, input)
	}
}

func TestSanitizeHTML_PreservesColorClasses(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{`text-primary`, `<span class="text-primary">A</span>`, `<span class="text-primary">A</span>`},
		{`text-success`, `<p class="text-success">B</p>`, `<p class="text-success">B</p>`},
		{`text-warning`, `<span class="text-warning">C</span>`, `<span class="text-warning">C</span>`},
		{`text-danger`, `<p class="text-danger">D</p>`, `<p class="text-danger">D</p>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizer.SanitizeHTML(tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeHTML_StripsInvalidClasses(t *testing.T) {
	got := sanitizer.SanitizeHTML(`<span class="text-custom">Text</span>`)
	if got != `<span>Text</span>` {
		t.Errorf("got %q, want bare <span>", got)
	}
}

func TestSanitizeHTML_StripsInvalidStyleProperties(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"strips color", `<p style="color: red">T</p>`, `<p>T</p>`},
		{"strips font-size", `<p style="font-size: 100px">T</p>`, `<p>T</p>`},
		{"keeps text-align, strips rest", `<p style="text-align: center; color: red; font-size: 100px">T</p>`, `<p style="text-align: center">T</p>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizer.SanitizeHTML(tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeHTML_MarkHighlight(t *testing.T) {
	input := `<mark data-color="#D33C3C" style="background-color: rgb(211, 60, 60); color: inherit;">hello</mark>`
	want := `<mark data-color="#D33C3C" style="background-color: rgb(211, 60, 60); color: inherit">hello</mark>`
	if got := sanitizer.SanitizeHTML(input); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSanitizeHTML_Empty(t *testing.T) {
	if got := sanitizer.SanitizeHTML(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// ─── StripAllHTML ─────────────────────────────────────────────────────────────

func TestStripAllHTML(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"removes formatting", `<b>Bold</b> <i>Italic</i>`, `Bold Italic`},
		{"removes script", `<b>Safe</b><script>alert(1)</script>`, `Safe`},
		{"plain text preserved", `Hello world`, `Hello world`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizer.StripAllHTML(tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// ─── Sanitizer with options ───────────────────────────────────────────────────

func TestNew_WithExtraElements(t *testing.T) {
	// Default policy strips <a> tags; adding it should preserve them.
	s := sanitizer.New(
		sanitizer.WithElements("a"),
		sanitizer.WithAttrs("href").OnElements("a"),
	)

	input := `<a href="https://example.com">link</a>`
	got := s.Sanitize(input)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestNewBlank_OnlyAllowedElements(t *testing.T) {
	s := sanitizer.NewBlank(sanitizer.WithElements("b"))
	got := s.Sanitize(`<b>Bold</b> <i>Italic</i>`)
	want := `<b>Bold</b> Italic`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNewStrict_StripsAll(t *testing.T) {
	s := sanitizer.NewStrict()
	got := s.Sanitize(`<b>Bold</b><script>evil</script>`)
	if got != "Bold" {
		t.Errorf("got %q, want %q", got, "Bold")
	}
}
