package subprocess

import (
	"encoding/json"
	"fmt"

	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

func Output(output types.AnalyzedMigrationsSummary) {
	outputBytes, err := json.Marshal(output)
	if err != nil {
		panic(fmt.Errorf(
			"Error marshalling analyzer output to JSON: %w",
			err,
		))
	}

	// output the marshalled JSON string to the stdout, as expected
	fmt.Println(string(outputBytes))
}
