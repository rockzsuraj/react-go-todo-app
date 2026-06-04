package repository

import (
	"context"
	"time"
)

type NoopBlacklistRepository struct{}

func (n *NoopBlacklistRepository) Add(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (n *NoopBlacklistRepository) IsBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (n *NoopBlacklistRepository) BlacklistAllForUser(_ context.Context, _ string) error {
	return nil
}

func (n *NoopBlacklistRepository) IsUserBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (n *NoopBlacklistRepository) UnblockUser(_ context.Context, _ string) error {
	return nil
}
