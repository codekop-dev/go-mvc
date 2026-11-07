package database

import (
	"strings"
	"testing"

	"github.com/codekop/gomvc/internal/app/config"
)

// TestDSN verifies the two supported connection styles without opening a database.
func TestDSN(t *testing.T) {
	mysql := config.DatabaseConfig{Driver: "mysql", User: "root", Password: "secret", Socket: "/tmp/mysql.sock", Name: "demo"}
	mysqlDSN, err := dsn(mysql)
	if err != nil || !strings.Contains(mysqlDSN, "@unix(/tmp/mysql.sock)/demo") {
		t.Fatalf("mysql socket DSN = %q, err = %v", mysqlDSN, err)
	}

	postgres := config.DatabaseConfig{Driver: "postgres", Host: "127.0.0.1", Port: 5432, User: "postgres", Name: "demo", SSLMode: "disable"}
	postgresDSN, err := dsn(postgres)
	if err != nil || !strings.Contains(postgresDSN, "host=127.0.0.1 port=5432") {
		t.Fatalf("postgres DSN = %q, err = %v", postgresDSN, err)
	}
}
