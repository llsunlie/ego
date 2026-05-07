package auth

import "time"

type JWTIssuer struct {
	Secret []byte
	Exp    time.Duration
}

func (i JWTIssuer) Issue(userID string) (string, error) {
	return GenerateJWT(userID, i.Secret, i.Exp)
}
