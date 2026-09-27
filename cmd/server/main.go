package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"react-go-cms-auth-service/internal/application/service/permission"
	"react-go-cms-auth-service/internal/application/service/role"
	"react-go-cms-auth-service/internal/application/service/user"
	"react-go-cms-auth-service/internal/handler"
	"react-go-cms-auth-service/internal/infrastructure/configuration"
	"react-go-cms-auth-service/internal/infrastructure/httpx"
	"react-go-cms-auth-service/internal/infrastructure/jwt"
	"react-go-cms-auth-service/internal/infrastructure/repository/repo"
)

const (
	defaultPort         = "8081"
	databasePingTimeout = 10 * time.Second
	readHeaderTimeout   = 10 * time.Second
)

func main() {
	var cfg configuration.Config
	cfg = configuration.LoadConfig(defaultPort)
	var db *sql.DB
	var err error
	db, err = configuration.OpenDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database: %v", err)
	}
	var userRepository *repo.UserRepository
	userRepository = repo.NewUserRepository(db)
	var roleRepository *repo.RoleRepository
	roleRepository = repo.NewRoleRepository(db)
	var permissionRepository *repo.PermissionRepository
	permissionRepository = repo.NewPermissionRepository(db)
	var tokens *jwt.Issuer
	tokens = jwt.NewIssuer(cfg.JWTSecret, cfg.JWTKeyID, cfg.JWTIssuer)
	var lifespan time.Duration
	lifespan = time.Duration(cfg.JWTLifespan) * time.Second
	var userService *user.Service
	userService = user.New(userRepository, tokens, lifespan)
	var roleService *role.Service
	roleService = role.New(roleRepository)
	var permissionService *permission.Service
	permissionService = permission.New(permissionRepository)
	var httpHandler *handler.Handler
	httpHandler = handler.New(userService, roleService, permissionService, cfg.JWTSecret, cfg.JWTIssuer)
	var mux *http.ServeMux
	mux = http.NewServeMux()
	httpHandler.Register(mux)

	var addr string
	addr = ":" + cfg.Port
	log.Printf("react-go-cms-auth-service listening on %s", addr)
	var server *http.Server
	server = &http.Server{
		Addr:              addr,
		Handler:           httpx.CORS(cfg.CORSOrigins, mux),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	log.Fatal(server.ListenAndServe())
}
