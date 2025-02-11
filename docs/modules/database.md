# Database Module Documentation

## Overview

The database module provides a unified interface for data persistence using SQLite. It handles all database operations including trade history, position tracking, and performance metrics. The database is fully persistent across bot restarts and sessions.

## Key Features

1. **Persistent Storage**

   - Full data persistence across bot restarts
   - One-time schema initialization
   - No data loss between sessions
   - Safe table and index creation

2. **Database Structure**

   - Tables created only if they don't exist
   - Automatic index optimization
   - Performance-optimized views
   - Safe concurrent access

3. **Performance Optimizations**
   - Minimal logging for better performance
   - Optimized indexes for common queries
   - Efficient view for trading summaries
   - No redundant schema updates

## Components

### Database Struct

```go
type Database struct {
    db   *sql.DB
    gorm *gorm.DB
}
```

### Initialization Process

```go
func Initialize(dsn string) (*Database, error)
```

The initialization process is designed to be safe and non-destructive:

1. **First Run (New Database)**

   - Creates data directory if needed
   - Creates trades table
   - Sets up necessary indexes
   - Creates trading summary view

2. **Subsequent Runs**
   - Connects to existing database
   - Preserves all existing data
   - No schema modifications
   - No table recreation

## Schema

### trades Table

| Column       | Type     | Description                     |
| ------------ | -------- | ------------------------------- |
| id           | INTEGER  | Primary key                     |
| position_id  | TEXT     | Unique position identifier      |
| symbol       | TEXT     | Trading pair (e.g., BTCUSDT)    |
| side         | TEXT     | BUY or SELL                     |
| price        | REAL     | Execution price                 |
| quantity     | REAL     | Trade quantity                  |
| value        | REAL     | Total value (price \* quantity) |
| fee          | REAL     | Trading fee                     |
| timestamp    | DATETIME | When the trade occurred         |
| pn_l         | REAL     | Profit/Loss in USDT             |
| pn_l_percent | REAL     | Profit/Loss percentage          |
| status       | TEXT     | OPEN or CLOSED                  |
| created_at   | DATETIME | Record creation time            |
| updated_at   | DATETIME | Last update time                |

### Indexes

Created only for new databases:

```sql
CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp)
CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol)
CREATE INDEX IF NOT EXISTS idx_trades_side ON trades(side)
CREATE INDEX IF NOT EXISTS idx_trades_status ON trades(status)
```

### Views

The `trade_summary` view provides aggregated trading statistics:

```sql
CREATE VIEW IF NOT EXISTS trade_summary AS
SELECT
    symbol,
    COUNT(*) as total_trades,
    SUM(CASE WHEN pn_l > 0 THEN 1 ELSE 0 END) as winning_trades,
    SUM(CASE WHEN pn_l < 0 THEN 1 ELSE 0 END) as losing_trades,
    ROUND(SUM(pn_l), 4) as total_pnl,
    ROUND(AVG(pn_l_percent), 2) as avg_pnl_percent,
    ROUND(SUM(value), 4) as total_volume,
    MIN(timestamp) as first_trade,
    MAX(timestamp) as last_trade
FROM trades
GROUP BY symbol
```

## Best Practices

1. **Data Safety**

   - Never manually modify the database file
   - Use provided functions for all operations
   - Keep regular offsite backups
   - Don't delete database during active trading

2. **Performance**

   - Use appropriate indexes for queries
   - Keep transactions small
   - Monitor database size growth
   - Regular maintenance checks

3. **Error Handling**
   - Always check error returns
   - Log database errors appropriately
   - Implement retry logic for transient errors
   - Handle "no rows" cases gracefully

## Usage Examples

### Saving a Trade

```go
trade := &models.Trade{
    Symbol:    "BTCUSDT",
    Side:      "BUY",
    Price:     50000.0,
    Quantity:  0.1,
    Timestamp: time.Now(),
}
if err := db.SaveTrade(trade); err != nil {
    log.Errorf("Failed to save trade: %v", err)
}
```

### Getting Trading Summary

```go
summary, err := db.GetTradingSummary()
if err != nil {
    log.Errorf("Failed to get summary: %v", err)
    return
}
for _, s := range summary {
    log.Infof("Symbol: %s, Total PnL: %.2f", s.Symbol, s.TotalPnL)
}
```

## Limitations

1. **SQLite Constraints**

   - Single writer at a time
   - Limited concurrent connections
   - Not suitable for high-frequency trading

2. **Performance Considerations**
   - Large datasets may require cleanup
   - Regular maintenance recommended
   - Monitor query performance

## Error Handling

Common error scenarios and handling:

```go
// Example: Handling trade save errors
if err := db.SaveTrade(trade); err != nil {
    if strings.Contains(err.Error(), "UNIQUE constraint") {
        // Handle duplicate trade
        log.Warnf("Duplicate trade detected: %v", err)
    } else {
        // Handle other errors
        log.Errorf("Failed to save trade: %v", err)
    }
}
```

## Future Improvements

1. **Potential Enhancements**

   - Automated database optimization
   - Query performance monitoring
   - Advanced analytics functions
   - Data archival system

2. **Maintenance Features**
   - Database health checks
   - Automated vacuuming
   - Performance analytics
   - Size monitoring
