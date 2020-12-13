package main

import (
	"github.com/josephsalimin/go-simple-ioc/ioc"
	"log"
)

type Config struct {
	IsDebug bool
}

type User struct {
	id uint64
}

type UserRepository interface {
	GetUser(id uint64) (*User, error)
}

type userRepository struct{}

func (r *userRepository) GetUser(id uint64) (*User, error) {
	return &User{id: id}, nil
}

type UserService interface {
	FetchProfile(id uint64) (*User, error)
}

type userService struct {
	// You can also defines tag with ioc tag to fetch dependency with alias.
	cfg        *Config `ioc:"service_cfg"`
	repository UserRepository
}

func (s *userService) FetchProfile(id uint64) (*User, error) {
	user, err := s.repository.GetUser(id)
	if err != nil {
		return nil, err
	}
	if s.cfg.IsDebug {
		log.Printf("user: %+v", user)
	}

	return user, nil
}

func main() {
	cfg := &Config{IsDebug: true}
	ioc.MustBindSingleton(func() *Config {
		return cfg
	}, ioc.WithBindAlias("service_cfg"))
	ioc.MustBindSingleton(func() UserRepository {
		return &userRepository{}
	}, ioc.WithBindMeta(&userRepository{}))
	// You must define dependencies that you want to inject as parameter.
	ioc.MustBindSingleton(func(cfg *Config, ur UserRepository) UserService {
		return &userService{cfg: cfg, repository: ur}
	}, ioc.WithBindMeta(&userService{}))

	var s UserService
	// Internal container will resolve the dependency for you.
	if err := ioc.Resolve(&s); err != nil {
		log.Fatal(err)
	}

	// The variable will be filled with userService instance.
	if v, ok := s.(*userService); ok {
		log.Printf("actual implementation: %+v\n", v)
	}

	u, _ := s.FetchProfile(0)
	log.Printf("user result: %+v\n", u)
}
