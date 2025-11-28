package orders

import (
	"certwarden-backend/pkg/datatypes/environment"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// doScriptOrBinaryPost executes the certificate's post processing command. if the cert
// does not have a command, this is a no-op
func (j *postProcessJob) doScriptOrBinaryPostProcess(order Order, workerID int) {
	// no-op if no command
	if order.Certificate.PostProcessingCommand == "" {
		j.service.logger.Debugf("orders: post processing worker %d: order %d: skipping command (cert does not have a command to run) (cert: %d, cn: %s)", workerID, order.ID, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	j.service.logger.Infof("orders: post processing worker %d: order %d: attempting to run command (cert: %d, cn: %s)", workerID, order.ID, order.Certificate.ID, order.Certificate.Subject)

	// nil checks
	if order.Pem == nil {
		err := fmt.Errorf("orders: post processing worker %d: order %d: command failed: order pem is nil (should never happen)", workerID, order.ID)
		j.service.logger.Error(err)
		return
	}
	if order.FinalizedKey == nil {
		err := fmt.Errorf("orders: post processing worker %d: order %d: command failed: finalized key no longer exists", workerID, order.ID)
		j.service.logger.Error(err)
		return
	}

	// user specified environment can have placeholders for certain values (so user can set
	// their own name for the environment variable)
	// {{PRIVATE_KEY_NAME}}					= the `Name` of the private key used to finalize the order
	// {{PRIVATE_KEY_PEM}}					= the pem of the private key
	// {{PRIVATE_KEY_API_KEY}}			= the api key of the private key
	// {{CERTIFICATE_NAME}}					= the `Name` of the certificate
	// {{CERTIFICATE_PEM}}					= the pem of the complete certificate chain for the order
	// {{CERTIFICATE_API_KEY}}			= the api key of the certificate
	// {{CERTIFICATE_COMMON_NAME}}	= the common name of the certificate

	// make Params (which sanitizes the env params and handles things like removing quotes)
	envParams, invalidParams := environment.NewParams(order.Certificate.PostProcessingEnvironment)
	if len(invalidParams) > 0 {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: %s are not properly formatted environment param(s), they will be skipped", workerID, order.ID, invalidParams)
	}

	// make environ from Params and update placeholders with proper values
	environ := []string{}
	for key, val := range envParams.KeyValMap() {
		switch upperVal := strings.ToUpper(val); upperVal {
		case "{{PRIVATE_KEY_NAME}}":
			val = order.FinalizedKey.Name

		case "{{PRIVATE_KEY_PEM}}":
			val = order.FinalizedKey.Pem

		case "{{PRIVATE_KEY_API_KEY}}":
			val = order.FinalizedKey.ApiKey

		case "{{CERTIFICATE_NAME}}":
			val = order.Certificate.Name

		case "{{CERTIFICATE_PEM}}":
			val = *order.Pem

		case "{{CERTIFICATE_API_KEY}}":
			val = order.Certificate.ApiKey

		case "{{CERTIFICATE_COMMON_NAME}}":
			val = order.Certificate.Subject

		default:
			// no-op - user specified some other value
		}

		// append to environment
		environ = append(environ, key+"="+val)
	}

	// open and read (up to) the first 512 bytes of post processing script/binary to decide if it is binary or not
	// and also check if the file has a shebang
	f, err := os.Open(order.Certificate.PostProcessingCommand)
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: script/binary failed to open: %s", workerID, order.ID, err)
		return
	}
	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: script/binary failed to stat: %s", workerID, order.ID, err)
		return
	}

	bufLen := 512
	if fInfo.Size() < 512 {
		bufLen = int(fInfo.Size())
	}
	firstBytes := make([]byte, bufLen)

	_, err = io.ReadFull(f, firstBytes)
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: script/binary failed to read: %s", workerID, order.ID, err)
		return
	}

	// run binary or shebang file directly
	cmd := &exec.Cmd{}
	if http.DetectContentType(firstBytes) == "application/octet-stream" || strings.HasPrefix(string(firstBytes), "#!") {
		// binary found
		cmd = exec.Command(order.Certificate.PostProcessingCommand)

	} else {
		// try to run as script if it wasn't an octet-stream and didn't have shebang
		// if app failed to get suitable default shell at startup, post processing will fail
		if j.service.defaultShellPath == "" {
			j.service.logger.Errorf("orders: post processing worker %d: order %d: commaind failed to run post processing script (no suitable default shell was found during startup)", workerID, order.ID)
			return
		}

		// make args for command
		// 0 - script name (e.g. /path/to/script.sh)
		args := []string{order.Certificate.PostProcessingCommand}

		// make command
		cmd = exec.Command(j.service.defaultShellPath, args...)
	}

	// set command environment (default OS + environ from above)
	cmd.Env = append(os.Environ(), environ...)

	// run command
	result, err := cmd.Output()
	j.service.logger.Debugf("orders: post processing worker %d: order %d: command output: %s", workerID, order.ID, string(result))
	if err != nil {
		// try to get stderr and log it too
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			j.service.logger.Errorf("orders: post processing worker %d: order %d: command std err: %s", workerID, order.ID, exitErr.Stderr)
		}

		j.service.logger.Errorf("orders: post processing worker %d: order %d: command failed: error: %s", workerID, order.ID, err)
		return
	}

	j.service.logger.Infof("orders: post processing worker %d: order %d: command completed", workerID, order.ID)
}
