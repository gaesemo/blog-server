/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

const (
	appName = "gsm"
	version = "0.0.0"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     appName,
	Version: version,
	Short:   "gaesemo tech-blog-server",
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func ExecuteWithContext(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}
