package main

import (
	"fmt"
	"os"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
	"github.com/Ktulue/KtulueKit-Migration/internal/reporter"
	"github.com/Ktulue/KtulueKit-Migration/internal/runner"
	"github.com/spf13/cobra"
)

var (
	configPaths []string
	dryRun      bool
)

var rootCmd = &cobra.Command{
	Use:   "ktuluekit-migration",
	Short: "KtulueKit Migration restores your personal state from backup to a fresh Windows 11 install",
	Long: `KtulueKit-Migration reads a declarative JSON config and copies application
profiles, configs, and media assets from a backup location to the correct
Windows 11 paths. It is purely additive — it never deletes anything.`,
	RunE: runMigration,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show which migration items have source data available",
	RunE:  runStatus,
}

func init() {
	rootCmd.PersistentFlags().StringSliceVarP(&configPaths, "config", "c", []string{"ktuluekit-migration.json"}, "Path(s) to migration config files")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Preview what would be copied without actually copying")

	rootCmd.AddCommand(statusCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runMigration(cmd *cobra.Command, args []string) error {
	for _, path := range configPaths {
		cfg, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("loading config %s: %w", path, err)
		}

		if dryRun {
			fmt.Println("DRY RUN — no files will be copied")
			fmt.Println()
		}

		rep := reporter.New("logs")
		defer rep.Close()

		r := runner.New(cfg, rep)
		r.SetOnProgress(func(evt runner.ProgressEvent) {
			fmt.Printf("[%d/%d] %s %s — %s: %s\n",
				evt.Index, evt.Total, statusIcon(evt.Status),
				evt.App, evt.Label, evt.Status)
			if evt.Detail != "" {
				fmt.Printf("        %s\n", evt.Detail)
			}
		})

		result := r.Run()

		fmt.Println()
		fmt.Printf("Migration complete in %s\n", result.Elapsed.Round(1))
		fmt.Printf("Total bytes copied: %d\n", result.TotalBytes)
		fmt.Printf("Log: %s\n", rep.LogPath())

		if rep.HasFailures() {
			fmt.Println("\nSome items failed — check the log for details.")
			return fmt.Errorf("migration completed with failures")
		}
	}
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	for _, path := range configPaths {
		cfg, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("loading config %s: %w", path, err)
		}

		fmt.Printf("Backup root: %s\n\n", cfg.BackupRoot)

		for _, app := range cfg.Apps {
			fmt.Printf("[%s]\n", app.Name)
			for _, item := range app.Items {
				// TODO: resolve paths and check if source exists
				fmt.Printf("  %s: %s -> %s\n", item.Label, item.Source, item.Target)
			}
			fmt.Println()
		}
	}
	return nil
}

func statusIcon(status string) string {
	switch status {
	case "copying":
		return "\u23f3"
	case "copied":
		return "\u2705"
	case "skipped":
		return "\u23ed\ufe0f"
	case "failed":
		return "\u274c"
	default:
		return "\u2753"
	}
}
