package cmd

import (
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/spf13/cobra"
	"os"
)

func NewPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Generates a plan based on the specified configuration, grept plan [path to config files]",
		RunE:  planFunc(),
	}
	return cmd
}

func planFunc() func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		var cfgDir string
		if len(args) == 1 {
			cfgDir = "."
		} else {
			cfgDir = args[1]
		}
		configPath, cleaner, err := getConfigFolder(cfgDir, c.Context())
		if cleaner != nil {
			defer cleaner()
		}
		if err != nil {
			return fmt.Errorf("error getting config %s: %+v", cfgDir, err)
		}

		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting os wd: %+v", err)
		}
		config, err := pkg.LoadConfig(pkg.NewGreptConfig(), pwd, configPath, c.Context())
		if err != nil {
			return fmt.Errorf("error parsing config: %+v\n", err)
		}

		plan, err := pkg.RunGreptPlan(config)
		if err != nil {
			return fmt.Errorf("error generating plan: %s\n", err.Error())
		}

		if len(plan.FailedRules) == 0 {
			fmt.Println("All rule checks successful, nothing to do.")
			return nil
		}
		fmt.Println(plan.String())
		return nil
	}
}
