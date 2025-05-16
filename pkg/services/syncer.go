package services

import (
	"context"

	"github.com/vpnda/sandwich-sync/db"
	"github.com/vpnda/sandwich-sync/pkg/http/lm"
)

type LunchMoneySyncer struct {
	client        lm.LunchMoneyClientInterface
	database      db.DBInterface
	accountMapper *AccountMapper
	forceSync     bool
}

func NewLunchMoneySyncer(ctx context.Context, apiKey string, database db.DBInterface) (*LunchMoneySyncer, error) {
	c, err := lm.NewLunchMoneyClient(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	as, err := NewAccountMapper(ctx, apiKey, database)
	if err != nil {
		return nil, err
	}

	return &LunchMoneySyncer{
		client:        c,
		database:      database,
		accountMapper: as,
	}, nil
}

func (s *LunchMoneySyncer) GetAccountMapper() *AccountMapper {
	return s.accountMapper
}

func (s *LunchMoneySyncer) GetClient() lm.LunchMoneyClientInterface {
	return s.client
}
