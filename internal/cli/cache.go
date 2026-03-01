package cli

import (
	"fmt"

	"github.com/k15z/axiom/internal/config"
	"github.com/k15z/axiom/internal/runner"
	"github.com/spf13/cobra"
)

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the test cache",
	}
	cmd.AddCommand(newCacheClearCmd())
	return cmd
}

func newCacheClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear the test cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Default()
			if err := runner.ClearCache(cfg.Cache.Dir); err != nil {
				return fmt.Errorf("clearing cache: %w", err)
			}
			fmt.Println("Cache cleared.")
			return nil
		},
	}
}
