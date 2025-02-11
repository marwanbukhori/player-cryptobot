CREATE TABLE IF NOT EXISTS trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    position_id TEXT,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL,
    price REAL NOT NULL,
    quantity REAL NOT NULL,
    value REAL NOT NULL,
    fee REAL DEFAULT 0,
    timestamp DATETIME NOT NULL,
    pn_l REAL DEFAULT 0,
    pn_l_percent REAL DEFAULT 0,
    status TEXT DEFAULT 'OPEN',
    created_at DATETIME,
    updated_at DATETIME
);

-- For reporting
CREATE VIEW trade_summary AS
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
GROUP BY symbol;
