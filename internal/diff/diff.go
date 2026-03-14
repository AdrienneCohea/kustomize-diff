// Package diff compares two kustomize build outputs using dyff.
package diff

import (
	"fmt"

	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
)

// Compare diffs two kustomize build outputs and returns a dyff Report.
// Either argument may be nil or empty, in which case all content in the
// other will appear as added or removed.
func Compare(baselineYAML, workingYAML []byte) (dyff.Report, error) {
	from, err := inputFileFrom(baselineYAML, "baseline")
	if err != nil {
		return dyff.Report{}, fmt.Errorf("parsing baseline YAML: %w", err)
	}
	to, err := inputFileFrom(workingYAML, "working")
	if err != nil {
		return dyff.Report{}, fmt.Errorf("parsing working YAML: %w", err)
	}
	return dyff.CompareInputFiles(from, to)
}

func inputFileFrom(data []byte, label string) (ytbx.InputFile, error) {
	if len(data) == 0 {
		return ytbx.InputFile{Location: label}, nil
	}
	docs, err := ytbx.LoadDocuments(data)
	if err != nil {
		return ytbx.InputFile{}, err
	}
	return ytbx.InputFile{
		Location:  label,
		Documents: docs,
	}, nil
}
