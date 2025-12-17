package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func TestMobileDeviceTestSuite(t *testing.T) {
	suite.Run(t, new(MobileDeviceTestSuite))
}

type MobileDeviceTestSuite struct {
	suite.Suite
}

func (s *MobileDeviceTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileDevices(s.T())
}

func (s *MobileDeviceTestSuite) TestCreateMobileDevice() {
	type testCase struct {
		deviceName       string
		fcmToken         string
		expectedHttpCode int
	}
	defaultFCMToken := "some-fake-token"
	testsCases := []testCase{
		{deviceName: "", fcmToken: defaultFCMToken, expectedHttpCode: 400},
		{deviceName: " ", fcmToken: defaultFCMToken, expectedHttpCode: 400},
		{deviceName: "   ", fcmToken: defaultFCMToken, expectedHttpCode: 400},
		{deviceName: "john`s android", fcmToken: defaultFCMToken, expectedHttpCode: 200},
		{deviceName: "john ", fcmToken: defaultFCMToken, expectedHttpCode: 200},
		{deviceName: " john doe", fcmToken: defaultFCMToken, expectedHttpCode: 200},
		// empty FCM token should be also valid.
		{deviceName: " john doe", fcmToken: "", expectedHttpCode: 200},
	}

	for _, tc := range testsCases {
		response := createDevice(s.T(), tc.deviceName, tc.fcmToken)

		s.Equal(tc.expectedHttpCode, response.StatusCode)
	}
}

func createDevice(t *testing.T, name, fcmToken string) *http.Response {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"name":"%s","platform":"android","fcm_token":"%s"}`, name, fcmToken))
	return e2e_tests.DoAPIRequest(t, "mobile/devices", http.MethodPost, payload, nil)
}
