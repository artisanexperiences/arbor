package prompts

// DatabaseOption represents an available database choice for reuse.
type DatabaseOption struct {
    Label    string
    DbSuffix string
}

// DbPrompter defines the prompt contract for database-related steps.
type DbPrompter interface {
    SelectDatabase(options []DatabaseOption) (string, error)
    ConfirmMigrations() (bool, error)
    ConfirmDatabaseDrop(suffix string, databases []string) (bool, error)
}
