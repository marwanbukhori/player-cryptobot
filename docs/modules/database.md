# Database Module Documentation

## Overview

The database module provides a unified interface for data persistence using SQLite. It handles all database operations including trade history, position tracking, and performance metrics.

## Components

### Database

Main struct that handles database operations:

```go
type Database struct {
    db   *sql.DB
    gorm *gorm.DB
}
```

### Key Functions

#### Initialize

```go
func Initialize(dsn string) (*Database, error)
```

Creates a new database connection and initializes the schema.

#### SaveTrade

```go
func (db *Database) SaveTrade(trade *models.Trade) error
```

Saves a trade record to the database.

#### GetTradingSummary

```go
func (db *Database) GetTradingSummary() ([]models.TradingSummary, error)
```

Retrieves trading performance metrics grouped by symbol.

#### GetLastBuyTrade

```go
func (db *Database) GetLastBuyTrade(symbol string) (*models.Trade, error)
```

Gets the most recent buy trade for a given symbol.

#### GetOpenPositions

```go
func (db *Database) GetOpenPositions() ([]models.Trade, error)
```

Retrieves all currently open trading positions.

## Schema

The database uses the following tables:

### trades

- id (INTEGER PRIMARY KEY)
- symbol (TEXT)
- side (TEXT)
- price (REAL)
- quantity (REAL)
- value (REAL)
- fee (REAL)
- timestamp (DATETIME)
- pnl (REAL)
- pnl_percent (REAL)
- position_id (TEXT)
- status (TEXT)

## Error Handling

- Connection errors
- Query execution errors
- Schema migration errors
- Data validation errors

## Usage Example

```go
db, err := database.Initialize("trading_bot.db")
if err != nil {
    log.Fatal(err)
}

trade := &models.Trade{
    Symbol:    "BTCUSDT",
    Side:      "BUY",
    Price:     50000.0,
    Quantity:  0.1,
    Timestamp: time.Now(),
}

if err := db.SaveTrade(trade); err != nil {
    log.Error("Failed to save trade:", err)
}
```

## Best Practices

1. **Regular Backups**

   - Backup database file daily
   - Store backups in separate location
   - Test backup restoration periodically

2. **Performance Optimization**

   - Run Cleanup() daily
   - Run Vacuum() weekly
   - Monitor database size growth

3. **Error Handling**

   - Always check error returns
   - Log database errors appropriately
   - Implement retry logic for transient errors

4. **Data Integrity**
   - Use transactions for multiple operations
   - Validate data before insertion
   - Regular data consistency checks

## Limitations

1. **SQLite Constraints**

   - Single writer at a time
   - Not suitable for high-frequency trading
   - Limited concurrent connections

2. **Performance**
   - Queries may slow with large datasets
   - Regular maintenance required
   - Limited to single file storage

## Future Improvements

1. **Planned Enhancements**

   - Trade performance analytics
   - Strategy performance tracking
   - Real-time statistics calculation
   - Data export functionality

2. **Potential Upgrades**
   - Migration to PostgreSQL for higher volume
   - Partitioned tables for better performance
   - Advanced analytics functions
