package mask

import (
	"strings"
	"testing"
)

func TestMaskSQLLogLine_InsertPasswordEmail(t *testing.T) {
	line := `INSERT INTO "tbl_users" ("name","age","email","password") VALUES ('test01',1,'test01@gmail.com','$2a$10$nlm/LlA3PIJl5H1zQnwe1esoVAWUtSHcxfupybHdulGpF5a1qnheC') RETURNING "id","created_at","updated_at","deleted_at"`
	m := map[string]string{
		"password": "2:2",
		"email":    "4:@gmail.com",
	}
	got := MaskSQLLogLine(line, m)
	if strings.Contains(got, "$2a$10$") {
		t.Fatalf("password hash should be masked: %s", got)
	}
	if strings.Contains(got, "'test01@gmail.com'") {
		t.Fatalf("email local part should be masked: %s", got)
	}
	if !strings.Contains(got, "test**@gmail.com") {
		t.Fatalf("email should show masked local + clear @gmail.com: %s", got)
	}
	if strings.Contains(got, "$2a$10$nlm") {
		t.Fatalf("password middle should be masked: %s", got)
	}
	if !strings.Contains(got, "$2") || !strings.Contains(got, "eC')") {
		t.Fatalf("password should keep prefix/suffix (2:2): %s", got)
	}
}

func TestMaskSQLLogLine_NoMaskWhenEmptyMap(t *testing.T) {
	line := `INSERT INTO "t" ("p") VALUES ('secret')`
	if MaskSQLLogLine(line, nil) != line {
		t.Fatal("expected unchanged")
	}
	if MaskSQLLogLine(line, map[string]string{}) != line {
		t.Fatal("expected unchanged")
	}
}

func TestMaskSQLLogLine_NonInsertUnchanged(t *testing.T) {
	line := `ERROR: relation "tbl_users" does not exist`
	m := map[string]string{"password": "2:2"}
	if MaskSQLLogLine(line, m) != line {
		t.Fatal("expected unchanged")
	}
}

func TestMaskSQLLogLine_ANSIStrippedWhenMasked(t *testing.T) {
	line := "\x1b[35mINSERT INTO \"t\" (\"password\") VALUES ('abc')\x1b[0m"
	m := map[string]string{"password": "1:1"}
	got := MaskSQLLogLine(line, m)
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("ANSI should be stripped when masking: %q", got)
	}
	if !strings.Contains(got, "a*c") {
		t.Fatalf("expected masked value: %s", got)
	}
}
