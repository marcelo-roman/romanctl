// Package services contém as regras da aplicação.
package services

import "context"

// SampleItemService concentra as operações de SampleItem.
type SampleItemService struct{}

// Get busca pelo identificador.
func (s *SampleItemService) Get(ctx context.Context, id string) (string, error) {
	return id, nil
}
