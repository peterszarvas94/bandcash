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

	usersSpec := usersTableQuerySpec()
	if usersSpec.DefaultSort != "email" || usersSpec.DefaultDir != "asc" {
		t.Fatalf("unexpected users model defaults: %+v", usersSpec)
	}
	for _, key := range []string{"email", "role", "status", "createdAt"} {
		if _, ok := usersSpec.AllowedSorts[key]; !ok {
			t.Fatalf("expected %q in users allowlisted sorts", key)
		}
	}
}
