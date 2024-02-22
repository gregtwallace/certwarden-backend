package orders

import (
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
		j.service.logger.Debugf("post processing worker %d: order %d: skipping command (cert does not have a command to run) (cert: %d, cn: %s)", workerID, order.ID, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	j.service.logger.Infof("post processing worker %d: order %d: attempting to run command (cert: %d, cn: %s)", workerID, order.ID, order.Certificate.ID, order.Certificate.Subject)

	// nil checks
	if order.Pem == nil {
		err := fmt.Errorf("post processing worker %d: order %d: command failed: order pem is nil (should never happen)", workerID, order.ID)
		j.service.logger.Error(err)
		return
	}
	if order.FinalizedKey == nil {
		err := fmt.Errorf("post processing worker %d: order %d: command failed: finalized key no longer exists", workerID, order.ID)
		j.service.logger.Error(err)
		return
	}

	// user specified environment can have placeholders for certain values (so user can set
	// their own name for the environment variable)
	// {{PRIVATE_KEY_NAME}}					= the name of the private key used to finalize the order, as defined in LeGo
	// {{PRIVATE_KEY_PEM}}					= the pem of the private key
	// {{CERTIFICATE_NAME}}					= the name of the certificate as defined in LeGo
	// {{CERTIFICATE_PEM}}					= the pem of the complete certificate chain for the order
	// {{CERTIFICATE_COMMON_NAME}}	= the common name of the certificate

	environ := []string{}

	// update placeholders and set user values
	for i := range order.Certificate.PostProcessingEnvironment {
		envItemSplit := strings.SplitN(order.Certificate.PostProcessingEnvironment[i], "=", -1)

		// only append properly formatted vars (i.e. VAR=someValue; len after = split == 2)
		if len(envItemSplit) == 2 {
			// do replacements if user val is a value that should be replaced
			switch val := strings.ToUpper(envItemSplit[1]); val {
			case "{{PRIVATE_KEY_NAME}}":
				envItemSplit[1] = order.FinalizedKey.Name

			case "{{PRIVATE_KEY_PEM}}":
				envItemSplit[1] = order.FinalizedKey.Pem

			case "{{CERTIFICATE_NAME}}":
				envItemSplit[1] = order.Certificate.Name

			case "{{CERTIFICATE_PEM}}":
				envItemSplit[1] = *order.Pem

			case "{{CERTIFICATE_COMMON_NAME}}":
				envItemSplit[1] = order.Certificate.Subject

			default:
				// no-op - user specified some other value
			}

			// append to environment
			environ = append(environ, envItemSplit[0]+"="+envItemSplit[1])
		} else {
			j.service.logger.Errorf("post processing worker %d: order %d: %s is not a properly formatted environment variable, it will be skipped", workerID, order.ID, order.Certificate.PostProcessingEnvironment[i])
		}
	}

	// open and read (up to) the first 512 bytes of post processing script/binary to decide if it is binary or not
	f, err := os.Open(order.Certificate.PostProcessingCommand)
	if err != nil {
		j.service.logger.Errorf("post processing worker %d: order %d: script/binary failed to open: %w", workerID, order.ID, err)
		return
	}
	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		j.service.logger.Errorf("post processing worker %d: order %d: script/binary failed to stat: %w", workerID, order.ID, err)
		return
	}

	bufLen := 512
	if fInfo.Size() < 512 {
		bufLen = int(fInfo.Size())
	}
	firstBytes := make([]byte, bufLen)

	_, err = io.ReadFull(f, firstBytes)
	if err != nil {
		j.service.logger.Errorf("post processing worker %d: order %d: script/binary failed to read: %w", workerID, order.ID, err)
		return
	}

	// check if the file is binary and run it directly if so
	cmd := &exec.Cmd{}
	if http.DetectContentType(firstBytes) == "application/octet-stream" {
		// binary found
		cmd = exec.Command(order.Certificate.PostProcessingCommand)

	} else {
		// try to run as script if it wasn't an octet-stream
		// if app failed to get suitable shell at startup, post processing is disabled
		if j.service.shellPath == "" {
			j.service.logger.Errorf("post processing worker %d: order %d: commaind failed to run post processing script (no suitable shell was found during lego startup)", workerID, order.ID)
			return
		}

		// make args for command
		// 0 - script name (e.g. /path/to/script.sh)
		args := []string{order.Certificate.PostProcessingCommand}

		// make command
		cmd = exec.Command(j.service.shellPath, args...)
	}

	// set command environment (default OS + environ from above)
	cmd.Env = append(os.Environ(), environ...)

	// run command
	result, err := cmd.Output()
	j.service.logger.Debugf("post processing worker %d: order %d: command output: %s", workerID, order.ID, string(result))
	if err != nil {
		// try to get stderr and log it too
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			j.service.logger.Errorf("post processing worker %d: order %d: command std err: %s", workerID, order.ID, exitErr.Stderr)
		}

		j.service.logger.Errorf("post processing worker %d: order %d: command failed: error: %s", workerID, order.ID, err)
		return
	}

	j.service.logger.Infof("post processing worker %d: order %d: command completed", workerID, order.ID)
}
