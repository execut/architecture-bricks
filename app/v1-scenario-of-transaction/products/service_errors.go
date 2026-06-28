package products

import (
    "errors"

    "architecture-bricks/contract"

    "github.com/jackc/pgx/v5/pgconn"
)

func mapCreateProductError(err error) error {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
        return contract.ErrProductAlreadyExists
    }

    return err
}
