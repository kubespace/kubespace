package user

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/types"
	"time"
)

type TokenManager struct {
	manager.CommonManager
}

func NewTokenManager(redisClient *redis.Client) *TokenManager {
	return &TokenManager{
		manager.CommonManager{
			ModelKey: "osp:token",
			Context:  context.Background(),
			Client:   redisClient,
		},
	}
}

func (tk *TokenManager) Create(tkObj *types.Token) error {

	if err := tk.CommonManager.Save(tkObj.Token.String(), tkObj, 43200*time.Second, false); err != nil {
		return err
	}
	return nil
}

func (tk *TokenManager) Get(name string) (*types.Token, error) {
	tkObj := &types.Token{}
	if err := tk.CommonManager.Get(name, tkObj); err != nil {
		return nil, err
	}
	return tkObj, nil
}
