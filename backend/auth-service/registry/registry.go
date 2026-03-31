package registry

// import (
// 	authCtrl "auth-service/controllers/auth"
// 	"auth-service/pkg/middleware"
// 	"auth-service/repositories"
// 	"auth-service/routes"
// 	authService "auth-service/services/auth"
// 	mfaService "auth-service/services/mfa"
// 	userService "auth-service/services/user"
// )

// // initMiddleware initializes all middleware
// func (r *Registry) initMiddleware() {
// 	r.AuthMiddleware = middleware.NewAuthMiddleware(r.Config.JWT.Secret)
// }

// // initRepositories initializes all repositories
// func (r *Registry) initRepositories() {
// 	r.UserRepository = repositories.NewUserRepository(r.DB)
// 	r.MFASetupRepository = repositories.NewMFASetupRepository(r.DB)
// }

// // initServices initializes all services
// func (r *Registry) initServices() {
// 	r.AuthService = authService.NewAuthService(r.UserRepository, r.Redis, r.Config, r.KafkaProducer)
// 	r.UserService = userService.NewUserService(r.UserRepository, r.KafkaProducer)
// 	r.MfaService = mfaService.NewMfaService(r.MFASetupRepository, r.UserRepository, r.Redis)
// }

// // initControllers initializes all controllers
// func (r *Registry) initControllers() {
// 	r.AuthController = authCtrl.NewAuthController(r.Validator, r.AuthService)
// }

// // RegisterRoutes registers all routes
// func (r *Registry) RegisterRoutes() {
// 	routes.RegisterAuthRoutes(r.GinEngine, r.AuthController, r.AuthMiddleware)
// }
