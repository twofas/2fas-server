package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestMobileDeviceTestSuite(t *testing.T) {
	suite.Run(t, new(MobileDeviceTestSuite))
}

type MobileDeviceTestSuite struct {
	suite.Suite
}

func (s *MobileDeviceTestSuite) SetupTest() {
	tests.RemoveAllMobileDevices(s.T())
}

func (s *MobileDeviceTestSuite) TestCreateMobileDevice() {
	type testCase struct {
		deviceName       string
		expectedHttpCode int
	}

	testsCases := []testCase{
		{deviceName: "", expectedHttpCode: 400},
		{deviceName: " ", expectedHttpCode: 400},
		{deviceName: "   ", expectedHttpCode: 400},
		{deviceName: "john`s android", expectedHttpCode: 200},
		{deviceName: "john ", expectedHttpCode: 200},
		{deviceName: " john doe", expectedHttpCode: 200},
	}

	for _, tc := range testsCases {
		response := createDevice(tc.deviceName)

		assert.Equal(s.T(), tc.expectedHttpCode, response.StatusCode)
	}
}

func createDevice(name string) *http.Response {
	fcmToken := "some-fake-token"
	payload := []byte(fmt.Sprintf(`{"name":"%s","platform":"android","fcm_token":"%s"}`, name, fcmToken))

	return tests.DoPost("mobile/devices", payload, nil)
}
