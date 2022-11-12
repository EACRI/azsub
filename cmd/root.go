package cmd

import (
	"azsub/azsub"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func logIfErr(err error) {
	if err != nil {
		log.Errorln(err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "azsub",
	Short: "azsub submits containerized tasks to azure batch",
	Long: `
azsub is a simplified CLI for submitting containerized tasks to azure batch
and syncing outputs to an azure blob storage`,
	Run: func(cmd *cobra.Command, args []string) {

		flags := cmd.Flags()
		image, err := flags.GetString("image")
		logIfErr(err)

		storageAccountName, err := flags.GetString("storage-account-name")
		logIfErr(err)

		prefix, err := flags.GetString("container-prefix")
		logIfErr(err)

		local, err := flags.GetBool("local")
		logIfErr(err)

		command, err := flags.GetString("command")
		logIfErr(err)

		script, err := flags.GetString("script")
		logIfErr(err)

		var task azsub.AzSubTask
		if command != "" {
			if script != "" {
				log.Errorln(errors.New("error: script and command are mutally exclusive"))
			}
			task = azsub.AzSubTask{Type: azsub.CommandTypeTask, Task: command}
		} else if script != "" {
			task = azsub.AzSubTask{Type: azsub.ScriptTypeTask, Task: script}
		}

		// entrypoint to run the main submission logic
		azsub.NewAzsub().
			WithStorage(storageAccountName, prefix).
			WithImage(image).
			WithLocal(local).
			WithTask(task).
			Run()

	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	// batch account configuration flags
	rootCmd.Flags().String("batch-account-url", "", "azure batch account url, https://${name}.${region}.batch.azure.com")
	rootCmd.MarkFlagRequired("batch-account-url")
	rootCmd.Flags().String("batch-account-key", "", "azure batch account key")
	rootCmd.Flags().StringP("image", "i", "ubuntu", "container image url for running the job")

	// storage account configuration flags
	rootCmd.Flags().String("storage-account-name", "", "azure storage account name")
	rootCmd.Flags().String("storage-account-key", "", "azure storage account key")
	rootCmd.Flags().String("container-prefix", "", "storage account container prefix to put results")

	// task specification flags
	rootCmd.Flags().Bool("local", false, "run the job or script locally")
	rootCmd.Flags().StringP("command", "c", "", "command to submit to the batch task")
	rootCmd.Flags().StringP("script", "s", "", "shell script to submit to the batch task")
}
