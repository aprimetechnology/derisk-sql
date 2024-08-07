package dbmate

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	dbm "github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"

	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

var migrationFileRegexp = regexp.MustCompile(`^(\d+).*\.sql$`)

func FilterMigrationsByAppliedStatus(migrations []dbm.Migration, appliedStatus bool) []dbm.Migration {
	var filtered []dbm.Migration
	for _, migration := range migrations {
		if migration.Applied == appliedStatus {
			filtered = append(filtered, migration)
		}
	}
	return filtered
}

func SearchDirectoryForMigrations(directory string) ([]dbm.Migration, error) {
	migrations := []dbm.Migration{}
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf(
			"could not find migrations directory %q: %w",
			directory,
			err,
		)
	}

	for _, file := range files {
		// skip sub-directories in the migrations directory
		if file.IsDir() {
			continue
		}

		matches := migrationFileRegexp.FindStringSubmatch(file.Name())
		if len(matches) < 2 {
			continue
		}
		fileName, version := matches[0], matches[1]
		filePath := path.Join(directory, fileName)
		migrations = append(migrations, dbm.Migration{
			Applied:  false,
			FileName: fileName,
			FilePath: filePath,
			Version:  version,
		})
	}

	// sort migrations by filename
	sort.Slice(
		migrations,
		func(i, j int) bool {
			return migrations[i].FileName < migrations[j].FileName
		},
	)
	return migrations, nil
}

func SetRelativeFilePathOnParsedMigrations(migrations []types.ParsedMigration, migrationsDir string) []types.ParsedMigration {
	var updated []types.ParsedMigration
	for _, migration := range migrations {
		migration.RelativeFilePath = strings.TrimPrefix(
			migration.FilePath,
			migrationsDir+"/",
		)
		updated = append(updated, migration)
	}
	return updated
}

func ParseMigrations(migrations []dbm.Migration) ([]types.ParsedMigration, error) {
	var parsedMigrations []types.ParsedMigration
	for _, migration := range migrations {
		parsed, err := migration.Parse()
		if err != nil {
			return nil, fmt.Errorf(
				"error parsing migration file %q: %w",
				migration.FileName,
				err,
			)
		}

		parsedMigrations = append(parsedMigrations, types.ParsedMigration{
			Applied:  migration.Applied,
			FileName: migration.FileName,
			FilePath: migration.FilePath,
			Version:  migration.Version,
			Up:       parsed.Up,
			// currently Migration options only support one field: boolean "transaction"
			UpOptions: map[string]string{
				"transaction": strconv.FormatBool(parsed.UpOptions.Transaction()),
			},
			Down: parsed.Down,
			// currently Migration options only support one field: boolean "transaction"
			// NOTE: the options are set INDEPENDENTLY for the up migration and down migration
			// ie, you absolutely can have `-- migrate:up transaction:false` and `--migrate down: transaction:true`
			DownOptions: map[string]string{
				"transaction": strconv.FormatBool(parsed.DownOptions.Transaction()),
			},
		})
	}
	return parsedMigrations, nil
}
