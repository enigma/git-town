package opcodes

import "github.com/git-town/git-town/v13/src/vm/shared"

// AbortRebase represents aborting on ongoing merge conflict.
// This opcode is used in the abort scripts for Git Town commands.
type AbortRebase struct {
	undeclaredOpcodeMethods
}

func (self *AbortRebase) Run(args shared.RunArgs) error {
	return args.Runner.Frontend.AbortRebase()
}
