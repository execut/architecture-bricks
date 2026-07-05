package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"architecture-bricks/app"
	"architecture-bricks/contract"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

const (
	TestProductName    = "Coffee"
	TestProductNameCat = "Кот"
	TestProductNameDog = "Собака"
)

var TestApplicationVariants = testApplicationVariants()

func testApplicationVariants() []string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot detect test application variants: caller is unknown")
	}

	appDir := filepath.Join(filepath.Dir(filename), "..", "..", "app")
	entries, err := os.ReadDir(appDir)
	if err != nil {
		panic(err)
	}

	variants := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "v") {
			continue
		}

		variants = append(variants, strings.ReplaceAll(entry.Name(), "-", "_"))
	}

	sort.Strings(variants)

	return variants
}

func TestApplicationVariantList(t testing.TB) []string {
	t.Helper()

	variant := os.Getenv("TEST_APP_VARIANT")
	if variant == "" {
		return TestApplicationVariants
	}

	for _, testVariant := range TestApplicationVariants {
		if testVariant == variant {
			return []string{variant}
		}
	}

	require.Failf(t, "unknown test application variant", "TEST_APP_VARIANT=%s", variant)

	return nil
}

func NewTestApplication(t testing.TB, ctx context.Context, variant string) contract.Application {
	t.Helper()

	application, err := app.NewApplicationByVariant(ctx, variant)
	require.NoError(t, err)

	return application
}

func TestDatabaseURL(t testing.TB) string {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}

	return databaseURL
}

func CleanupProduct(t testing.TB, ctx context.Context, databaseURL string, productID string) {
	t.Helper()

	CleanupProducts(t, ctx, databaseURL, []string{productID})
}

func CleanupProducts(t testing.TB, ctx context.Context, databaseURL string, productIDList []string) {
	t.Helper()

	if len(productIDList) == 0 {
		return
	}

	conn, err := pgx.Connect(ctx, databaseURL)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close(ctx))
	}()

	_, err = conn.Exec(ctx, `DELETE FROM event WHERE entry_id = ANY($1)`, productIDList)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `DELETE FROM product WHERE id = ANY($1)`, productIDList)
	require.NoError(t, err)
}

func WaitDB(t testing.TB, ctx context.Context, databaseURL string) {
	t.Helper()

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var err error
	for {
		var conn *pgx.Conn
		conn, err = pgx.Connect(waitCtx, databaseURL)
		if err == nil {
			err = conn.Ping(waitCtx)
			closeErr := conn.Close(waitCtx)
			if err == nil && closeErr == nil {
				return
			}

			if err == nil {
				err = closeErr
			}
		}

		select {
		case <-waitCtx.Done():
			require.Failf(t, "wait db", "last error: %v", err)
			return
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// SupportsConflict возвращает true для всех вариантов приложения,
// потому что optimistic locking поддерживается каждой реализацией.
func SupportsConflict(variant string) bool {
	return true
}
