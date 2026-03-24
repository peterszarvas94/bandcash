package group

import "testing"

func TestGroupTableSpecs(t *testing.T) {
	model := NewModel()
	spec := model.TableQuerySpec()

	if spec.DefaultSort != "name" || spec.DefaultDir != "asc" {
		t.Fatalf("unexpected group model defaults: %+v", spec)
	}
	if _, ok := spec.AllowedSorts["name"]; !ok {
		t.Fatalf("expected name to be allowlisted in group model spec")
	}

	accessSpec := (&AccessModel{}).TableQuerySpec()
	if accessSpec.DefaultSort != "createdAt" || accessSpec.DefaultDir != "desc" {
		t.Fatalf("unexpected access model defaults: %+v", accessSpec)
	}
	for _, key := range []string{"email", "role", "status", "createdAt"} {
		if _, ok := accessSpec.AllowedSorts[key]; !ok {
			t.Fatalf("expected %q in access allowlisted sorts", key)
		}
	}
}
