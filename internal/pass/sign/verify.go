package sign

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidClaims = errors.New("invalid claims")

// CanI establish connection with type tp given claims in token.
func (s Service) CanI(tokenString string, ct ConnectionType) error {
	cl := jwt.MapClaims{}

	tokenString = jwtHeader + tokenString

	token, err := jwt.ParseWithClaims(
		tokenString,
		&cl,
		func(token *jwt.Token) (interface{}, error) {
			return s.publicKey, nil
		},
		jwt.WithValidMethods([]string{"ES256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	claims, err := token.Claims.GetAudience()
	if err != nil {
		return fmt.Errorf("failed to get claims: %w", err)
	}

	for _, aud := range claims {
		if aud == string(ct) {
			return nil
		}
	}

	return fmt.Errorf("%w: claim %q not found in claims", ErrInvalidClaims, ct)
}
