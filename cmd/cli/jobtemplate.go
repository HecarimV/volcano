package main

import (
	"github.com/spf13/cobra"
	"volcano.sh/volcano/pkg/cli/job"
)

func buildJobTemplateCmd() *cobra.Command {
	jobTemplateCmd := &cobra.Command{
		Use:   "template",
		Short: "vcctl command line operation job template",
	}

	jobTemplateRunCmd := &cobra.Command{
		Use:   "run",
		Short: "run job by parameters from the command line",
		Run: func(cmd *cobra.Command, args []string) {
			checkError(cmd, job.RunJobTemplate())
		},
	}
	job.InitTemplateRunFlags(jobTemplateRunCmd)
	jobTemplateCmd.AddCommand(jobTemplateRunCmd)

	jobTemplateGenerateCmd := &cobra.Command{
		Use:   "generate",
		Short: "generate jobTemplate by parameters from the command line",
		Run: func(cmd *cobra.Command, args []string) {
			checkError(cmd, job.GenerateJobTemplate())
		},
	}
	job.InitTemplateGenerateFlags(jobTemplateGenerateCmd)
	jobTemplateCmd.AddCommand(jobTemplateGenerateCmd)

	return jobTemplateCmd
}
