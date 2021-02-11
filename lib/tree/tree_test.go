package tree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type fileExtension string

const (
	extJSON fileExtension = "json"
	extYAML fileExtension = "yaml"
)

type testName string

const (
	testBranchSimple testName = "branch_simple"
	testBranchHard   testName = "branch_hard"
	testTreeSimple   testName = "tree_simple"
	testTreeHard     testName = "tree_hard"
)

func TestTree_MarshalYAML(t *testing.T) {
	tests := []struct {
		name testName
		ext  fileExtension
	}{
		{name: testBranchSimple, ext: extYAML},
		{name: testBranchHard, ext: extYAML},
		{name: testTreeSimple, ext: extYAML},
		{name: testTreeHard, ext: extYAML},
	}

	for i, tc := range tests {
		fileName := fmt.Sprintf("%s.%s", tc.name, tc.ext)
		testName := fmt.Sprintf("MarshalYAML test from fixture %d: %s", i+1, fileName)

		filePath, err := filepath.Abs(filepath.Join("lib", "tree", "fixtures", fileName))
		if err != nil {
			t.Fatalf("create file path of file %q: %v", fileName, err)
		}
		expRaw, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read file %q: %v", filePath, err)
		}
		t.Run(testName, func(t *testing.T) {
			var m Marshalable
			switch tc.name {
			case testBranchSimple, testBranchHard:
				m = NewBranch("", "", 0)
			case testTreeSimple, testTreeHard:
				m = New()
			default:
				t.Fatalf("unsupportable fixture name %q", tc.name)
			}
			switch tc.ext {
			case extYAML:
				if err := yaml.Unmarshal(expRaw, m); err != nil {
					t.Fatalf("prepare tree %q: %v", filePath, err)
				}
			default:
				t.Fatalf("wrong testfile extension: %s", tc.ext)
			}

			var buf bytes.Buffer
			yamlEncoder := yaml.NewEncoder(&buf)
			yamlEncoder.SetIndent(2)
			if err := yamlEncoder.Encode(&m); err != nil {
				t.Fatalf("marshaling error: %v", err)
			}

			if bytes.Compare(buf.Bytes(), expRaw) != 0 {
				t.Errorf("result %q != expectation %q", buf.String(), string(expRaw))
			}
		})
	}
}

func TestTree_MarshalJSON(t *testing.T) {
	tests := []struct {
		name testName
		ext  fileExtension
	}{
		{name: testBranchSimple, ext: extJSON},
		{name: testBranchHard, ext: extJSON},
		{name: testTreeSimple, ext: extJSON},
		{name: testTreeHard, ext: extJSON},
	}

	for i, tc := range tests {
		fileName := fmt.Sprintf("%s.%s", tc.name, tc.ext)
		testName := fmt.Sprintf("MarshalJSON test from fixture %d: %s", i+1, fileName)

		filePath, err := filepath.Abs(filepath.Join("lib", "tree", "fixtures", fileName))
		if err != nil {
			t.Fatalf("create file path of file %q: %v", fileName, err)
		}
		expRaw, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read file %q: %v", filePath, err)
		}
		t.Run(testName, func(t *testing.T) {
			var m Marshalable
			switch tc.name {
			case testBranchSimple, testBranchHard:
				m = NewBranch("", "", 0)
			case testTreeSimple, testTreeHard:
				m = New()
			default:
				t.Fatalf("unsupportable fixture name %q", tc.name)
			}
			switch tc.ext {
			case extJSON:
				if err := json.Unmarshal(expRaw, &m); err != nil {
					t.Fatalf("prepare tree %q: %v", filePath, err)
				}
			default:
				t.Fatalf("wrong testfile extension: %s", tc.ext)
			}

			resBuf := bytes.Buffer{}
			jsonEncoder := json.NewEncoder(&resBuf)
			jsonEncoder.SetIndent("", "  ")
			if err := jsonEncoder.Encode(&m); err != nil {
				t.Fatalf("encoding error: %v", err)
			}
			if err != nil {
				t.Fatalf("marshaling error: %v", err)
			}
			if bytes.Compare(resBuf.Bytes(), expRaw) != 0 {
				t.Errorf("result %q != expectation %q", resBuf.String(), string(expRaw))
			}
		})
	}
}

func TestTree_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name testName
		ext  fileExtension
	}{
		{name: testBranchSimple, ext: extYAML},
		{name: testBranchHard, ext: extYAML},
		{name: testTreeSimple, ext: extYAML},
		{name: testTreeHard, ext: extYAML},
	}

	for i, tc := range tests {
		fileName := fmt.Sprintf("%s.%s", tc.name, tc.ext)
		testName := fmt.Sprintf("UnmarshalYAML test from fixture %d: %s", i+1, fileName)

		filePath, err := filepath.Abs(filepath.Join("lib", "tree", "fixtures", fileName))
		if err != nil {
			t.Fatalf("create file path of file %q: %v", fileName, err)
		}
		expRaw, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read file %q: %v", filePath, err)
		}
		t.Run(testName, func(t *testing.T) {
			var m Marshalable
			switch tc.name {
			case testBranchSimple, testBranchHard:
				m = NewBranch("", "", 0)
			case testTreeSimple, testTreeHard:
				m = New()
			default:
				t.Fatalf("unsupportable fixture name %q", tc.name)
			}

			err := yaml.Unmarshal(expRaw, m)
			if err != nil {
				t.Fatalf("unmarshaling error: %v", err)
			}

			var resBuf bytes.Buffer

			switch tc.ext {
			case extYAML:
				yamlEncoder := yaml.NewEncoder(&resBuf)
				yamlEncoder.SetIndent(2)
				if err := yamlEncoder.Encode(&m); err != nil {
					t.Fatalf("encoding error: %v", err)
				}
			default:
				t.Fatalf("wrong testfile extension: %s", tc.ext)
			}

			if bytes.Compare(resBuf.Bytes(), expRaw) != 0 {
				t.Errorf("result %q != expectation %q", resBuf.String(), string(expRaw))
			}
		})
	}
}

func TestTree_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name testName
		ext  fileExtension
	}{
		{name: testBranchSimple, ext: extJSON},
		{name: testBranchHard, ext: extJSON},
		{name: testTreeSimple, ext: extJSON},
		{name: testTreeHard, ext: extJSON},
	}

	for i, tc := range tests {
		fileName := fmt.Sprintf("%s.%s", tc.name, tc.ext)
		testName := fmt.Sprintf("UnmarshalJSON test from fixture %d: %s", i+1, fileName)

		filePath, err := filepath.Abs(filepath.Join("lib", "tree", "fixtures", fileName))
		if err != nil {
			t.Fatalf("create file path of file %q: %v", fileName, err)
		}
		expRaw, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read file %q: %v", filePath, err)
		}
		t.Run(testName, func(t *testing.T) {
			var m Marshalable
			switch tc.name {
			case testBranchSimple, testBranchHard:
				m = NewBranch("", "", 0)
			case testTreeSimple, testTreeHard:
				m = New()
			default:
				t.Fatalf("unsupportable fixture name %q", tc.name)
			}

			err := json.Unmarshal(expRaw, m)
			if err != nil {
				t.Fatalf("unmarshaling error: %v", err)
			}

			var resBuf bytes.Buffer

			switch tc.ext {
			case extJSON:
				jsonEncoder := json.NewEncoder(&resBuf)
				jsonEncoder.SetIndent("", "  ")
				if err := jsonEncoder.Encode(&m); err != nil {
					t.Fatalf("encoding error: %v", err)
				}
			default:
				t.Fatalf("wrong testfile extension: %s", tc.ext)
			}

			if bytes.Compare(resBuf.Bytes(), expRaw) != 0 {
				t.Errorf("result %q != expectation %q", resBuf.String(), string(expRaw))
			}
		})
	}
}
