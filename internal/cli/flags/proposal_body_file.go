package flags

import (
	"github.com/git-town/git-town/v16/internal/git/gitdomain"
	"github.com/spf13/cobra"
)

const (
	bodyFileLong  = "body-file" // long form of the "body-file" CLI flag
	bodyFileShort = "f"         // short form of the "body-file" CLI flag
)

// provides type-safe access to the CLI arguments of type gitdomain.ProposalBodyFile
func ProposalBodyFile() (AddFunc, ReadProposalBodyFileFlagFunc) {
	addFlag := func(cmd *cobra.Command) {
		cmd.PersistentFlags().StringP(bodyFileLong, bodyFileShort, "", "Read the proposal body from the given file (use \"-\" to read from STDIN)")
	}
	readFlag := func(cmd *cobra.Command) gitdomain.ProposalBodyFile {
		value, err := cmd.Flags().GetString(bodyFileLong)
		if err != nil {
			panic(err)
		}
		return gitdomain.ProposalBodyFile(value)
	}
	return addFlag, readFlag
}

// reads gitdomain.ProposalBodyFile from the CLI args
type ReadProposalBodyFileFlagFunc func(*cobra.Command) gitdomain.ProposalBodyFile
