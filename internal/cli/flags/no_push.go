package flags

import (
	"github.com/git-town/git-town/v16/internal/config/configdomain"
	"github.com/spf13/cobra"
)

const noPushLong = "no-push"

// type-safe access to the CLI arguments of type gitdomain.NoPush
func NoPush() (AddFunc, ReadNoPushFlagFunc) {
	addFlag := func(cmd *cobra.Command) {
		cmd.PersistentFlags().BoolP(noPushLong, "", false, "do not push local branches")
	}
	readFlag := func(cmd *cobra.Command) configdomain.PushBranches {
		value, err := cmd.Flags().GetBool(noPushLong)
		if err != nil {
			panic(err)
		}
		return configdomain.PushBranches(!value)
	}
	return addFlag, readFlag
}

// the type signature for the function that reads the dry-run flag from the args to the given Cobra command
type ReadNoPushFlagFunc func(*cobra.Command) configdomain.PushBranches
