package orders

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// executePostProcessing executes the order's certificate's post processing script
// if the script field is blank, this is a no-op
func (of *orderFulfiller) executePostProcessing(order Order) error {
	// if no post processing script, done
	if order.Certificate.PostProcessingCommand == "" {
		of.logger.Debugf("post processing %d: skipping (no post process command)", order.ID)
		return nil
	}

	of.logger.Infof("post processing %d: attempting to run (cert: %d, cn: %s)", order.ID, order.Certificate.ID, order.Certificate.Subject)

	// if app failed to get suitable shell at startup, post processing is disabled
	if of.shellPath == "" {
		err := fmt.Errorf("post processing %d: failed to run post processing (no suitable shell was found during app startup)", order.ID)
		of.logger.Error(err)
		return err
	}

	// nil checks
	if order.Pem == nil {
		err := fmt.Errorf("post processing %d: order pem is nil (should never happen)", order.ID)
		of.logger.Error(err)
		return err
	}
	if order.FinalizedKey == nil {
		err := fmt.Errorf("post processing %d: failed to run post processing (finalized key no longer exists)", order.ID)
		of.logger.Error(err)
		return err
	}

	// make environment; certain values are always automatically added
	// LEGO_PRIVATE_KEY_NAME				= the name of the private key used to finalize the order, as defined in LeGo
	// LEGO_PRIVATE_KEY_PEM					= the pem of the private key
	// LEGO_CERTIFICATE_NAME				= the name of the certificate as defined in LeGo
	// LEGO_CERTIFICATE_PEM					= the pem of the complete certificate chain for the order
	// LEGO_CERTIFICATE_COMMON_NAME = the common name of the certificate
	environ := []string{}

	// LEGO_PRIVATE_KEY_NAME
	environ = append(environ, fmt.Sprintf("LEGO_PRIVATE_KEY_NAME=%s", order.FinalizedKey.Name))

	// LEGO_PRIVATE_KEY_PEM
	environ = append(environ, fmt.Sprintf("LEGO_PRIVATE_KEY_PEM=%s", order.FinalizedKey.Pem))

	// LEGO_CERTIFICATE_NAME
	environ = append(environ, fmt.Sprintf("LEGO_CERTIFICATE_NAME=%s", order.Certificate.Name))

	// LEGO_CERTIFICATE_PEM
	environ = append(environ, fmt.Sprintf("LEGO_CERTIFICATE_PEM=%s", *order.Pem))

	// LEGO_CERTIFICATE_COMMON_NAME
	environ = append(environ, fmt.Sprintf("LEGO_CERTIFICATE_COMMON_NAME=%s", order.Certificate.Subject))

	// add custom values from user
	environ = append(environ, order.Certificate.PostProcessingEnvironment...)

	// make args for command
	// 0 - script name (e.g. /path/to/script.sh)
	args := []string{order.Certificate.PostProcessingCommand}

	// make command
	cmd := exec.Command(of.shellPath, args...)

	// set command environment (default OS + environ from above)
	cmd.Env = append(os.Environ(), environ...)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		// try to get stderr and log it too
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			of.logger.Errorf("post processing %d: std err: %s", order.ID, exitErr.Stderr)
		}

		of.logger.Errorf("post processing %d: error: %s", order.ID, err)
		return err
	}
	of.logger.Debugf("post processing %d: output: %s", order.ID, string(result))

	of.logger.Infof("post processing %d: completed", order.ID)

	return nil
}
