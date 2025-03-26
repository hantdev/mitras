package cli

import "github.com/spf13/cobra"

// NewHealthCmd returns health check command.
func NewHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health <service>",
		Short: "Health Check",
		Long: "Mitras service Health Check\n" +
			"usage:\n" +
			"\tmitras-cli health <service>",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}
			v, err := sdk.Health(args[0])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logJSONCmd(*cmd, v)
		},
	}
}