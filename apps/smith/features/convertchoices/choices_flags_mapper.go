package convertchoices

import "strings"

// Flags spells the choices as the command-line flags that make them.
func Flags(c Choices) string {
	return strings.Join([]string{
		"--level " + string(c.Level), "--lint " + c.Lint, "--test " + c.Test, "--build " + c.Build, "--ci " + c.CI,
	}, " ")
}

// Set sets the choice a question's key names.
func (c *Choices) Set(key, value string) {
	switch key {
	case "lint":
		c.Lint = value
	case "test":
		c.Test = value
	case "build":
		c.Build = value
	case "ci":
		c.CI = value
	}
}

// Value is the choice a question's key names.
func (c Choices) Value(key string) string {
	switch key {
	case "lint":
		return c.Lint
	case "test":
		return c.Test
	case "build":
		return c.Build
	case "ci":
		return c.CI
	}
	return ""
}
