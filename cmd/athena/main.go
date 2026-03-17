package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/akhil-datla/athena/internal/engine"
	"github.com/akhil-datla/athena/internal/loader"
	"github.com/akhil-datla/athena/internal/server"
	"github.com/akhil-datla/athena/internal/tui"
	"github.com/akhil-datla/athena/internal/watcher"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "athena",
		Short: "Lightning-fast Q&A search engine",
		Long: `Athena is a lightning-fast Q&A search engine for your FAQ databases.

Load your question-and-answer data from CSV or JSON, then search
instantly via the CLI or a REST API. Athena ranks results using
fuzzy matching with Levenshtein distance so you always get the
most relevant answers first.`,
		SilenceUsage: true,
	}

	root.AddCommand(serveCmd())
	root.AddCommand(searchCmd())
	root.AddCommand(versionCmd())

	return root
}

func serveCmd() *cobra.Command {
	var (
		file  string
		port  int
		watch bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Athena REST API server with web UI",
		Example: `  athena serve -f faq.csv
  athena serve -f data.json -p 3000
  athena serve -f faq.csv --watch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := loader.Load(file)
			if err != nil {
				return err
			}

			eng := engine.New(entries)
			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

			logger.Info("loaded entries", "count", eng.Count(), "file", file)

			fmt.Println(banner())
			fmt.Printf("  Web UI:  http://localhost:%d\n", port)
			fmt.Printf("  API:     http://localhost:%d/api/search?q=your+query\n", port)
			fmt.Printf("  Entries: %d loaded from %s\n", eng.Count(), file)
			if watch {
				fmt.Printf("  Watch:   enabled (auto-reload on file changes)\n")
			}
			fmt.Println()

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			if watch {
				done := make(chan struct{})
				go func() {
					<-ctx.Done()
					close(done)
				}()
				go watcher.Watch(file, eng, logger, done)
			}

			srv := server.New(eng, port, logger)
			return srv.Start(ctx)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "path to Q&A data file (csv or json)")
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch file for changes and auto-reload")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func searchCmd() *cobra.Command {
	var (
		file        string
		limit       int
		format      string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search Q&A entries from the command line",
		Long: `Search Q&A entries using fuzzy matching.

Run without a query to open interactive mode with live search.
Run with a query for one-shot results.`,
		Example: `  athena search -f faq.csv                           # interactive mode
  athena search -f faq.csv "how to reset password"   # one-shot search
  athena search -f faq.csv -l 5 "shipping"           # limit results
  athena search -f faq.csv --format json "refund"    # JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := loader.Load(file)
			if err != nil {
				return err
			}
			eng := engine.New(entries)

			// Interactive mode: no query provided or -i flag
			if len(args) == 0 || interactive {
				return runInteractive(eng)
			}

			query := strings.Join(args, " ")
			results := eng.Search(query, limit)

			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			switch format {
			case "json":
				return printJSON(results)
			default:
				return printTable(results, query)
			}
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "path to Q&A data file (csv or json)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "maximum number of results")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table or json")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "force interactive mode")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("athena %s\n", version)
		},
	}
}

func runInteractive(eng *engine.Engine) error {
	m := tui.New(eng)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func printTable(results []engine.SearchResult, query string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tRELEVANCE\tQUESTION\tANSWER\n")
	for i, r := range results {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
			i+1,
			formatRelevance(r.Score, query),
			r.Question,
			truncate(r.Answer, 80),
		)
	}
	return w.Flush()
}

func printJSON(results []engine.SearchResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func formatRelevance(score int, query string) string {
	if len(query) == 0 {
		return "-"
	}
	maxDist := len(query) * 2
	if maxDist == 0 {
		maxDist = 1
	}
	pct := 100 - (score*100)/maxDist
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%d%%", pct)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func banner() string {
	return `
   ___   __  __
  / _ | / /_/ /  ___ ___  ___ _
 / __ |/ __/ _ \/ -_) _ \/ _ '/
/_/ |_|\__/_//_/\__/_//_/\_,_/

  Lightning-fast Q&A search engine
`
}
