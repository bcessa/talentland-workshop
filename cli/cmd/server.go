package cmd

import (
	"context"
	"os"
	"sync"
	"syscall"

	"github.com/bcessa/echo-service/handler"
	"github.com/bcessa/echo-service/internal"
	"github.com/bcessa/echo-service/internal/dx"
	dxOtel "github.com/bcessa/echo-service/internal/dx/modules/otel"
	dxRpc "github.com/bcessa/echo-service/internal/dx/modules/rpc"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	viperUtils "go.bryk.io/pkg/cli/viper"
	"go.bryk.io/pkg/net/rpc"
	otelSdk "go.bryk.io/pkg/otel/sdk"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a server instance to handle incoming requests",
	RunE:  runServer,
}

// module registry.
var reg *dx.Registry

func init() {
	// register required dependencies
	reg = dx.NewRegistry(appName)
	reg.Add(new(dxRpc.Module))
	reg.Add(new(dxOtel.Module))
	params := reg.Get("rpc").Flags(appName)
	if err := cli.SetupCommandParams(serverCmd, params); err != nil {
		panic(err)
	}
	if err := viperUtils.BindFlags(serverCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(serverCmd)
}

// nolint: funlen
func runServer(_ *cobra.Command, _ []string) (err error) {
	var (
		wg         = sync.WaitGroup{}       // background tasks handler
		telemetry  *otelSdk.Instrumentation // telemetry implementation
		server     *rpc.Server              // rpc server
		svcHandler *handler.ServiceOperator // service handler
	)

	// wait for "start" signals
	startSig := make(chan struct{}, 1)

	// wait for "reload" signals
	reloadSig := cli.SignalsHandler([]os.Signal{syscall.SIGHUP})
	viper.OnConfigChange(func(_ fsnotify.Event) {
		// fake a "SIGHUP" on configuration changes
		reloadSig <- syscall.SIGHUP
	})

	// wait for "close" signals
	closeSig := cli.SignalsHandler([]os.Signal{
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	})

	// start server
	startSig <- struct{}{}

signals:
	for {
		select {
		case <-startSig:
			// load application settings
			if err := reg.Load(viper.GetViper()); err != nil {
				return err
			}

			// telemetry instrumentation
			log.WithField("module", "otel").Debug("loading module")
			obOpts := []otelSdk.Option{}
			_ = reg.Get("otel").Customize(&obOpts)
			if len(obOpts) > 0 {
				obOpts = append(obOpts, otelSdk.WithBaseLogger(log))
				telemetry, err = otelSdk.Setup(obOpts...)
				if err != nil {
					return err
				}
			}

			// rpc server settings
			log.WithField("module", "rpc").Debug("loading module")
			serverOptions := []rpc.ServerOption{}
			if err := reg.Get("rpc").Customize(&serverOptions); err != nil {
				return err
			}

			// add build information as server middleware
			buildMW := rpc.WithGatewayMiddleware(internal.BuildDetails().Middleware())
			serverOptions = append(serverOptions, rpc.WithHTTPGatewayOptions(buildMW))

			// service handler
			log.Info("starting service handler")
			svcHandler, err = handler.New()
			if err != nil {
				return err
			}

			// start server
			log.Info("starting server")
			serverOptions = append(serverOptions, rpc.WithServiceProvider(svcHandler.RPC()))
			server, err = rpc.NewServer(serverOptions...)
			if err != nil {
				return err
			}
			ready := make(chan bool)
			wg.Add(1)
			go func() {
				_ = server.Start(ready)
				wg.Done()
			}()
			<-ready
			log.Info("server is ready and waiting for requests")
		case <-reloadSig:
			log.Info("reloading server")
			_ = server.Stop(true) // gracefully stop server
			if telemetry != nil {
				// drain telemetry operator
				telemetry.Flush(context.Background())
			}
			wg.Wait()               // wait for background tasks
			_ = svcHandler.Reload() // reload service handler
			startSig <- struct{}{}  // signal server to start again
		case <-closeSig:
			log.Info("closing server")
			// stop signal processing and continue to regular shutdown process
			break signals
		}
	}

	// shutdown process
	if err = svcHandler.Close(); err != nil {
		log.WithField("error", err.Error()).Error("service handler close")
	}
	err = server.Stop(true) // gracefully stop server
	if telemetry != nil {
		// drain telemetry operator
		telemetry.Flush(context.Background())
	}
	wg.Wait()        // wait for background tasks
	close(startSig)  // clean up "start" signals channel
	close(reloadSig) // clean up "reload" signals channel
	close(closeSig)  // clean up "close" signals channel
	return err       // return final result
}
