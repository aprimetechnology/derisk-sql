package subprocess

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

func Input() types.ParsedMigrationsSummary {
	// 1. read the passed in migrations and metadata from standard input
	var input types.ParsedMigrationsSummary
	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()

	// 2. unmarshal the JSON from standard input
	if err := decoder.Decode(&input); err != nil {
		panic(fmt.Errorf(
			"Error unmarshalling analyzer input (from process stdin) from JSON: %w",
			err,
		))
	}

	return input
}
