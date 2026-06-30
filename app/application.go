package app

import (
    "context"
    "fmt"
    "os"

    v1 "architecture-bricks/app/v1-scenario-of-transaction/products"
    v2 "architecture-bricks/app/v2-repository/products"
    v3 "architecture-bricks/app/v3-domain-driven-design-light/products"
    v4 "architecture-bricks/app/v4-domain-driven-design-light-with-events/products"
    v5 "architecture-bricks/app/v5-optimistic-locking/products"
    "architecture-bricks/contract"
)

const VariantV1ScenarioOfTransaction = "v1_scenario_of_transaction"
const VariantV2Repository = "v2_repository"
const VariantV3DomainDrivenDesignLight = "v3_domain_driven_design_light"
const VariantV4DomainDrivenDesignLightWithEvents = "v4_domain_driven_design_light_with_events"
const VariantV5OptimisticLocking = "v5_optimistic_locking"

func NewApplication(ctx context.Context) (contract.Application, error) {
    return NewApplicationByVariant(ctx, os.Getenv("APP_VARIANT"))
}

func NewApplicationByVariant(ctx context.Context, variant string) (contract.Application, error) {
    if variant == "" {
        variant = VariantV1ScenarioOfTransaction
    }

    switch variant {
    case VariantV1ScenarioOfTransaction:
        return v1.NewService(ctx)
    case VariantV2Repository:
        repo, err := v2.NewPostgresRepository(ctx)
        if err != nil {
            return nil, err
        }

        return v2.NewService(repo), nil
    case VariantV3DomainDrivenDesignLight:
        return v3.NewService(ctx)
    case VariantV4DomainDrivenDesignLightWithEvents:
        return v4.NewService(ctx)
    case VariantV5OptimisticLocking:
        return v5.NewService(ctx)
    default:
        return nil, fmt.Errorf("unknown APP_VARIANT %q", variant)
    }
}
