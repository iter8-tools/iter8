package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/spf13/cobra"
)

// docsDesc is the description for the docs command
const docsDesc = `
Generate markdown documentation for Iter8 CLI commands. Documentation will be generated for all commands that are not hidden.
`

// newDocsCmd creates the docs command
func newDocsCmd() *cobra.Command {
	docsDir := ""
	cmd := &cobra.Command{
		Use:          "docs",
		Short:        "Generate markdown documentation for Iter8 CLI",
		Long:         docsDesc,
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			standardLinks := func(s string) string { return s }

			hdrFunc := func(filename string) string {
				base := filepath.Base(filename)
				name := strings.TrimSuffix(base, path.Ext(base))
				caser := cases.Title(language.English)
				title := caser.String(strings.Replace(name, "_", " ", -1))
				tpl := `---
template: main.html
title: "%s"
hide:
- toc
---
`
				return fmt.Sprintf(tpl, title)
			}

			// automatically generate markdown documentation for all Iter8 commands
			return doc.GenMarkdownTreeCustom(rootCmd, docsDir, hdrFunc, standardLinks)
		},
	}
	addDocsFlags(cmd, &docsDir)
	return cmd
}

// addDocsFlags defines the flags for the docs command
func addDocsFlags(cmd *cobra.Command, docsDirPtr *string) {
	cmd.Flags().StringVar(docsDirPtr, "commandDocsDir", "", "directory where Iter8 CLI documentation will be created")
	_ = cmd.MarkFlagRequired("commandDocsDir")
}

// initialize with docs command
func init() {
	rootCmd.AddCommand(newDocsCmd())
}
