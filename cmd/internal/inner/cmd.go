package inner

import "github.com/spf13/cobra"

var (
	cmds = []*cobra.Command{}
)

// Command .
func Command() []*cobra.Command {
	return cmds
}
