package user

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/types"
	"time"
)

type SessionManager struct {
	manager.CommonManager
}

func NewTokenManager(redisClient *redis.Client) *SessionManager {
	return &SessionManager{
		manager.CommonManager{
			ModelKey: "kubespace:user:session",
			Context:  context.Background(),
			Client:   redisClient,
		},
	}
}

func (tk *SessionManager) Create(tkObj *types.UserSession) error {

	if err := tk.CommonManager.Save(tkObj.SessionId.String(), tkObj, 43200*time.Second, false); err != nil {
		return err
	}
	return nil
}

func (tk *SessionManager) Get(name string) (*types.UserSession, error) {
	tkObj := &types.UserSession{}
	if err := tk.CommonManager.Get(name, tkObj); err != nil {
		return nil, err
	}
	return tkObj, nil
}
