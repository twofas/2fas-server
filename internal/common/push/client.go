package push

import (
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	"github.com/twofas/2fas-server/internal/common/logging"
	"google.golang.org/api/option"
	"io/ioutil"
	"time"
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
	data, _ := json.Marshal(message)

	logging.WithFields(logging.Fields{
		"notification": string(data),
	}).Debug("Sending push notifications")

	contextWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := p.FcmMessaging.Send(contextWithTimeout, message)

	if err != nil {
		return err
	}

	logging.Info("FCM notification has been sent: %s", response)

	return nil
}

type FakePushClient struct {
}

func NewFakePushClient() *FakePushClient {
	return &FakePushClient{}
}

func (p *FakePushClient) Send(ctx context.Context, message *messaging.Message) error {
	data, _ := json.Marshal(message)

	logging.WithFields(logging.Fields{
		"notification": string(data),
	}).Debug("Sending fake push notifications")

	return nil
}
