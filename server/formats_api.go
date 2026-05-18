package server

import "context"

func (s *graphQLServer) Formats(ctx context.Context) ([]*GameFormat, error) {
	return formatDefinitions(), nil
}
