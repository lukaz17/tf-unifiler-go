package exec

import (
	"os/exec"

	"github.com/tforceaio/tf-unifiler-go/extension"
)

type CommandArgs interface {
	Compile() []string
}

func Run(app string, arg CommandArgs) (string, error) {
	args := append([]string{app}, arg.Compile()...)
	logger.Debug().Array("cmd", extension.StringSlice(args)).Msg("Preparing to execute command")
	cmd := exec.Command(app, arg.Compile()...)
	stdout, err := cmd.Output()

	if err != nil {
		logger.Err(err).Msg("Command execution failed")
		return "", err
	}
	logger.Debug().Str("stdout", string(stdout)).Msg("Executed command sucessfully")
	return string(stdout), nil
}
