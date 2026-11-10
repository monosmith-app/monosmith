package recipewizard

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Catalog lists every level some plugin can fill, the plugins of each, and for
// each language plugin the plugins of the other levels that fit it (the same
// rule the questions of `smith recipe new` apply).
func (c *Candidates) Catalog() LevelCatalog {
	out := LevelCatalog{Fits: map[string]map[string][]string{}}
	levels := c.Levels()
	for _, level := range levels {
		offer := LevelOffer{Level: level, Additive: Additive(level)}
		for _, info := range c.For(level, nil) {
			offer.Plugins = append(offer.Plugins, PluginOffer{Name: info.Plugin.Name, Description: Describe(info)})
		}
		out.Levels = append(out.Levels, offer)
	}
	for _, lang := range c.For(LevelLanguage, nil) {
		scenario := []Scenario{{Members: []pluginhost.Member{{Info: lang, Role: protocol.RoleLanguage}}}}
		fits := map[string][]string{}
		for _, level := range levels {
			if level == LevelLanguage {
				continue
			}
			for _, info := range c.For(level, scenario) {
				fits[level] = append(fits[level], info.Plugin.Name)
			}
		}
		out.Fits[lang.Plugin.Name] = fits
	}
	return out
}

// WriteCatalog prints the catalog: as JSON, or as lines a person reads.
//
// It returns an error when the writer fails.
func WriteCatalog(w io.Writer, cat LevelCatalog, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(cat)
	}
	for _, l := range cat.Levels {
		kind := "one plugin"
		if l.Additive {
			kind = "any number"
		}
		if _, err := fmt.Fprintf(w, "%s (%s)\n", l.Level, kind); err != nil {
			return err
		}
		for _, p := range l.Plugins {
			if _, err := fmt.Fprintf(w, "  %-16s %s\n", p.Name, strings.TrimSpace(p.Description)); err != nil {
				return err
			}
		}
	}
	return nil
}
