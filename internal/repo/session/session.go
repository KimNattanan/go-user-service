package session

import (
	"context"
	"time"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/redis/go-redis/v9"
)

type SessionRepo struct {
	rdb *redis.Client
}

func NewSessionRepo(rdb *redis.Client) *SessionRepo {
	return &SessionRepo{rdb: rdb}
}

func (r *SessionRepo) Create(ctx context.Context, session *entity.Session) error {
	key := "session:" + session.ID
	ttl := time.Until(session.ExpiresAt)

	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, key, session)
	pipe.Expire(ctx, key, ttl)
	pipe.SAdd(ctx, "user_sessions:"+session.UserID, session.ID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *SessionRepo) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	var session entity.Session
	err := r.rdb.HGetAll(ctx, "session:"+id).Scan(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) FindByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	userSessionsKey := "user_sessions:" + userID
	sessionIDs, err := r.rdb.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, err
	}
	if len(sessionIDs) == 0 {
		return []*entity.Session{}, nil
	}

	pipe := r.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(sessionIDs))
	for i, id := range sessionIDs {
		cmds[i] = pipe.HGetAll(ctx, "session:"+id)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return []*entity.Session{}, err
	}

	var sessions []*entity.Session
	var staleIDs []string
	for i, cmd := range cmds {
		s := &entity.Session{}
		err := cmd.Scan(s)
		if err != nil || s.ID == "" {
			staleIDs = append(staleIDs, sessionIDs[i])
		} else {
			sessions = append(sessions, s)
		}
	}
	if len(staleIDs) > 0 {
		go r.rdb.SRem(ctx, userSessionsKey, staleIDs)
	}

	return sessions, nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id string) error {
	return r.rdb.HSet(ctx, "session:"+id, "is_revoked", true).Err()
}

func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, "session:"+id).Err()
}
