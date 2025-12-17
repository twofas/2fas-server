package command

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/avast/retry-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/push"
)

var tokenPushNotificationTtl = time.Minute * 3

type PushNotificationStatus string

const (
	PushNotificationStatusOK           = "ok"
	PushNotificationStatusNoFCM        = "no_fcm"
	PushNotificationStatusError        = "error"
	PushNotificationStatusUnregistered = "unregistered"
)

type Request2FaTokenPushNotification struct {
	ExtensionId  string `json:"extension_id"`
	IssuerDomain string `json:"issuer_domain"`
	RequestId    string `json:"request_id"`
}

type Request2FaToken struct {
	Id          string `validate:"required,uuid4"`
	ExtensionId string `uri:"extension_id" validate:"required,uuid4"`
	Domain      string `json:"domain" validate:"required,lte=256"`
}

func New2FaTokenRequestFromGin(c *gin.Context) (*Request2FaToken, bool) {
	id := uuid.New()

	cmd := &Request2FaToken{}
	cmd.Id = id.String()

	if err := c.BindJSON(&cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return nil, false
	}
	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return nil, false
	}

	u, err := url.Parse(cmd.Domain)

	if err != nil {
		cmd.Domain = ""
		return cmd, true
	}

	cmd.Domain = fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	return cmd, true
}

type Request2FaTokenHandler struct {
	BrowserExtensionsRepository          domain.BrowserExtensionRepository
	BrowserExtension2FaRequestRepository domain.BrowserExtension2FaRequestRepository
	PairedDevicesRepository              domain.BrowserExtensionDevicesRepository
	Pusher                               push.Pusher
}

func (h *Request2FaTokenHandler) Handle(ctx context.Context, cmd *Request2FaToken) (map[string]PushNotificationStatus, error) {
	log := logging.FromContext(ctx)
	extId, _ := uuid.Parse(cmd.ExtensionId)

	pairedDevices, err := h.findPairedDevices(extId, cmd)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"extension_id": extId.String(),
		"request_id":   cmd.Id,
		"domain":       cmd.Domain,
		"type":         "browser_extension_request",
	}

	result := map[string]PushNotificationStatus{}

	for _, device := range pairedDevices {
		if device.FcmToken == "" {
			log.WithFields(logging.Fields{
				"extension_id":     extId.String(),
				"device_id":        device.Id.String(),
				"token_request_id": cmd.Id,
				"domain":           cmd.Domain,
				"platform":         device.Platform,
				"type":             "browser_extension_request",
			}).Info("Cannot send push notification, missing FCM token")
			result[device.Id.String()] = PushNotificationStatusNoFCM
			continue
		}

		err := h.sendNotification(ctx, device, data)
		if err == nil {
			result[device.Id.String()] = PushNotificationStatusOK
		} else if messaging.IsUnregistered(err) {
			result[device.Id.String()] = PushNotificationStatusUnregistered
		} else {
			result[device.Id.String()] = PushNotificationStatusError
			log.WithFields(logging.Fields{
				"extension_id":     extId.String(),
				"device_id":        device.Id.String(),
				"token_request_id": cmd.Id,
				"domain":           cmd.Domain,
				"platform":         device.Platform,
				"type":             "browser_extension_request",
				"error":            err.Error(),
			}).Error("Cannot send push notification for \"2fa_request\"")
		}
	}

	return result, nil
}

func (h *Request2FaTokenHandler) findPairedDevices(extId uuid.UUID, cmd *Request2FaToken) ([]*domain.ExtensionDevice, error) {
	browserExtension, err := h.BrowserExtensionsRepository.FindById(extId)
	if err != nil {
		return nil, err
	}

	tokenRequestId, _ := uuid.Parse(cmd.Id)
	browserExtension2FaRequest := domain.NewBrowserExtension2FaRequest(tokenRequestId, browserExtension.Id, cmd.Domain)

	err = h.BrowserExtension2FaRequestRepository.Save(browserExtension2FaRequest)
	if err != nil {
		return nil, err
	}

	pairedDevices := h.PairedDevicesRepository.FindAll(browserExtension.Id)
	return pairedDevices, nil
}

func (h *Request2FaTokenHandler) sendNotification(ctx context.Context, device *domain.ExtensionDevice, data map[string]interface{}) error {
	var notification *messaging.Message

	switch device.Platform {
	case domain.Android:
		notification = createPushNotificationForAndroid(device.FcmToken, data)
	case domain.IOS:
		notification = createPushNotificationForIos(device.FcmToken, data)
	}

	return retry.Do(
		func() error {
			return h.Pusher.Send(ctx, notification)
		},
		retry.Attempts(5),
		retry.LastErrorOnly(true),
	)
}

func createPushNotificationForIos(token string, data map[string]interface{}) *messaging.Message {
	ttl := time.Now().Add(tokenPushNotificationTtl)

	return &messaging.Message{
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-expiration": fmt.Sprintf("%d", ttl.Unix()),
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: "2FA request",
						Body:  fmt.Sprintf("2FA request for %s", data["domain"]),
					},
					Category: "authReq",
					Sound:    "default",
				},
				CustomData: data,
			},
		},
		Token: token,
	}
}

func createPushNotificationForAndroid(token string, data map[string]interface{}) *messaging.Message {
	androidData := make(map[string]string, len(data))

	for key, value := range data {
		str, ok := value.(string)
		if !ok {
			continue
		}
		androidData[key] = str
	}

	androidData["click_action"] = "auth_request"

	return &messaging.Message{
		Android: &messaging.AndroidConfig{
			Data: androidData,
			TTL:  &tokenPushNotificationTtl,
		},
		Token: token,
	}
}
