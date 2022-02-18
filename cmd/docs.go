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

// commandDocsDir is the location where command docs need to land
var commandDocsDir string

// docsCmd represents the docsCmd command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate markdown documentation for Iter8 CLI.",
	Long: `
	Generate markdown documentation for Iter8 CLI.`,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		err := doc.GenMarkdownTreeCustom(rootCmd, commandDocsDir, hdrFunc, standardLinks)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
		return nil
	},
}

func init() {
	docsCmd.Flags().StringVarP(&commandDocsDir, "commandDocsDir", "c", "", "directory where CLI documentation will be created")
	docsCmd.MarkFlagRequired("commandDocsDir")
	rootCmd.AddCommand(docsCmd)
}
