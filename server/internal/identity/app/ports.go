package app

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Verify(hash, plaintext string) error
}

type TokenIssuer interface {
	Issue(userID string) (string, error)
}

type IDGenerator interface {
	New() string
}
