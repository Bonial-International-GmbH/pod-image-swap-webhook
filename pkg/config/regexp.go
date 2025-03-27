package config

import "regexp"

// Regexp is a custom wrapper around *regexp.Regexp in order to implement
// automatic regexp parsing when parsing the configuration.
type Regexp struct {
	*regexp.Regexp
}

// MustCompileRegexp is like Compile but panics if the expression cannot be
// parsed. It simplifies safe initialization of global variables holding
// compiled regular expressions.
func MustCompileRegexp(expr string) *Regexp {
	return &Regexp{Regexp: regexp.MustCompile(expr)}
}

// CompileRegexp parses a regular expression and returns, if successful, a
// Regexp object that can be used to match against text.
func CompileRegexp(expr string) (*Regexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &Regexp{Regexp: re}, nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (r *Regexp) UnmarshalText(text []byte) error {
	rr, err := CompileRegexp(string(text))
	if err != nil {
		return err
	}

	*r = *rr
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (r *Regexp) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// IsEmpty returns true if the Regexp is nil or the source text is an empty string.
func (r *Regexp) IsEmpty() bool {
	return r == nil || r.String() == ""
}
