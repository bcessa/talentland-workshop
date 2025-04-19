package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/errors"
	xlog "go.bryk.io/pkg/log"
)

var (
	log     xlog.Logger // app main logger
	cfgFile = ""        // configuration file used
	silent  = false     // suppress log output
	appName = "echoctl" // used for ENV variables prefix (uppercase) and home directories
)

var rootCmdDesc = `
Your Application Name.

Here you can enter a more complete description of the purpose
and main features of your app. You can also add links to more
complete documentation or resources.
`

var rootCmd = &cobra.Command{
	Use:           "echoctl",
	Short:         "One-liner description for your application",
	SilenceErrors: true,
	SilenceUsage:  true,
	Long:          rootCmdDesc,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVarP(&silent, "silent", "s", false, "discard log output")
}

// Execute provides the main entry point for the application.
func Execute() {
	defer func() {
		if err := errors.FromRecover(recover()); err != nil {
			exit(err)
		}
	}()
	exit(rootCmd.Execute())
}

// print error details and quit.
func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	if pe := new(errors.Error); errors.As(err, &pe) {
		// print full error with stack trace
		fmt.Printf("%+v", err)
	} else {
		// log simple error message
		log.WithField("error", err).Error("exectution failed")
	}
	os.Exit(1)
}

func initConfig() {
	// ENV
	viper.SetEnvPrefix(appName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set configuration file
	if cfgFile == "" {
		// discovery mechanism
		//  - /etc
		//  - /${HOME}
		//  - /`pwd`
		viper.SetConfigName("config")
		viper.AddConfigPath(filepath.Join("/etc", appName))
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(filepath.Join(home, appName))
			viper.AddConfigPath(filepath.Join(home, fmt.Sprintf(".%s", appName)))
		}
		viper.AddConfigPath(".")
	} else {
		// directly provided
		viper.SetConfigFile(cfgFile)
	}

	// Setup main logger
	if silent {
		log = xlog.Discard()
	} else {
		log = xlog.WithCharm(xlog.CharmOptions{
			WithColor: true,
		})
	}

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil && viper.ConfigFileUsed() != "" {
		log.WithField("file", viper.ConfigFileUsed()).Error("failed to load configuration file")
	} else {
		if cf := viper.ConfigFileUsed(); cf != "" {
			log.WithField("file", cf).Debug("configuration loaded")
			log.Debug("start watching for configuration updates")
			viper.WatchConfig()
		}
	}
}
