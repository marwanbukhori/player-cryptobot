CREATE TABLE IF NOT EXISTS trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL,
    price REAL NOT NULL,
    quantity REAL NOT NULL,
    value REAL NOT NULL,
    fee REAL NOT NULL,
    timestamp DATETIME NOT NULL,
    pnl REAL,
    pnl_percent REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- For reporting
CREATE VIEW trade_summary AS
SELECT
    symbol,
    COUNT(*) as total_trades,
    SUM(CASE WHEN pnl > 0 THEN 1 ELSE 0 END) as winning_trades,
    SUM(CASE WHEN pnl < 0 THEN 1 ELSE 0 END) as losing_trades,
    ROUND(SUM(pnl), 4) as total_pnl,
    ROUND(AVG(pnl_percent), 2) as avg_pnl_percent,
    MIN(timestamp) as first_trade,
    MAX(timestamp) as last_trade
FROM trades
GROUP BY symbol;
