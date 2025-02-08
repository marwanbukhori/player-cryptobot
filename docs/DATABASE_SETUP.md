# Database Setup Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Initial Setup](#initial-setup)
4. [Database Management](#database-management)
5. [Troubleshooting](#troubleshooting)

## Prerequisites

### SQLite Installation

**Ubuntu/Debian:**

```bash
sudo apt-get update
sudo apt-get install sqlite3
```

**macOS:**

```bash
brew install sqlite3
```

**Windows:**

1. Download from https://www.sqlite.org/download.html
2. Add to system PATH

## Installation

1. Create data directory:

```bash
mkdir -p data
```

2. Update `.env` file:

```env
# Database Configuration
DB_PATH="data/trading_bot.db"
```

3. Update `.gitignore`:

```gitignore
.env
*.db
data/
```

## Initial Setup

### 1. Database Creation

The database will be automatically created when the bot starts, but you can manually initialize it:

```bash
sqlite3 data/trading_bot.db
```

### 2. Verify Schema

```bash
# Connect to database
sqlite3 data/trading_bot.db

# Show all tables
.tables

# Show trades table schema
.schema trades

# Exit
.quit
```

Expected schema:

```sql
CREATE TABLE trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL,
    price REAL NOT NULL,
    quantity REAL NOT NULL,
    timestamp DATETIME NOT NULL
);

CREATE INDEX idx_trades_timestamp ON trades(timestamp);
CREATE INDEX idx_trades_symbol ON trades(symbol);
```

## Database Management

### View Recent Trades

```bash
sqlite3 data/trading_bot.db "SELECT * FROM trades ORDER BY timestamp DESC LIMIT 5;"
```

### Trading Statistics

```bash
sqlite3 data/trading_bot.db "
SELECT
    COUNT(*) as total_trades,
    SUM(CASE WHEN price > 0 THEN 1 ELSE 0 END) as winning_trades,
    SUM(quantity * price) as total_volume
FROM trades;"
```

### Backup Database

```bash
# Create backup
sqlite3 data/trading_bot.db ".backup 'data/trading_bot_backup.db'"

# Restore from backup
sqlite3 data/trading_bot_backup.db ".backup 'data/trading_bot.db'"
```

### Export Data

```bash
# Export to CSV
sqlite3 -header -csv data/trading_bot.db "SELECT * FROM trades;" > trades_export.csv
```

### Database Maintenance

```bash
# Optimize database
sqlite3 data/trading_bot.db "VACUUM;"

# Remove old trades
sqlite3 data/trading_bot.db "DELETE FROM trades WHERE timestamp < datetime('now', '-30 days');"
```

## Troubleshooting

### Common Issues

1. **Permission Denied**

```bash
# Fix permissions
chmod 644 data/trading_bot.db
chmod 755 data/
```

2. **Database Locked**

- Ensure only one process is accessing the database
- Check for hung processes:

```bash
lsof | grep trading_bot.db
```

3. **Database Corruption**

```bash
# Check database integrity
sqlite3 data/trading_bot.db "PRAGMA integrity_check;"

# If corrupted, restore from backup
cp data/trading_bot_backup.db data/trading_bot.db
```

### Maintenance Schedule

1. **Daily Tasks**

- Backup database
- Check error logs
- Monitor database size

2. **Weekly Tasks**

- Run VACUUM
- Export data
- Check integrity

3. **Monthly Tasks**

- Archive old data
- Review performance
- Update indexes if needed

### Best Practices

1. **Backups**

- Keep multiple backup copies
- Store in different locations
- Test restoration regularly

2. **Monitoring**

- Watch database size growth
- Monitor query performance
- Check error logs regularly

3. **Security**

- Restrict file permissions
- Keep SQLite updated
- Backup sensitive data

## Additional Commands

### SQLite Console Commands

```bash
.tables          # List all tables
.schema         # Show complete schema
.indexes        # Show all indexes
.backup        # Create backup
.restore       # Restore from backup
.quit          # Exit SQLite console
```

### Useful Queries

```sql
-- Check database size
SELECT page_count * page_size as size_bytes
FROM pragma_page_count(), pragma_page_size();

-- Get table info
PRAGMA table_info(trades);

-- Get index info
PRAGMA index_list(trades);
```
