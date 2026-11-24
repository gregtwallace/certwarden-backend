package dns01manual

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
func (service *Service) makeCommand(dnsRecordName, dnsRecordValue string, doDelete bool) *exec.Cmd {
	// create or delete?
	scriptPath := service.createScriptPath
	if doDelete {
		scriptPath = service.deleteScriptPath
	}

	// make args for command
	args := []string{
		// 0 - script name (e.g. /path/to/script.sh)
		scriptPath,
		// 1 - RecordName (e.g. _acme-challenge.www.example.com)
		dnsRecordName,
		// 2 - RecordValue (e.g. XKrxpRBosdIKFzxW_CT3KLZNf6q0HG9i01zxXp5CPBs)
		dnsRecordValue,
	}

	// make command
	cmd := exec.Command(service.shellPath, args...)

	// set command environment
	cmd.Env = append(os.Environ(), service.environmentParams.StringSlice()...)

	return cmd
}
