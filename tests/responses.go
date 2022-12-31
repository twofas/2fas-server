package tests

type DeviceResponse struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	FcmToken  string `json:"fcm_token"`
	CreatedAt string `json:"expire_at"`
	UpdatedAt string `json:"updated_at"`
}

type BrowserExtensionResponse struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type PairingResultResponse struct {
	ExtensionId        string `json:"extension_id"`
	ExtensionPublicKey string `json:"extension_public_key"`
}

type DevicePairedBrowserExtensionResponse struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	PairedAt       string `json:"paired_at"`
}

type ExtensionPairedDeviceResponse struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	UserDeviceName string `json:"user_device_name"`
	Platform       string `json:"platform"`
	CreatedAt      string `json:"paired_at"`
}

type AuthTokenRequestResponse struct {
	Id          string `json:"token_request_id"`
	ExtensionId string `json:"extension_id"`
	Domain      string `json:"domain"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}
