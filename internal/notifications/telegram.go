package notifications

import (
	"fmt"
	"net/http"
	"net/url"
)

type TelegramNotifier struct {
	token   string
	chatID  string
	enabled bool
}

/* Telegram Notifier is a component that sends messages to a telegram chat
*  It sends messages to a telegram chat
*  It is enabled if the token and chatID are not empty
 */

func NewTelegramNotifier(token, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		token:   token,
		chatID:  chatID,
		enabled: token != "" && chatID != "",
	}
}

/*
*  Send a message to the telegram chat
*  It is enabled if the token and chatID are not empty
 */
func (t *TelegramNotifier) SendMessage(message string) error {
	if !t.enabled {
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id": {t.chatID},
		"text":    {message},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status code: %d", resp.StatusCode)
	}

	return nil
}

/*
*  Notify a trade
*  It sends a message to the telegram chat
*  It is enabled if the token and chatID are not empty
 */
func (t *TelegramNotifier) NotifyTrade(symbol, side string, price, quantity float64) error {
	message := fmt.Sprintf("Trade Executed:\n\nSymbol: %s\nSide: %s\nPrice: %.2f\nQuantity: %.8f",
		symbol, side, price, quantity)
	return t.SendMessage(message)
}

/*
*  Notify an error
*  It sends a message to the telegram chat
*  It is enabled if the token and chatID are not empty
 */
func (t *TelegramNotifier) NotifyError(err error) error {
	message := fmt.Sprintf("⚠️ Error\n\n%v", err)
	return t.SendMessage(message)
}
