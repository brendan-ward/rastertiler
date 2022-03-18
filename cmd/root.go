package cmd

import (
	"github.com/spf13/cobra"
)

var VERSION = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "rastertiler",
	Short:   "A Go-based single-band GeoTIFF to PNG mbtiles creator",
	Version: VERSION,
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(createCmd)
}
