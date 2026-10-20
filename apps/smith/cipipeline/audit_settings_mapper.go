package cipipeline

import (
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// auditSettings turns the [ci] audit_* settings into what plugins are told.
func auditSettings(ci workspace.CISection) protocol.AuditSettings {
	return protocol.AuditSettings{Policies: ci.AuditPolicies, Warn: ci.AuditWarn, AllowInstallScripts: ci.AuditAllowInstallScripts}
}
