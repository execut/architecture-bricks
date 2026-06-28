package commands

import (
    "architecture-bricks/contract"

    "github.com/spf13/cobra"
)

func newProductCommand(application contract.Application) (*cobra.Command, error) {
    productCmd := &cobra.Command{
        Use:   "product",
        Short: "Manage products",
    }

    createCmd, err := newProductCreateCommand(application)
    if err != nil {
        return nil, err
    }

    getCmd, err := newProductGetCommand(application)
    if err != nil {
        return nil, err
    }

    historyCmd, err := newProductHistoryCommand(application)
    if err != nil {
        return nil, err
    }

    updateCmd, err := newProductUpdateCommand(application)
    if err != nil {
        return nil, err
    }

    productCmd.AddCommand(createCmd, getCmd, historyCmd, updateCmd)

    return productCmd, nil
}

func newProductUpdateCommand(application contract.Application) (*cobra.Command, error) {
    var productID string
    var userID string
    var name string

    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update product name",
        RunE: func(cmd *cobra.Command, _ []string) error {
            if err := application.UpdateProduct(cmd.Context(), contract.UpdateProduct{
                ProductID: productID,
                UserID:    userID,
                Name:      name,
            }); err != nil {
                return err
            }

            product, err := application.GetProduct(cmd.Context(), contract.GetProduct{ProductID: productID})
            if err != nil {
                return err
            }

            return writeJSON(cmd, product)
        },
    }

    cmd.Flags().StringVar(&productID, "id", "", "product id")
    cmd.Flags().StringVar(&userID, "user-id", "", "user id")
    cmd.Flags().StringVar(&name, "name", "", "new product name")

    if err := requireFlags(cmd, "id", "user-id", "name"); err != nil {
        return nil, err
    }

    return cmd, nil
}

func newProductCreateCommand(application contract.Application) (*cobra.Command, error) {
    var productID string
    var userID string
    var name string

    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create product",
        RunE: func(cmd *cobra.Command, _ []string) error {
            if err := application.CreateProduct(cmd.Context(), contract.CreateProduct{
                ProductID: productID,
                UserID:    userID,
                Name:      name,
            }); err != nil {
                return err
            }

            product, err := application.GetProduct(cmd.Context(), contract.GetProduct{ProductID: productID})
            if err != nil {
                return err
            }

            return writeJSON(cmd, product)
        },
    }

    cmd.Flags().StringVar(&productID, "id", "", "product id")
    cmd.Flags().StringVar(&userID, "user-id", "", "user id")
    cmd.Flags().StringVar(&name, "name", "", "product name")

    if err := requireFlags(cmd, "id", "user-id", "name"); err != nil {
        return nil, err
    }

    return cmd, nil
}

func newProductGetCommand(application contract.Application) (*cobra.Command, error) {
    var productID string

    cmd := &cobra.Command{
        Use:   "get",
        Short: "Get product",
        RunE: func(cmd *cobra.Command, _ []string) error {
            product, err := application.GetProduct(cmd.Context(), contract.GetProduct{ProductID: productID})
            if err != nil {
                return err
            }

            return writeJSON(cmd, product)
        },
    }

    cmd.Flags().StringVar(&productID, "id", "", "product id")

    if err := requireFlags(cmd, "id"); err != nil {
        return nil, err
    }

    return cmd, nil
}

func newProductHistoryCommand(application contract.Application) (*cobra.Command, error) {
    var productID string

    cmd := &cobra.Command{
        Use:   "history",
        Short: "Get product history",
        RunE: func(cmd *cobra.Command, _ []string) error {
            history, err := application.ProductHistory(cmd.Context(), contract.ProductHistory{ProductID: productID})
            if err != nil {
                return err
            }

            return writeJSON(cmd, history)
        },
    }

    cmd.Flags().StringVar(&productID, "id", "", "product id")

    if err := requireFlags(cmd, "id"); err != nil {
        return nil, err
    }

    return cmd, nil
}
