// Package functionhost knows the local tool each function host needs (Azure
// Functions Core Tools, the AWS SAM CLI, the Google Functions Framework) and
// the doctor finding that says how to get it. The language plugins that scaffold
// function apps (ADR 0072) share it, so the three languages word it the same.
package functionhost
