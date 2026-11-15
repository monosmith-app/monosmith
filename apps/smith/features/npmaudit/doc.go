// Package npmaudit is `smith npm-audit`, the node plugin's audit step: it
// fails only on a production advisory at moderate or above that npm can fix,
// and prints the rest as notes. Hidden: the audit phase calls it through
// {smith}, people call smith ci audit.
package npmaudit
