package run

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"

	dbm "github.com/amacneil/dbmate/v2/pkg/dbmate"
	"github.com/aprimetechnology/derisk-sql/internal/dbmate"
	"github.com/aprimetechnology/derisk-sql/internal/reportwriter"
	"github.com/aprimetechnology/derisk-sql/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagDSN              = "dsn"
	flagOutput           = "output-dir"
	flagAnalyzers        = "analyzers"
	flagVerbose          = "verbose"
	flagConfig           = "config"
	defaultOutputDir     = "reports"
	flagMigrationsDir    = "migrations-dir"
	defaultMigrationsDir = "migrations"
)

type runCheckFlags struct {
	Dsn           string
	OutputDir     string
	Analyzers     []string
	Config        map[string]string
	Verbose       bool
	MigrationsDir string
}

var (
	// list of analyzers provided by this repo and expected to be used
	// by default if user does not override with their own analyzers list
	// ie, see: github.com/aprimetechnology/derisk-sql/analyzers/* directories
	defaultAnalyzers = []string{
		"analyzer-create-index-concurrently",
		"analyzer-drop-index-concurrently",
		"analyzer-index-concurrently-within-transaction",
		"analyzer-naming-convention",
	}
	flags       = runCheckFlags{}
	RunCheckCmd = &cobra.Command{
		Use:  "run",
		Long: `Runs the SQL static linting checks against migrations directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.ConfigFileUsed() != "" {
				// by default viper.Unmarshal unwisely merges config file contents
				// on top of existing flag default values. eg:
				//  - defaultAnalyzers = ["a1", "a2", "a3"]
				//  - configAnalyzers = ["c1"]
				// flags.Analyzers would end up, confusingly, being ["c1", "a2", "a3"]

				// hence we wipe defaults here by instantiating a blank flags object
				configFlags := runCheckFlags{}
				if err := viper.Unmarshal(&configFlags); err != nil {
					return fmt.Errorf(
						"Failure to unmarshal config file via viper: %w",
						err,
					)
				}

				// copy over all flags from the config file that are set and have a value
				if configFlags.Dsn != "" {
					flags.Dsn = configFlags.Dsn
				}
				if configFlags.OutputDir != "" {
					flags.OutputDir = configFlags.OutputDir
				}
				if len(configFlags.Analyzers) != 0 {
					flags.Analyzers = configFlags.Analyzers
				}
				if len(configFlags.Config) != 0 {
					flags.Config = configFlags.Config

				}
				if configFlags.MigrationsDir != "" {
					flags.MigrationsDir = configFlags.MigrationsDir
				}
			}
			return runCheckRun(cmd, args, flags)
		},
	}
)

func init() {
	RunCheckCmd.Flags().StringVar(
		&flags.Dsn,
		flagDSN,
		"",
		"Database DSN",
	)
	RunCheckCmd.Flags().StringVar(
		&flags.OutputDir,
		flagOutput,
		defaultOutputDir,
		"Directory to output results",
	)
	RunCheckCmd.Flags().StringArrayVar(
		&flags.Analyzers,
		flagAnalyzers,
		defaultAnalyzers,
		"Analyzer executable names (or file paths) to run on migration files",
	)
	RunCheckCmd.Flags().StringToStringVar(
		&flags.Config,
		flagConfig,
		nil,
		"Config specified as key=value pairs in a comma separated list",
	)
	RunCheckCmd.Flags().BoolVar(
		&flags.Verbose,
		flagVerbose,
		false,
		"Include verbose output or not",
	)
	RunCheckCmd.Flags().StringVar(
		&flags.MigrationsDir,
		flagMigrationsDir,
		defaultMigrationsDir,
		"Directory containing migrations",
	)
}

func getParsedMigrations(cmd *cobra.Command, args []string, flags runCheckFlags) ([]types.ParsedMigration, error) {
	migrationsDir, err := filepath.Abs(flags.MigrationsDir)
	if err != nil {
		return nil, err
	}

	// if --dsn is provided, search the database + local filesystem for migrations
	// otherwise only search the local filesystem
	var migrations []dbm.Migration
	if flags.Dsn != "" {
		client, err := dbmate.NewDbMateClient(
			cmd.Context(),
			dbmate.DbMateClientOpts{
				MigrationsDir: migrationsDir,
				Dsn:           flags.Dsn,
			},
		)
		if err != nil {
			return nil, err
		}
		migrations, err = client.SearchDatabaseForMigrations()
		if err != nil {
			return nil, err
		}
		// only include migrations that have NOT yet been applied
		migrations = dbmate.FilterMigrationsByAppliedStatus(migrations, false)
	} else {
		migrations, err = dbmate.SearchDirectoryForMigrations(migrationsDir)
		if err != nil {
			return nil, err
		}
	}

	parsedMigrations, err := dbmate.ParseMigrations(migrations)
	if err != nil {
		return nil, err
	}
	parsedMigrations = dbmate.SetRelativeFilePathOnParsedMigrations(parsedMigrations, migrationsDir)
	return parsedMigrations, nil
}

func runAnalyzer(analyzerPath string, input types.ParsedMigrationsSummary) (*types.AnalyzedMigrationsSummary, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var output, errorOutput bytes.Buffer
	subprocess := exec.Command(analyzerPath)
	subprocess.Stdin = bytes.NewReader(inputBytes)
	subprocess.Stdout = &output
	subprocess.Stderr = &errorOutput

	err = subprocess.Run()
	// if there's any error, include the stdout and stderr contents in the error message
	fullOutputErr := fmt.Errorf(
		"Subprocess stdout:'''%s'''\nSubprocess stderr:'''%s'''",
		output.String(),
		errorOutput.String(),
	)
	if err != nil {
		return nil, errors.Join(fullOutputErr, err)
	}

	var result types.AnalyzedMigrationsSummary
	decoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return nil, errors.Join(fullOutputErr, err)
	}
	return &result, nil
}

func runCheckRun(cmd *cobra.Command, args []string, flags runCheckFlags) error {
	parsedMigrations, err := getParsedMigrations(cmd, args, flags)
	if err != nil {
		return err
	}
	if len(parsedMigrations) == 0 {
		fmt.Printf("No migrations detected in migration directory %q\n", flags.MigrationsDir)
		return nil
	}

	hasFatalErrors := false

	for _, analyzer := range flags.Analyzers {
		summary, err := runAnalyzer(analyzer, types.ParsedMigrationsSummary{
			Metadata: types.MigrationManagerMetadata{
				Name:             "dbmate",
				ConnectionString: flags.Dsn,
				Config:           flags.Config,
			},
			Migrations: parsedMigrations,
		})
		if err != nil {
			// do not return early, continue running other analyzers
			fmt.Printf("Error encountered running analyzer %q: \n%s\n\n", analyzer, err)
			continue
		}

		reportwriter.WriteReportsToStdout(analyzer, summary.Reports, flags.Verbose)

		if flags.OutputDir != "" {
			err = reportwriter.WriteReportsToJsonFile(analyzer, summary.Reports, flags.OutputDir)
			if err != nil {
				// do not return early, continue running other analyzers
				fmt.Printf("Error encountered writing reports to a JSON file for analyzer %q:\n%s\n", analyzer, err)
			}
		}

		if reportwriter.GetReportFatalityStatus(summary.Reports) {
			hasFatalErrors = true
		}
	}

	if hasFatalErrors {
		return errors.New("Encountered FATAL errors!")
	}
	return nil
}
