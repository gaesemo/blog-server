package token

import (
	"testing"
)

func TestParseToken(t *testing.T) {

	// config.Load()
	// timeNow := func() time.Time {
	// 	return time.Now()
	// }

	// tok := NewWithUserClaims(UserClaims{
	// 	Audience:       []string{},
	// 	Issuer:         "gsm",
	// 	IssuedAt:       timeNow(),
	// 	ExpirationTime: timeNow().Add(time.Hour),
	// 	NotBefore:      timeNow(),
	// 	UserID:         1,
	// })
	// secret := os.Getenv("JWT_SIGNING_SECRET")
	// slog.Info("secret", slog.String("sec", secret))

	// tokstr, _ := tok.SignedString([]byte(secret))
	// claims := NewUserClaims()
	// _, err := ParseWithClaims(tokstr, claims)
	// if err != nil {
	// 	slog.Error("parsing with claims", slog.Any("error", err))
	// }
	// slog.Info("claims", slog.Any("claims", claims))
	// require.NoError(t, err)
}
