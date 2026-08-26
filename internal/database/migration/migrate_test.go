package migration

import "testing"

func TestLoadAndSplitStatements(t *testing.T) {
	migrations, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) < 2 || migrations[0].version >= migrations[1].version {
		t.Fatalf("migrations are not ordered: %#v", migrations)
	}
	got := splitStatements(" CREATE TABLE one(id int);\n; CREATE TABLE two(id int); ")
	if len(got) != 2 {
		t.Fatalf("statements = %#v, want 2", got)
	}
}
