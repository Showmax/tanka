package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-clix/cli"
	"github.com/posener/complete"

	"github.com/grafana/tanka/pkg/process"
	"github.com/grafana/tanka/pkg/tanka"
	"github.com/grafana/tanka/pkg/term"
)

// special exit codes for tk diff
const (
	// no changes
	ExitStatusClean = 0
	// differences between the local config and the cluster
	ExitStatusDiff = 16
)

func applyCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "apply <path>",
		Short: "apply the configuration to the cluster",
		Args:  workflowArgs,
	}

	var opts tanka.ApplyOpts
	cmd.Flags().BoolVar(&opts.Force, "force", false, "force applying (kubectl apply --force)")
	cmd.Flags().BoolVar(&opts.Validate, "validate", true, "validation of resources (kubectl --validate=false)")
	cmd.Flags().BoolVar(&opts.AutoApprove, "dangerous-auto-approve", false, "skip interactive approval. Only for automation!")
	cmd.Flags().StringVar(&opts.DiffToFile, "diff-to-file", "", "Outputs diff as usual but also writes it to specified file")

	vars := workflowFlags(cmd.Flags())
	getJsonnetOpts := jsonnetFlags(cmd.Flags())

	cmd.Run = func(cmd *cli.Command, args []string) error {
		filters, err := process.StrExps(vars.targets...)
		if err != nil {
			return err
		}
		opts.Filters = filters
		opts.JsonnetOpts = getJsonnetOpts()
		opts.Name = vars.name

		return tanka.Apply(args[0], opts)
	}
	return cmd
}

func pruneCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "prune <path>",
		Short: "delete resources removed from Jsonnet",
		Args:  workflowArgs,
	}

	var opts tanka.PruneOpts
	cmd.Flags().BoolVar(&opts.Force, "force", false, "force deleting (kubectl delete --force)")
	cmd.Flags().BoolVar(&opts.AutoApprove, "dangerous-auto-approve", false, "skip interactive approval. Only for automation!")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Selects an environment from inline environments")
	cmd.Flags().StringVar(&opts.DiffToFile, "diff-to-file", "", "Outputs diff as usual but also writes it to specified file")
	getJsonnetOpts := jsonnetFlags(cmd.Flags())

	cmd.Run = func(cmd *cli.Command, args []string) error {
		opts.JsonnetOpts = getJsonnetOpts()

		return tanka.Prune(args[0], opts)
	}

	return cmd
}

func deleteCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "delete <path>",
		Short: "delete the environment from cluster",
		Args:  workflowArgs,
	}

	var opts tanka.DeleteOpts
	cmd.Flags().BoolVar(&opts.Force, "force", false, "force deleting (kubectl delete --force)")
	cmd.Flags().BoolVar(&opts.Validate, "validate", true, "validation of resources (kubectl --validate=false)")
	cmd.Flags().BoolVar(&opts.AutoApprove, "dangerous-auto-approve", false, "skip interactive approval. Only for automation!")
	cmd.Flags().StringVar(&opts.DiffToFile, "diff-to-file", "", "Outputs diff as usual but also writes it to specified file")

	vars := workflowFlags(cmd.Flags())
	getJsonnetOpts := jsonnetFlags(cmd.Flags())

	cmd.Run = func(cmd *cli.Command, args []string) error {
		filters, err := process.StrExps(vars.targets...)
		if err != nil {
			return err
		}
		opts.Filters = filters
		opts.JsonnetOpts = getJsonnetOpts()
		opts.Name = vars.name

		return tanka.Delete(args[0], opts)
	}
	return cmd
}

func diffCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "diff <path>",
		Short: "differences between the configuration and the cluster",
		Args:  workflowArgs,
		Predictors: complete.Flags{
			"diff-strategy": cli.PredictSet("native", "subset", "validate"),
		},
	}

	var opts tanka.DiffOpts
	cmd.Flags().StringVar(&opts.Strategy, "diff-strategy", "", "force the diff-strategy to use. Automatically chosen if not set.")
	cmd.Flags().BoolVarP(&opts.Summarize, "summarize", "s", false, "print summary of the differences, not the actual contents")
	cmd.Flags().BoolVarP(&opts.WithPrune, "with-prune", "p", false, "include objects deleted from the configuration in the differences")
	cmd.Flags().BoolVarP(&opts.ExitZero, "exit-zero", "z", false, "Exit with 0 even when differences are found.")
	cmd.Flags().StringVar(&opts.DiffToFile, "diff-to-file", "", "Outputs diff as usual but also writes it to specified file")

	vars := workflowFlags(cmd.Flags())
	getJsonnetOpts := jsonnetFlags(cmd.Flags())

	cmd.Run = func(cmd *cli.Command, args []string) error {
		filters, err := process.StrExps(vars.targets...)
		if err != nil {
			return err
		}
		opts.Filters = filters
		opts.JsonnetOpts = getJsonnetOpts()
		opts.Name = vars.name

		changes, err := tanka.Diff(args[0], opts)
		if err != nil {
			return err
		}

		if changes == nil {
			log.Println("No differences.")
			os.Exit(ExitStatusClean)
		}

		r := term.Colordiff(*changes)
		if err := fPageln(r); err != nil {
			return err
		}

		exitStatusDiff := ExitStatusDiff
		if opts.ExitZero {
			exitStatusDiff = ExitStatusClean
		}
		os.Exit(exitStatusDiff)
		return nil
	}

	return cmd
}

func showCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "show <path>",
		Short: "jsonnet as yaml",
		Args:  workflowArgs,
	}

	allowRedirect := cmd.Flags().Bool("dangerous-allow-redirect", false, "allow redirecting output to a file or a pipe.")

	vars := workflowFlags(cmd.Flags())
	getJsonnetOpts := jsonnetFlags(cmd.Flags())

	cmd.Run = func(cmd *cli.Command, args []string) error {
		if !interactive && !*allowRedirect {
			fmt.Fprintln(os.Stderr, `Redirection of the output of tk show is discouraged and disabled by default.
If you want to export .yaml files for use with other tools, try 'tk export'.
Otherwise run tk show --dangerous-allow-redirect to bypass this check.`)
			return nil
		}

		filters, err := process.StrExps(vars.targets...)
		if err != nil {
			return err
		}

		pretty, err := tanka.Show(args[0], tanka.Opts{
			JsonnetOpts: getJsonnetOpts(),
			Filters:     filters,
			Name:        vars.name,
		})

		if err != nil {
			return err
		}

		return pageln(pretty.String())
	}
	return cmd
}
