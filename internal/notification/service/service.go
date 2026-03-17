package notificationservice

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TelegramBot struct {
	Id    string
	Token string
}

func NewTelegramBot(id, token string) *TelegramBot {
	return &TelegramBot{
		Id:    id,
		Token: token,
	}
}

func (tb *TelegramBot) SendMessage(ctx context.Context, order_id, user_id string) error {

	text := fmt.Sprintf(
		"<b>User created order </b>\n\n"+
			"<b>OrderId</b> <code>%s</code>\n"+
			"<b>UserId</b> <code>%s</code>\n", order_id, user_id)

	URL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tb.Token)
	params := url.Values{}
	params.Set("chat_id", tb.Id)
	params.Set("text", text)
	params.Set("parse_mode", "HTML")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, strings.NewReader(params.Encode()))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification error")
	}
	return nil
}
