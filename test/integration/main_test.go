package integration

import (
	"context"
	"os"
	"testing"

	"github.com/meteoradev/BlogApi/test/fixtures"
)

var testDB *TestDatabase
var testCache *TestCache

func TestMain(m *testing.M) {
	testDB = startTestDatabase()
	testCache = startTestCache()

	code := m.Run()

	if testDB.DB != nil {
		testDB.DB.Close()
	}

	if testDB.Container != nil {
		testDB.Container.Terminate(context.Background())
	}
	if testCache.Container != nil {
		testCache.Container.Terminate(context.Background())
	}

	os.Exit(code)
}

func startTestDatabase() *TestDatabase {
	ctx := context.Background()
	testDB, err := StartPostgresContainer(ctx)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}

	if err := fixtures.RunMigrations(testDB.ConnStr); err != nil {
		panic(err)
	}

	return testDB
}

func startTestCache() *TestCache {
	ctx := context.Background()
	testCache, err := StartRedisContainer(ctx)
	if err != nil {
		panic("failed to start redis container: " + err.Error())
	}
	return testCache
}
