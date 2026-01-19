package steps

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/michaeldyrynda/arbor/internal/config"
	"github.com/michaeldyrynda/arbor/internal/scaffold/types"
	"github.com/michaeldyrynda/arbor/internal/utils"
)

type DatabaseStep struct {
	name     string
	args     []string
	priority int
}

func NewDatabaseStep(cfg config.StepConfig, priority int) *DatabaseStep {
	return &DatabaseStep{
		name:     "database.create",
		args:     cfg.Args,
		priority: priority,
	}
}

func (s *DatabaseStep) Name() string {
	return s.name
}

func (s *DatabaseStep) Priority() int {
	return s.priority
}

func (s *DatabaseStep) Condition(ctx types.ScaffoldContext) bool {
	return true
}

func (s *DatabaseStep) Run(ctx types.ScaffoldContext, opts types.StepOptions) error {
	dbType := ""
	dbName := ""

	for i, arg := range s.args {
		if arg == "--type" && i+1 < len(s.args) {
			dbType = s.args[i+1]
		}
		if arg == "--database" && i+1 < len(s.args) {
			dbName = s.args[i+1]
		}
	}

	if dbType == "" && dbName == "" {
		env := utils.ReadEnvFile(ctx.WorktreePath, ".env")
		dbType = env["DB_CONNECTION"]
		dbName = env["DB_DATABASE"]
	}

	if dbType == "" {
		if opts.Verbose {
			fmt.Printf("  No database type specified, skipping.\n")
		}
		return nil
	}

	if opts.Verbose {
		fmt.Printf("  Creating database (%s)...\n", dbType)
	}

	if dbType == "sqlite" {
		return s.createSqlite(ctx, dbName, opts)
	}

	return s.createMysqlOrPgsql(ctx, dbType, dbName, opts)
}

func (s *DatabaseStep) createSqlite(ctx types.ScaffoldContext, dbName string, opts types.StepOptions) error {
	if dbName == "" {
		dbName = "database/database.sqlite"
	}

	dbFile := filepath.Join(ctx.WorktreePath, dbName)
	dbDir := filepath.Dir(dbFile)

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("creating database directory: %w", err)
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		if opts.Verbose {
			fmt.Printf("  Creating SQLite database: %s\n", dbName)
		}

		file, err := os.Create(dbFile)
		if err != nil {
			return fmt.Errorf("creating SQLite database file: %w", err)
		}
		file.Close()
	} else {
		if opts.Verbose {
			fmt.Printf("  SQLite database already exists: %s\n", dbName)
		}
	}

	return nil
}

func (s *DatabaseStep) createMysqlOrPgsql(ctx types.ScaffoldContext, dbType, dbName string, opts types.StepOptions) error {
	if dbName == "" {
		dbName = generateDatabaseName()
	}

	dbUser := "root"
	dbPass := ""
	dbHost := "127.0.0.1"
	dbPort := ""

	for i, arg := range s.args {
		if arg == "--username" && i+1 < len(s.args) {
			dbUser = s.args[i+1]
		}
		if arg == "--password" && i+1 < len(s.args) {
			dbPass = s.args[i+1]
		}
		if arg == "--host" && i+1 < len(s.args) {
			dbHost = s.args[i+1]
		}
		if arg == "--port" && i+1 < len(s.args) {
			dbPort = s.args[i+1]
		}
	}

	if dbPort == "" && dbType == "mysql" {
		dbPort = "3306"
	} else if dbPort == "" && dbType == "pgsql" {
		dbPort = "5432"
	}

	if opts.Verbose {
		fmt.Printf("  Generated database name: %s\n", dbName)
	}

	var createCmd *exec.Cmd
	if dbType == "mysql" {
		if _, err := exec.LookPath("mysql"); err == nil {
			createCmd = exec.Command("mysql", "-u", dbUser, "-h", dbHost, "-P", dbPort)
			if dbPass != "" {
				createCmd.Args = append(createCmd.Args, fmt.Sprintf("-p%s", dbPass))
			}
			createCmd.Args = append(createCmd.Args, "-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
		}
	} else if dbType == "pgsql" {
		if _, err := exec.LookPath("psql"); err == nil {
			env := os.Environ()
			if dbPass != "" {
				env = append(env, fmt.Sprintf("PGPASSWORD=%s", dbPass))
			}
			createCmd = exec.Command("psql", "-U", dbUser, "-h", dbHost, "-p", dbPort, "-c", fmt.Sprintf("CREATE DATABASE \"%s\"", dbName))
			createCmd.Env = env
		}
	}

	if createCmd != nil {
		if opts.Verbose {
			fmt.Printf("  Creating database with: %s\n", createCmd.Path)
		}
		output, err := createCmd.CombinedOutput()
		if err != nil {
			if opts.Verbose {
				fmt.Printf("  Database creation output: %s\n", string(output))
			}
			fmt.Printf("  Could not create database automatically: %v\n", err)
			fmt.Printf("  Please create database '%s' manually before running migrations.\n", dbName)
		} else {
			if opts.Verbose {
				fmt.Printf("  Database '%s' created successfully.\n", dbName)
			}
		}
	} else {
		fmt.Printf("  No %s client found.\n", dbType)
		fmt.Printf("  Please create database '%s' manually before running migrations.\n", dbName)
	}

	return nil
}

func generateDatabaseName() string {
	bytes := make([]byte, 4)
	if _, err := cryptorand.Read(bytes); err != nil {
		return fmt.Sprintf("app_%d", rand.Int63())
	}
	return fmt.Sprintf("app_%s", hex.EncodeToString(bytes))
}
