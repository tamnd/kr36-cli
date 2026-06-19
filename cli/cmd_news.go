package cli

import (
	"github.com/spf13/cobra"
)

// newsCmd returns the news command.
func (a *App) newsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "news",
		Short: "List the latest articles from 36kr",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching %d articles from 36kr RSS...", n)
			arts, err := a.client.News(cmd.Context(), n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(arts, len(arts))
		},
	}
}
