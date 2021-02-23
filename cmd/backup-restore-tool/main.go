package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "backup-restore",
	Short: "Tool for running backup and restore, using S3 as a storage.",
	// SilenceErrors allows us to explicitly log the error returned from rootCmd below.
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("database", "d", "", "Database connection string.")
	rootCmd.PersistentFlags().String("storage-endpoint", "s3.amazonaws.com", "File storage endpoint.")
	rootCmd.PersistentFlags().String("storage-bucket", "", "File storage bucket in which the backup should be stored.")
	rootCmd.PersistentFlags().String("storage-region", "", "Storage region.")
	rootCmd.PersistentFlags().String("storage-object-key", "", "Object key under which backup file will be stored.")

	rootCmd.PersistentFlags().String("storage-access-key", "", "File storage access key id.")
	rootCmd.PersistentFlags().String("storage-secret-key", "", "File storage secret key.")

	rootCmd.PersistentFlags().Bool("storage-tls", true, "Enable storage TLS.")
	rootCmd.PersistentFlags().String("storage-type", "", "Can indicate special type of file storage, ex. bifrost.")

	rootCmd.PersistentFlags().Bool("preserve", false, "If set to true, the backup file will be preserved on disk.")
	rootCmd.PersistentFlags().String("log-file", "", "If set output of underlying commands will be redirected there.")

	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}

func initConfig() {
	bindFlags(rootCmd)

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("BRT")
	viper.AutomaticEnv()
}

// Binds all flags as viper values
func bindFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.PersistentFlags())
	for _, c := range cmd.Commands() {
		bindFlags(c)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Error("command failed")
		os.Exit(1)
	}
}
