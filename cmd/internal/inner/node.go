package inner

import "github.com/spf13/cobra"

func init() {
	cmds = append(cmds, nodeCommand())
}

func nodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "inode database",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	cmd.AddCommand(getNodeCommand())

	return cmd
}

func getNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get or list inode infomation",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
