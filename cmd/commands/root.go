package commands

import (
    "context"
    "encoding/json"

    "architecture-bricks/app"
    "architecture-bricks/contract"

    "github.com/spf13/cobra"
)

func Execute() error {
    ctx := context.Background()

    application, err := app.NewApplication(ctx)
    if err != nil {
        return err
    }

    rootCmd, err := NewRootCommand(application)
    if err != nil {
        return err
    }

    return rootCmd.ExecuteContext(ctx)
}

func NewRootCommand(application contract.Application) (*cobra.Command, error) {
    rootCmd := &cobra.Command{
        Use:          "example",
        Short:        "Product CLI example",
        SilenceUsage: true,
    }

    productCmd, err := newProductCommand(application)
    if err != nil {
        return nil, err
    }

    rootCmd.AddCommand(productCmd)

    return rootCmd, nil
}

func writeJSON(cmd *cobra.Command, value any) error {
    encoder := json.NewEncoder(cmd.OutOrStdout())
    encoder.SetIndent("", "  ")

    return encoder.Encode(value)
}

func requireFlags(cmd *cobra.Command, names ...string) error {
    for _, name := range names {
        if err := cmd.MarkFlagRequired(name); err != nil {
            return err
        }
    }

    return nil
}
