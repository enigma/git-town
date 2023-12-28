package sync

import (
	"github.com/git-town/git-town/v11/src/cmd/cmdhelpers"
	"github.com/git-town/git-town/v11/src/git/gitdomain"
	"github.com/git-town/git-town/v11/src/vm/opcode"
)

// BranchesProgram syncs all given branches.
func BranchesProgram(args BranchesProgramArgs) {
	for _, branch := range args.BranchesToSync {
		BranchProgram(branch, args.BranchProgramArgs)
	}
	args.Program.Add(&opcode.CheckoutIfExists{Branch: args.InitialBranch})
	if args.Remotes.HasOrigin() && args.ShouldPushTags && args.IsOnline.Bool() {
		args.Program.Add(&opcode.PushTags{})
	}
	cmdhelpers.Wrap(args.Program, cmdhelpers.WrapOptions{
		RunInGitRoot:             true,
		StashOpenChanges:         args.HasOpenChanges,
		PreviousBranchCandidates: gitdomain.LocalBranchNames{args.PreviousBranch},
	})
}

type BranchesProgramArgs struct {
	BranchProgramArgs
	BranchesToSync gitdomain.BranchInfos
	HasOpenChanges bool
	InitialBranch  gitdomain.LocalBranchName
	PreviousBranch gitdomain.LocalBranchName
	ShouldPushTags bool
}