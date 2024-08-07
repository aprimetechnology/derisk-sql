package types

type MigrationManagerMetadata struct {
	Name             string            `json:"name"`
	ConnectionString string            `json:"connectionString"`
	Config           map[string]string `json:"config,omitempty"`
}

type ParsedMigrationsSummary struct {
	Metadata   MigrationManagerMetadata `json:"metadata"`
	Migrations []ParsedMigration        `json:"migrations"`
}

type AnalyzedMigrationsSummary struct {
	Reports []Report `json:"reports"`
}
