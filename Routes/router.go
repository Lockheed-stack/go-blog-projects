package Routes

import (
	"BlogProject/Controller"
	"BlogProject/Middlewares"

	"github.com/gin-gonic/gin"
)

func InitRouter() {

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	// r := gin.New()
	// r.Use(Middlewares.Logger())
	// r.Use(cors.Default())
	r.Use(Middlewares.Cors())
	// r.SetTrustedProxies([]string{
	// 	"1.1.1.1",
	// })

	public_route := r.Group("/controller")
	{
		// login
		public_route.POST("login", Controller.Login)

		//user
		public_route.GET("users", Controller.GetUser)
		public_route.POST("user/add", Controller.AddUser)
		// router_Controller.PUT("user/:id", Controller.EditUser)

		//calalog
		public_route_catalog := public_route.Group("/categories")
		public_route_catalog.Use(Middlewares.RedisCatalogs())
		{
			public_route_catalog.GET("", Controller.QueryCatalog)
			public_route_catalog.GET("/list", Controller.QueryAllCategoriesWithAllArticles)
		}

		//article
		public_route_article := public_route.Group("/article")
		public_route_article.Use(Middlewares.RedisArticles())
		{
			public_route_article.GET("/list", Controller.QueryArticles)
			public_route_article.GET("/list/last", Controller.QueryLast3Articles)
			public_route_article.GET("/list/hot", Controller.QueryHot3Articles)
			public_route_article.GET("/list/:cid", Controller.QueryArticlesInSameCatalog)
			public_route_article.GET("/:id", Controller.QuerySingleArticle)
		}

	}
	auth_route := r.Group("/controller")
	auth_route.Use(Middlewares.JwtToken())
	{

		// user
		auth_route.DELETE("user/:id", Controller.RemoveUser)
		// catalog
		auth_route.POST("category/add", Controller.AddCatalog)
		auth_route.PUT("category/:id", Controller.EditCatalog)
		auth_route.DELETE("category/:id", Controller.RemoveCatalog)
		// article
		auth_route.POST("article/add", Controller.AddArticle)
		auth_route.PUT("article/update/:id", Controller.EditArticle)
		auth_route.DELETE("article/:id", Controller.RemoveArticle)
		// upload
		auth_route.POST("upload", Controller.Upload)
		// check login status
		auth_route.POST("checkLoginStatus", Controller.TokenCheck)
	}
	r.Run("0.0.0.0:8080")
}
