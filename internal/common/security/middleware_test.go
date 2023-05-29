package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/twofas/2fas-server/config"
	httpsec "github.com/twofas/2fas-server/internal/common/http"
)

func Test_DoNotAllowUntrustedIp(t *testing.T) {
	c := config.SecurityConfig{TrustedIP: []string{
		"192.168.0.1/32",
	}}

	whitelistMiddleware := httpsec.IPWhitelistMiddleware(c)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	ctx.Request, _ = http.NewRequest("POST", "/", nil)
	ctx.Request.Header.Set("X-Forwarded-For", "192.168.0.2")

	whitelistMiddleware(ctx)
	assert.Equal(t, recorder.Code, http.StatusForbidden)
}
