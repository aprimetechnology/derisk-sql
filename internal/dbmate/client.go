package dbmate

import (
	"context"
	"net/url"

	dbm "github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
)

type DbMateClientOpts struct {
	MigrationsDir string
	Dsn           string
}

type DbMateClient struct {
	dbmate *dbm.DB
	opts   DbMateClientOpts
}

func NewDbMateClient(ctx context.Context, opts DbMateClientOpts) (*DbMateClient, error) {
	parsedUrl, err := url.Parse(opts.Dsn)
	if err != nil {
		return nil, err
	}

	db := dbm.New(parsedUrl)
	db.MigrationsDir = []string{opts.MigrationsDir}
	return &DbMateClient{
		dbmate: db,
		opts:   opts,
	}, nil
}

func (m *DbMateClient) SearchDatabaseForMigrations() ([]dbm.Migration, error) {
	return m.dbmate.FindMigrations()
}
