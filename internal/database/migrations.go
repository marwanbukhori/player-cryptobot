package database

import "database/sql"

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS trades (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        symbol TEXT NOT NULL,
        side TEXT NOT NULL,
        price REAL NOT NULL,
        quantity REAL NOT NULL,
        timestamp DATETIME NOT NULL
    )`,
	`CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp)`,
	`CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol)`,
}

func runMigrations(db *sql.DB) error {
	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			return err
		}
	}
	return nil
}
