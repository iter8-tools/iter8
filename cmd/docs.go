package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const docsDesc = `
Generate markdown documentation for Iter8 CLI commands. Documentation will be generated for all commands that are not hidden.

This command is intended for Iter8 documentation and CI.
`

// newDocsCmd creates the docs command
func newDocsCmd() *cobra.Command {
	docsDir := ""
	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate markdown documentation for Iter8 CLI",
		Long:   docsDesc,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			standardLinks := func(s string) string { return s }

			hdrFunc := func(filename string) string {
				base := filepath.Base(filename)
				name := strings.TrimSuffix(base, path.Ext(base))
				title := strings.Title(strings.Replace(name, "_", " ", -1))
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
			if err := doc.GenMarkdownTreeCustom(rootCmd, docsDir, hdrFunc, standardLinks); err != nil {
				log.Logger.Error(err)
			}
		},
	}
	addDocsFlags(cmd, &docsDir)
	return cmd
}

// addDocsFlags defines the flags for the docs command
func addDocsFlags(cmd *cobra.Command, docsDirPtr *string) {
	cmd.Flags().StringVar(docsDirPtr, "commandDocsDir", "", "directory where Iter8 CLI documentation will be created")
	cmd.MarkFlagRequired("commandDocsDir")
}

func init() {
	rootCmd.AddCommand(newDocsCmd())
}
