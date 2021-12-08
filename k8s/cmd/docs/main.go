package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	k8scli "github.com/iter8-tools/iter8/k8s/cmd"

	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

const (
	commandDocsDir = "COMMAND_DOCS_DIR"
)

func main() {

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

	// initialize command docs dir
	viper.BindEnv(commandDocsDir)
	viper.SetDefault(commandDocsDir, base.CompletePath("../../../mkdocs/docs/user-guide", "commands"))
	cdd := viper.GetString(commandDocsDir)

	// automatically generate markdown documentation for all Iter8 commands
	// err := doc.GenMarkdownTreeCustom(cmd.RootCmd, cdd, hdrFunc, standardLinks)
	root := k8scli.RootCmd
	// root.AddCommand(k8scli.NewKCommand())
	err := doc.GenMarkdownTreeCustom(root, cdd, hdrFunc, standardLinks)
	if err != nil {
		log.Logger.Error(err)
	}
}
