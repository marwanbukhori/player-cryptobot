# Notifications Module Documentation

## Overview

The notifications module handles alerts and notifications via Telegram.

## Components

### TelegramNotifier

```go
type TelegramNotifier struct {
    token   string
    chatID  string
    enabled bool
}
```

### Key Functions

#### NotifyTrade

```go
func (t *TelegramNotifier) NotifyTrade(symbol string, side string, price float64, quantity float64) error
```

Sends trade execution notifications.

#### NotifyError

```go
func (t *TelegramNotifier) NotifyError(err error) error
```

Sends error notifications.

## Notification Types

- Trade executions
- Error alerts
- System status updates

## Usage Example

```go
notifier := notifications.NewTelegramNotifier(token, chatID)
notifier.NotifyTrade("BTCUSDT", "BUY", 50000.0, 0.1)
```
