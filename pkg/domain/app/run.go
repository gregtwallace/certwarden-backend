package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const maxShutdownTime = 30 * time.Second

// RunLeGoAPI starts the application and also contains restart logic
// in the event the app calls for a restart after termination
func RunLeGoAPI() {
	// run LeGo
	restart := run()

	// if LeGo called restart, execute self before exit
	if restart {
		// get path, args, and environment for execution
		self, err := os.Executable()
		if err != nil {
			os.Exit(1)
		}
		args := os.Args
		env := os.Environ()

		// windows does not support syscall.Exec([...]).
		if runtime.GOOS == "windows" {
			cmd := exec.Command(self, args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Env = env

			// run
			err = cmd.Run()
		} else {
			// run non-Windows
			err = syscall.Exec(self, args, env)
		}

		// err check from either run command
		if err != nil {
			os.Exit(1)
		}
	}

	os.Exit(0)
}

// run starts an instance of the LeGo application
func run() (restart bool) {
	// create the app
	app, err := create()
	if err != nil {
		app.logger.Errorf("failed to create app (%s)", err)
		os.Exit(1)
		return
	}
	// defer storage close, logger close, and app nil
	defer func() {
		// close storage
		err := app.storage.Close()
		if err != nil {
			app.logger.Errorf("error closing storage: %s", err)
		} else {
			app.logger.Info("storage closed")
		}

		// flush and close logger
		app.logger.Debug("flushing (syncing) logger and closing underlying log file")
		// log if trying to restart, before closing logger
		if app.restart {
			app.logger.Info("restarting lego")
		} else {
			app.logger.Info("lego shutdown complete")
		}
		app.logger.syncAndClose()

		// nil app
		app = nil
	}()

	// start pprof if enabled
	if app.config.EnablePprof != nil && *app.config.EnablePprof {
		err = app.startPprof()
		if err != nil {
			app.logger.Errorf("failed to start pprof (%s), exiting", err)
			os.Exit(1)
		}
	}

	// http server config
	srv := &http.Server{
		Addr:         app.config.httpServAddress(),
		Handler:      app.routes(),
		IdleTimeout:  httpServerIdleTimeout,
		ReadTimeout:  httpServerReadTimeout,
		WriteTimeout: httpServerWriteTimeout,
	}

	// var for redirect server (if needed)
	redirectSrv := &http.Server{}

	// configure and launch https if app succesfully got a cert
	if app.httpsCert != nil {
		// https server config
		srv.Addr = app.config.httpsServAddress()
		srv.TLSConfig = app.tlsConf()

		// configure and launch http redirect server
		if *app.config.EnableHttpRedirect {
			redirectSrv = &http.Server{
				Addr: app.config.httpServAddress(),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// remove port (if present) to get request hostname alone (since changing port)
					hostName, _, _ := strings.Cut(r.Host, ":")

					// build redirect address
					newAddr := "https://" + hostName + ":" + strconv.Itoa(*app.config.HttpsPort) + r.RequestURI

					http.Redirect(w, r, newAddr, http.StatusTemporaryRedirect)
				}),
				IdleTimeout:  httpServerIdleTimeout,
				ReadTimeout:  httpServerReadTimeout,
				WriteTimeout: httpServerWriteTimeout,
			}

			app.logger.Infof("starting http redirect bound to %s", app.config.httpServAddress())

			// create listener for web server
			ln1, err := net.Listen("tcp", app.config.httpServAddress())
			if err != nil {
				app.logger.Errorf("http redirect server cannot bind to %s (%s), exiting", app.config.httpServAddress(), err)
				os.Exit(1)
			}

			// start server
			app.shutdownWaitgroup.Add(1)
			go func() {
				defer func() { _ = ln1.Close }()
				err := redirectSrv.Serve(ln1)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					app.logger.Errorf("http redirect server returned error (%s)", err)
				}
				app.logger.Info("http redirect server shutdown complete")
				app.shutdownWaitgroup.Done()
			}()
		}

		// launch https
		app.logger.Infof("starting lego-certhub (https) bound to %s", app.config.httpsServAddress())

		// create listener for web server
		ln2, err := net.Listen("tcp", app.config.httpsServAddress())
		if err != nil {
			app.logger.Errorf("lego-certhub (https) server cannot bind to %s (%s), exiting", app.config.httpServAddress(), err)
			os.Exit(1)
		}

		// start server
		app.shutdownWaitgroup.Add(1)
		go func() {
			defer func() { _ = ln2.Close }()
			err := srv.ServeTLS(ln2, "", "")
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.logger.Errorf("lego-certhub (https) server returned error (%s)", err)
			}
			app.logger.Info("https server shutdown complete")
			app.shutdownWaitgroup.Done()
		}()

	} else {
		// if https failed, launch localhost only http server
		app.logger.Warnf("starting insecure lego-certhub (http) bound to %s", app.config.httpServAddress())

		// create listener for web server
		ln3, err := net.Listen("tcp", app.config.httpServAddress())
		if err != nil {
			app.logger.Errorf("insecure lego-certhub (http) server cannot bind to %s (%s), exiting", app.config.httpServAddress(), err)
			os.Exit(1)
		}

		// start server
		app.shutdownWaitgroup.Add(1)
		go func() {
			defer func() { _ = ln3.Close }()
			err := srv.Serve(ln3)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.logger.Errorf("insecure lego-certhub (http) server returned error (%s)", err)
			}
			app.logger.Info("http server shutdown complete")
			app.shutdownWaitgroup.Done()
		}()
	}

	// shutdown logic
	// wait for shutdown context to signal
	<-app.shutdownContext.Done()

	// shutdown the main web server (and redirect server)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		err = srv.Shutdown(ctx)
		if err != nil {
			app.logger.Errorf("error shutting down http(s) server")
		}
	}()

	if app.httpsCert != nil && *app.config.EnableHttpRedirect {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
			defer cancel()

			err = redirectSrv.Shutdown(ctx)
			if err != nil {
				app.logger.Errorf("error shutting down http redirect server")
			}
		}()
	}

	// wait for each component/service to shutdown
	// but also implement a maxWait chan to force close (panic)
	maxWait := 2 * time.Minute
	waitChan := make(chan struct{})

	// close wait chan when wg finishes waiting
	go func() {
		defer close(waitChan)
		app.shutdownWaitgroup.Wait()
	}()

	select {
	case <-waitChan:
		// continue, normal
	case <-time.After(maxWait):
		// timed out
		app.logger.Panic("graceful shutdown of component(s) failed due to time out, forcing shutdown")
	}

	return app.restart
}
