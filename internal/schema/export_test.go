package schema

import (
	"encoding/json"
	"testing"
)

func TestExportSchema_ReturnsAllTypes(t *testing.T) {
	schema := ExportSchema()
	if schema.Version != "1" {
		t.Errorf("version = %q, want 1", schema.Version)
	}
	if len(schema.AssertionTypes) < 28 {
		t.Errorf("got %d assertion types, want at least 28", len(schema.AssertionTypes))
	}
}

func TestExportSchema_AllHaveRequiredFields(t *testing.T) {
	schema := ExportSchema()
	for _, at := range schema.AssertionTypes {
		if at.Name == "" {
			t.Error("assertion type missing name")
		}
		if at.YAML == "" {
			t.Errorf("%s: missing yaml_field", at.Name)
		}
		if len(at.Fields) == 0 {
			t.Errorf("%s: has no fields", at.Name)
		}
		for _, f := range at.Fields {
			if f.Name == "" || f.Type == "" {
				t.Errorf("%s.%s: field missing name or type", at.Name, f.Name)
			}
		}
	}
}

func TestExportSchemaJSON_ValidJSON(t *testing.T) {
	data, err := ExportSchemaJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}
