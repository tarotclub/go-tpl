package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tarotclub/go-tpl/internal/config"
	"github.com/tarotclub/go-tpl/internal/logger"
)

var (
	cfgFile string
	log     zerolog.Logger
	cfg     *config.Config
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "go-tpl",
	Short: "A Go project template",
	Long:  "go-tpl is a Go project template that demonstrates cobra, viper, and zerolog integration.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}
		log = logger.New(cfg.Log.Level)
		log.Debug().Str("config_file", viper.ConfigFileUsed()).Msg("configuration loaded")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags, then
// executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
}
