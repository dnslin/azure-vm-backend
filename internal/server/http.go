package server

import (
	apiV1 "azure-vm-backend/api/v1"
	"azure-vm-backend/docs"
	"azure-vm-backend/internal/handler"
	"azure-vm-backend/internal/middleware"
	"azure-vm-backend/pkg/jwt"
	"azure-vm-backend/pkg/log"
	"azure-vm-backend/pkg/server/http"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHTTPServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	userHandler *handler.UserHandler,
	accountsHandler *handler.AccountsHandler,
) *http.Server {
	gin.SetMode(gin.DebugMode)
	s := http.NewServer(
		gin.Default(),
		logger,
		http.WithServerHost(conf.GetString("http.host")),
		http.WithServerPort(conf.GetInt("http.port")),
	)

	// swagger doc
	docs.SwaggerInfo.BasePath = "/v1"
	s.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		//ginSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", conf.GetInt("app.http.port"))),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	))

	s.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		//middleware.SignMiddleware(log),
	)
	s.GET("/", func(ctx *gin.Context) {
		logger.WithContext(ctx).Info("hello")
		apiV1.HandleSuccess(ctx, map[string]interface{}{
			":)": "Thank you for using Azure-VM-Backend",
		})
	})

	v1 := s.Group("/v1")
	{
		// No route group has permission
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/register", userHandler.Register)
			noAuthRouter.POST("/login", userHandler.Login)
		}
		// Non-strict permission routing group
		noStrictAuthRouter := v1.Group("/").Use(middleware.NoStrictAuth(jwt, logger))
		{
			noStrictAuthRouter.GET("/user", userHandler.GetProfile)
		}

		// Strict permission routing group
		strictAuthRouter := v1.Group("/").Use(middleware.StrictAuth(jwt, logger))
		{
			// 用户接口
			strictAuthRouter.POST("/user", userHandler.UpdateProfile)
			// 账户接口
			strictAuthRouter.POST("/accounts/create", accountsHandler.CreateAccounts)
			strictAuthRouter.DELETE("/accounts/delete", accountsHandler.DeleteAccounts)
			strictAuthRouter.POST("/accounts/list", accountsHandler.ListAccounts)

			strictAuthRouter.POST("/accounts/update/:id", accountsHandler.UpdateAccount)
			strictAuthRouter.GET("/accounts/:id", accountsHandler.GetAccount)
		}
	}

	return s
}
