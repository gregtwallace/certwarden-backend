package dns01acmesh

import (
	"os"
	"os/exec"
)

// makeCreateCommand creates the command to make a dns record
func (service *Service) makeCreateCommand(dnsRecordName, dnsRecordValue string) *exec.Cmd {
	return service.makeCommand(dnsRecordName, dnsRecordValue, false)
}

// makeDeleteCommand creates the command to delete a dns record
func (service *Service) makeDeleteCommand(dnsRecordName, dnsRecordValue string) *exec.Cmd {
	return service.makeCommand(dnsRecordName, dnsRecordValue, true)
}

// makeCommand makes a command to create or delete a dns record
func (service *Service) makeCommand(dnsRecordName, dnsRecordValue string, delete bool) *exec.Cmd {
	// func name
	funcName := service.dnsHook + "_add"
	if delete {
		funcName = service.dnsHook + "_rm"
	}

	// make args for command
	// `-c`
	args := []string{"-c"}

	// actual command  `source [path] ; [func] [args]`
	args = append(args, "source "+service.shellScriptPath+" ; "+funcName+" "+dnsRecordName+" "+dnsRecordValue)

	// make command
	cmd := exec.Command(service.shellPath, args...)

	// set command environment
	cmd.Env = append(os.Environ(), service.environmentParams.StringSlice()...)

	return cmd
}
