package cmd

import (
	"github.com/spf13/cobra"
	"infra/internal/helper"
)

var rootCmd = &cobra.Command{
	Use: "infra",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Run() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: false})
	helper.Must(rootCmd.Execute())
}
