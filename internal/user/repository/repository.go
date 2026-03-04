package userrepo

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
)

type User struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	Is_active bool   `json:"is_active"`
}

type Repo struct {
	DB      *pgxpool.Pool
	RedisDB *redis.Client
	sfg     singleflight.Group
}

func NewRepo(DB *pgxpool.Pool, RedisDB *redis.Client) *Repo {
	return &Repo{
		DB:      DB,
		RedisDB: RedisDB,
	}
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.DB.QueryRow(ctx, "SELECT id,username,password,role,is_active FROM users WHERE username = $1", username).Scan(
		&u.Id,
		&u.Username,
		&u.Password,
		&u.Role,
		&u.Is_active,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil

}

func (r *Repo) GetUserById(ctx context.Context, id string) (*User, error) {
	UserRaw, err := r.RedisDB.Get(ctx, "user:"+id).Result()
	if err == nil {
		var user User
		err := json.Unmarshal([]byte(UserRaw), &user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}
	if err == redis.Nil {
		val, err, _ := r.sfg.Do(id, func() (interface{}, error) {
			var user User
			err = r.DB.QueryRow(ctx, "SELECT id,username,role,is_active FROM users WHERE id = $1", id).Scan(
				&user.Id,
				&user.Username,
				&user.Role,
				&user.Is_active,
			)
			return &user, err
		})

		if err != nil {
			return nil, err
		}
		user, _ := val.(*User)
		tmp, err := json.Marshal(&user)
		if err != nil {
			return nil, err
		}
		ttl := 3*time.Minute + time.Duration(rand.Int63n(30))*time.Second
		_ = r.RedisDB.Set(ctx, "user:"+id, tmp, ttl).Err()

		return user, nil

	}
	return nil, err

}
func (r *Repo) CreateUser(ctx context.Context, username string, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		var id string
		err = r.DB.QueryRow(ctx, "INSERT INTO users (username,password,role) VALUES ($1,$2,$3) RETURNING id", username, hash, "user").Scan(&id)
		if err != nil {
			return "", err
		}
		return id, nil
	}
	return "", err
}

func (r *Repo) CreateAdmin(ctx context.Context, username string, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		var id string
		err = r.DB.QueryRow(ctx, "INSERT INTO users (username,password,role) VALUES ($1,$2,$3) RETURNING id", username, hash, "admin").Scan(&id)
		if err != nil {
			return "", err
		}
		return id, nil
	}
	return "", err
}

func (r *Repo) SetActiveStatus(ctx context.Context, UserId string) error {
	_, err := r.DB.Exec(ctx, "UPDATE users SET is_active=true WHERE id = $1", UserId)
	if err != nil {
		return err
	}
	return nil
}

// доработать касаемо когда у нас пользователь выходит ручку exit + хочу сделать такую функцию что типо
// либо почту добавить что-то такое чтобы ещё с api другим можно было взаимодействовать вот так
