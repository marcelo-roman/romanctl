// Package services holds the application rules.
package services

import "context"

// SampleItemService gathers the SampleItem operations.
type SampleItemService struct{}

// Get finds by identifier.
func (s *SampleItemService) Get(ctx context.Context, id string) (string, error) {
	return id, nil
}
