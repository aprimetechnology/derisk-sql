package types

type ParsedMigration struct {
	Applied          bool              `json:"applied"`
	FileName         string            `json:"fileName"`
	FilePath         string            `json:"filePath"`
	RelativeFilePath string            `json:"relativeFilePath"`
	Version          string            `json:"version"`
	Up               string            `json:"up"`
	UpOptions        map[string]string `json:"upOptions"`
	Down             string            `json:"down"`
	DownOptions      map[string]string `json:"downOptions"`
}
