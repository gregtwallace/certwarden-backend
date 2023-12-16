package orders

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// executePostProcessing executes the order's certificate's post processing script
// if the script field is blank, this is a no-op
func (of *orderFulfiller) executePostProcessing(orderID int) error {
	// fetch most up to date order object
	order, err := of.storage.GetOneOrder(orderID)
	if err != nil {
		of.logger.Errorf("failed to fetch order id %d for post processing (%s)", orderID, err)
		return err
	}

	// if no post processing script, done
	if order.Certificate.PostProcessingCommand == "" {
		of.logger.Debugf("skipping post processing of order id %d (no post process command)", orderID)
		return nil
	}

	// if app failed to get suitable shell at startup, post processing is disabled
	if of.shellPath == "" {
		return fmt.Errorf("failed to run post processing of order id %d (no suitable shell was found during app startup)", orderID)
	}

	if order.Pem == nil {
		return errors.New("order pem is nil (should never happen)")
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
			of.logger.Errorf("post processing std err: %s", exitErr.Stderr)
		}

		of.logger.Errorf("post processing error: %s", err)
		return err
	}
	of.logger.Debugf("post processing output: %s", string(result))

	return nil
}
