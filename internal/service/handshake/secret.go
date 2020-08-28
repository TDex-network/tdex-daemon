package handshakeservice

import (
	"context"

	pbhandshake "github.com/tdex-network/tdex-protobuf/generated/go/handshake"
)

//UnarySecret is the domain controller for the UnarySecret RPC
func (s *Service) UnarySecret(ctx context.Context, req *pbhandshake.SecretMessage) (res *pbhandshake.SecretMessage, err error) {
	return &pbhandshake.SecretMessage{}, nil
}

// StreamSecret is the domain controller for the StreamSecret RPC
func (s *Service) StreamSecret(req *pbhandshake.SecretMessage, stream pbhandshake.Handshake_StreamSecretServer) error {
	if err := stream.Send(&pbhandshake.SecretMessage{}); err != nil {
		return err
	}
	return nil
}
