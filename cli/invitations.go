package cli

import (
	mitrassdk "github.com/hantdev/mitras/pkg/sdk"
	"github.com/spf13/cobra"
)

var cmdInvitations = []cobra.Command{
	{
		Use:   "send <user_id> <domain_id> <role_id> <user_auth_token>",
		Short: "Send invitation",
		Long: "Send invitation to user\n" +
			"For example:\n" +
			"\tmitras-cli invitations send 39f97daf-d6b6-40f4-b229-2697be8006ef 4ef09eff-d500-4d56-b04f-d23a512d6f2a ba4c904c-e6d4-4978-9417-1694aac6793e $USER_AUTH_TOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 4 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}
			inv := mitrassdk.Invitation{
				InviteeUserID: args[0],
				DomainID:      args[1],
				RoleID:        args[2],
			}
			if err := sdk.SendInvitation(inv, args[3]); err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logOKCmd(*cmd)
		},
	},
	{
		Use:   "get [all | <user_id> <domain_id> ] <user_auth_token>",
		Short: "Get invitations",
		Long: "Get invitations\n" +
			"Usage:\n" +
			"\tmitras-cli invitations get all <user_auth_token> - lists all invitations\n" +
			"\tmitras-cli invitations get all <user_auth_token> --offset <offset> --limit <limit> - lists all invitations with provided offset and limit\n" +
			"\tmitras-cli invitations get <user_id> <domain_id> <user_auth_token> - shows invitation by user id and domain id\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 && len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			pageMetadata := mitrassdk.PageMetadata{
				Identity: Identity,
				Offset:   Offset,
				Limit:    Limit,
			}
			if args[0] == all {
				l, err := sdk.Invitations(pageMetadata, args[1])
				if err != nil {
					logErrorCmd(*cmd, err)
					return
				}
				logJSONCmd(*cmd, l)
				return
			}
			u, err := sdk.Invitation(args[0], args[1], args[2])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logJSONCmd(*cmd, u)
		},
	},
	{
		Use:   "accept <domain_id> <user_auth_token>",
		Short: "Accept invitation",
		Long: "Accept invitation to domain\n" +
			"Usage:\n" +
			"\tmitras-cli invitations accept 39f97daf-d6b6-40f4-b229-2697be8006ef $USERTOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			if err := sdk.AcceptInvitation(args[0], args[1]); err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logOKCmd(*cmd)
		},
	},
	{
		Use:   "reject <domain_id> <user_auth_token>",
		Short: "Reject invitation",
		Long: "Reject invitation to domain\n" +
			"Usage:\n" +
			"\tmitras-cli invitations reject 39f97daf-d6b6-40f4-b229-2697be8006ef $USER_AUTH_TOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			if err := sdk.RejectInvitation(args[0], args[1]); err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logOKCmd(*cmd)
		},
	},
	{
		Use:   "delete <user_id> <domain_id> <user_auth_token>",
		Short: "Delete invitation",
		Long: "Delete invitation\n" +
			"Usage:\n" +
			"\tmitras-cli invitations delete 39f97daf-d6b6-40f4-b229-2697be8006ef 4ef09eff-d500-4d56-b04f-d23a512d6f2a $USERTOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			if err := sdk.DeleteInvitation(args[0], args[1], args[2]); err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logOKCmd(*cmd)
		},
	},
}

// NewInvitationsCmd returns invitations command.
func NewInvitationsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "invitations [send | get | accept | delete]",
		Short: "Invitations management",
		Long:  `Invitations management to send, get, accept and delete invitations`,
	}

	for i := range cmdInvitations {
		cmd.AddCommand(&cmdInvitations[i])
	}

	return &cmd
}