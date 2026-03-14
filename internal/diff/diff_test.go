package diff

import "testing"

var configMap = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
`)

var configMapModified = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: new-value
`)

var configMapTwo = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second-config
  namespace: default
data:
  other: thing
`)

func TestCompare_BothEmpty(t *testing.T) {
	report, err := Compare(nil, nil)
	if err != nil {
		t.Fatalf("Compare(nil, nil) error = %v", err)
	}
	if len(report.Diffs) != 0 {
		t.Errorf("Compare(nil, nil) diffs = %d, want 0", len(report.Diffs))
	}
}

func TestCompare_Identical(t *testing.T) {
	report, err := Compare(configMap, configMap)
	if err != nil {
		t.Fatalf("Compare(identical) error = %v", err)
	}
	if len(report.Diffs) != 0 {
		t.Errorf("Compare(identical) diffs = %d, want 0", len(report.Diffs))
	}
}

func TestCompare_Modified_HasDiffs(t *testing.T) {
	report, err := Compare(configMap, configMapModified)
	if err != nil {
		t.Fatalf("Compare(modified) error = %v", err)
	}
	if len(report.Diffs) == 0 {
		t.Error("Compare(modified) diffs = 0, want > 0")
	}
}

func TestCompare_BaselineEmpty_AllAdded(t *testing.T) {
	report, err := Compare(nil, configMap)
	if err != nil {
		t.Fatalf("Compare(nil, content) error = %v", err)
	}
	if len(report.Diffs) == 0 {
		t.Error("Compare(nil, content) diffs = 0, want > 0")
	}
}

func TestCompare_WorkingEmpty_AllRemoved(t *testing.T) {
	report, err := Compare(configMap, nil)
	if err != nil {
		t.Fatalf("Compare(content, nil) error = %v", err)
	}
	if len(report.Diffs) == 0 {
		t.Error("Compare(content, nil) diffs = 0, want > 0")
	}
}

func TestCompare_MultiDocument_Identical(t *testing.T) {
	report, err := Compare(configMapTwo, configMapTwo)
	if err != nil {
		t.Fatalf("Compare(multi-doc identical) error = %v", err)
	}
	if len(report.Diffs) != 0 {
		t.Errorf("Compare(multi-doc identical) diffs = %d, want 0", len(report.Diffs))
	}
}

func TestCompare_InvalidYAML_ReturnsError(t *testing.T) {
	_, err := Compare([]byte("invalid: yaml: :\n  bad"), configMap)
	if err == nil {
		t.Error("Compare(invalid YAML) error = nil, want non-nil")
	}
}

func TestInputFileFrom_EmptyData(t *testing.T) {
	f, err := inputFileFrom(nil, "test")
	if err != nil {
		t.Fatalf("inputFileFrom(nil) error = %v", err)
	}
	if f.Location != "test" {
		t.Errorf("inputFileFrom(nil).Location = %q, want %q", f.Location, "test")
	}
	if len(f.Documents) != 0 {
		t.Errorf("inputFileFrom(nil).Documents = %d, want 0", len(f.Documents))
	}
}

func TestInputFileFrom_ValidYAML(t *testing.T) {
	f, err := inputFileFrom(configMap, "baseline")
	if err != nil {
		t.Fatalf("inputFileFrom(yaml) error = %v", err)
	}
	if f.Location != "baseline" {
		t.Errorf("inputFileFrom(yaml).Location = %q, want %q", f.Location, "baseline")
	}
	if len(f.Documents) == 0 {
		t.Error("inputFileFrom(yaml).Documents = 0, want > 0")
	}
}
