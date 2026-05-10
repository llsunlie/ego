package bootstrap

import (
	"ego-server/internal/identity"

	pb "ego-server/proto/ego"
)

func NewIdentityHandler(p *Platform) pb.EgoServer {
	return identity.NewHandler(identity.Deps{
		DB:     p.Pool,
		Hasher: p.Hasher,
		Tokens: p.Tokens,
	})
}
