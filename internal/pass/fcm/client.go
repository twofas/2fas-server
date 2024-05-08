package fcm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/twofas/2fas-server/internal/common/logging"
)

// Response is returned from firebsase for push request.
type Response string

type Client interface {
	Send(ctx context.Context, message *messaging.Message) (Response, error)
}

type client struct {
	FcmMessaging *messaging.Client
}

func NewClient(ctx context.Context, credentials string) (*client, error) {
	opt := option.WithCredentialsJSON([]byte(credentials))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create firebase app: %w", err)
	}
	fcmClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create fcm client: %w", err)
	}
	return &client{FcmMessaging: fcmClient}, nil
}

func (c *client) Send(ctx context.Context, message *messaging.Message) (Response, error) {
	contextWithTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	logging.FromContext(ctx).Info()

	response, err := c.FcmMessaging.Send(contextWithTimeout, message)
	if err != nil {
		return "", fmt.Errorf("failed to send push message: %w", err)
	}

	return Response(response), nil
}

type FakePushClient struct {
}

func NewFakePushClient() *FakePushClient {
	return &FakePushClient{}
}

func (p *FakePushClient) Send(ctx context.Context, message *messaging.Message) (Response, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	logging.WithFields(logging.Fields{
		"notification": string(data),
	}).Debug("Sending fake push notifications")

	return "ok", nil
}
