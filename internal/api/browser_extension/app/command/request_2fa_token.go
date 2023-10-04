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

func New2FaTokenRequestFromGin(c *gin.Context) *Request2FaToken {
	id := uuid.New()

	cmd := &Request2FaToken{}
	cmd.Id = id.String()

	c.BindJSON(&cmd)
	c.BindUri(&cmd)

	u, err := url.Parse(cmd.Domain)

	if err != nil {
		cmd.Domain = ""

		return cmd
	}

	cmd.Domain = fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	return cmd
}

type Request2FaTokenHandler struct {
	BrowserExtensionsRepository          domain.BrowserExtensionRepository
	BrowserExtension2FaRequestRepository domain.BrowserExtension2FaRequestRepository
	PairedDevicesRepository              domain.BrowserExtensionDevicesRepository
	Pusher                               push.Pusher
}

func (h *Request2FaTokenHandler) Handle(cmd *Request2FaToken) error {
	extId, _ := uuid.Parse(cmd.ExtensionId)

	browserExtension, err := h.BrowserExtensionsRepository.FindById(extId)

	if err != nil {
		return err
	}

	tokenRequestId, _ := uuid.Parse(cmd.Id)
	browserExtension2FaRequest := domain.NewBrowserExtension2FaRequest(tokenRequestId, browserExtension.Id, cmd.Domain)

	err = h.BrowserExtension2FaRequestRepository.Save(browserExtension2FaRequest)

	if err != nil {
		return err
	}

	pairedDevices := h.PairedDevicesRepository.FindAll(browserExtension.Id)

	data := map[string]interface{}{
		"extension_id": extId.String(),
		"request_id":   cmd.Id,
		"domain":       cmd.Domain,
		"type":         "browser_extension_request",
	}

	for _, device := range pairedDevices {
		var err error
		var notification *messaging.Message

		switch device.Platform {
		case domain.Android:
			notification = createPushNotificationForAndroid(device.FcmToken, data)
		case domain.IOS:
			notification = createPushNotificationForIos(device.FcmToken, data)
		}

		err = retry.Do(
			func() error {
				return h.Pusher.Send(context.Background(), notification)
			},
			retry.Attempts(5),
			retry.LastErrorOnly(true),
		)

		if err != nil && !messaging.IsUnregistered(err) {
			logging.WithFields(logging.Fields{
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

	return nil
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
		androidData[key] = value.(string)
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
