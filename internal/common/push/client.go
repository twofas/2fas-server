package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type Pusher interface {
	Send(ctx context.Context, message *messaging.Message) error
}

type FcmPushClient struct {
	FcmMessaging *messaging.Client
}

func NewFcmPushClient(config *domain.FcmPushConfig) *FcmPushClient {
	fileContent, err := ioutil.ReadAll(config.FcmApiServiceAccountFile)

	if err != nil {
		logging.Fatal(err)
	}

	opt := option.WithCredentialsJSON(fileContent)
	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		logging.WithField("error", err.Error()).Fatal("Error initializing FCM App")
	}

	client, err := app.Messaging(context.Background())

	return &FcmPushClient{
		FcmMessaging: client,
	}
}

func (p *FcmPushClient) Send(ctx context.Context, message *messaging.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("could not marshal the message: %w", err)
	}

	logging.WithFields(logging.Fields{
		"notification": string(data),
	}).Debug("Sending push notifications")

	contextWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := p.FcmMessaging.Send(contextWithTimeout, message)

	if err != nil {
		return err
	}

	logging.Infof("FCM notification has been sent: %s", response)

	return nil
}

type FakePushClient struct {
}

func NewFakePushClient() *FakePushClient {
	return &FakePushClient{}
}

func (p *FakePushClient) Send(ctx context.Context, message *messaging.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("could not marshal the message: %w", err)
	}

	logging.WithFields(logging.Fields{
		"notification": string(data),
	}).Debug("Sending fake push notifications")

	return nil
}
