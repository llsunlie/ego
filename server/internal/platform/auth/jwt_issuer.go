package auth

import "time"

type JWTIssuer struct {
	Secret     []byte
	Exp        time.Duration
	RefreshExp time.Duration
}

func (i JWTIssuer) Issue(userID string) (string, error) {
	return GenerateJWT(userID, i.Secret, i.Exp, "access")
}

func (i JWTIssuer) IssueRefresh(userID string) (string, error) {
	return GenerateJWT(userID, i.Secret, i.RefreshExp, "refresh")
}
