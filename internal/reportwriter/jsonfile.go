package reportwriter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

const (
	DefaultFilePermissions = 0644
	ReportFileNameFormat   = "report.%s.json"
	ReportFileNameGlob     = "report.*.json"
)

func ListReportJsonFiles(directoryPath string) ([]string, error) {
	filePattern := filepath.Join(directoryPath, ReportFileNameGlob)
	return filepath.Glob(filePattern)
}

func GetReportsFromJsonFile(reportFileName string) (*[]types.Report, error) {
	contents, err := os.ReadFile(reportFileName)
	if err != nil {
		return nil, fmt.Errorf(
			"Failure reading report filename %q: %w",
			reportFileName,
			err,
		)
	}
	reports := []types.Report{}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reports); err != nil {
		return nil, fmt.Errorf(
			"Failure unmarshalling JSON file contents of report filename %q: %w",
			reportFileName,
			err,
		)
	}
	return &reports, nil
}

func GetAllReportsFromJsonFileDirectory(directoryPath string) ([]types.Report, error) {
	reportFiles, err := ListReportJsonFiles(directoryPath)
	if err != nil {
		return []types.Report{}, fmt.Errorf(
			"Failure finding report files in directory %q matching pattern %q: %w",
			directoryPath,
			ReportFileNameGlob,
			err,
		)
	}
	allReports := []types.Report{}
	for _, reportFile := range reportFiles {
		reports, err := GetReportsFromJsonFile(reportFile)
		if err != nil {
			return []types.Report{}, err
		}
		if reports == nil {
			continue
		}
		for _, report := range *reports {
			allReports = append(allReports, report)
		}
	}
	return allReports, nil
}

// Destructively overwrite any existing report json file
func overWriteReportJsonFile(directoryPath string, filename string, contents []byte) error {
	var err error
	outputFile := filepath.Join(directoryPath, filename)
	if _, err = os.Stat(outputFile); err == nil {
		if err = os.Remove(outputFile); err != nil {
			return err
		}
	}
	return os.WriteFile(outputFile, contents, DefaultFilePermissions)
}

func WriteReportsToJsonFile(analyzer string, reports []types.Report, outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf(
			"Can not write reports to output-dir when output-dir is an empty string (%q)!",
			outputDir,
		)
	}
	// ensure that the outputDir exists and create it if it does not exist
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}
	reportJson, err := json.Marshal(reports)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf(ReportFileNameFormat, filepath.Base(analyzer))
	if err = overWriteReportJsonFile(outputDir, filename, reportJson); err != nil {
		return err
	}
	return nil
}
