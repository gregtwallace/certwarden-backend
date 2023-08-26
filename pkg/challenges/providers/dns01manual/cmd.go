package dns01manual

import (
	"os"
	"os/exec"
)

// makeCreateCommand creates the command to make a dns record
func (service *Service) makeCreateCommand(domainName, resourceName, resourceContent string) *exec.Cmd {
	return service.makeCommand(domainName, resourceName, resourceContent, false)
}

// makeDeleteCommand creates the command to delete a dns record
func (service *Service) makeDeleteCommand(domainName, resourceName, resourceContent string) *exec.Cmd {
	return service.makeCommand(domainName, resourceName, resourceContent, true)
}

// makeCommand makes a command to create or delete a dns record
func (service *Service) makeCommand(domainName, resourceName, resourceContent string, delete bool) *exec.Cmd {
	// create or delete?
	scriptPath := service.createScriptPath
	if delete {
		scriptPath = service.deleteScriptPath
	}

	// make args for command
	// 0 - script name (e.g. /path/to/script.sh)
	args := []string{scriptPath}

	// 1 - Domain (e.g. example.com)
	args = append(args, domainName)

	// 2 - RecordName (e.g. _acme-challenge.www.example.com)
	args = append(args, resourceName)

	// 3 - RecordValue (e.g. XKrxpRBosdIKFzxW_CT3KLZNf6q0HG9i01zxXp5CPBs)
	args = append(args, resourceContent)

	// make command
	cmd := exec.Command(service.shellPath, args...)

	// set command environment
	cmd.Env = append(os.Environ(), service.environmentVars...)

	return cmd
}
