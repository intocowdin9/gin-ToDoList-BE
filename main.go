package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID    uint   `json:"id" form:"id"`
	Name  string `json:"name" form:"name"`
	Email string `json:"email" form:"email"`
}

type Todo struct {
	gorm.Model
	Title       string `json:title`
	Description string `json:description`
}

// ProductController represents a products-related controller
type ProductController struct{}

// GetProductList is a controller method to get product information
func (pl *ProductController) GetProductList(ctx *gin.Context) {
	productID := ctx.Param("id")
	// Fecth user information from the database or other data source
	// For Simplicity, we`ll just return a JSON response.
	ctx.JSON(200, gin.H{"id": productID, "name": "MSI", "price": "9.000.000"})
}

// UserController represents a users-related controller
type UserController struct{}

// GetUserInfo is a controller method to get user information
func (uc *UserController) GetUserInfo(ctx *gin.Context) {
	userID := ctx.Param("id")
	// Fecth user information from the database or other data source
	// For Simplicity, we`ll just return a JSON response.
	ctx.JSON(200, gin.H{"id": userID, "name": "John Doe", "email": "john@doe.com"})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey != "gintama" {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		ctx.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		duration := time.Since(start)
		log.Printf("Request-Method: %s | Status: %d | Duration: %v", ctx.Request.Method, ctx.Writer.Status(), duration)
	}
}

func main() {
	router := gin.Default()

	// authGroup := router.Group("/api")
	// authGroup.Use(AuthMiddleware())
	// {
	// 	authGroup.GET("/data", func(ctx *gin.Context) {
	// 		ctx.JSON(200, gin.H{"message": "Authenticated and autorized!"})
	// 	})
	// }

	// Connect to the SQLite database
	db, err := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate the Todo model to create the table
	db.AutoMigrate(&Todo{})

	// Auto-migrate the User model to create the table
	db.AutoMigrate(&User{})

	// Handle JSON data
	router.POST("/json", func(ctx *gin.Context) {
		var user User
		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid JSON data"})
			return
		}
		ctx.JSON(200, user)
	})

	// Handle form data
	router.POST("/form", func(ctx *gin.Context) {
		var user User
		if err := ctx.ShouldBind(&user); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid form data"})
			return
		}
		ctx.JSON(200, user)
	})

	// sHandle query parameters
	router.GET("/search", func(ctx *gin.Context) {
		query := ctx.DefaultQuery("q", "")
		// ctx.String(http.StatusOK, "Search query:"+query)

		var user []User

		if query == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search parameter 'q' is required!"})
			return
		}

		if err := db.Where("name LIKE ?", "%"+query+"%").Find(&user).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "No users matching the search query"})
			return
		}

		if len(user) == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		}

		ctx.JSON(http.StatusOK, user)
	})

	// Handle URL parameters
	// router.GET("/users/:id", func(ctx *gin.Context) {
	// 	userID := ctx.Param("id")
	// 	ctx.String(200, "User ID:"+userID)
	// })

	// Create a new user
	router.POST("/users", func(ctx *gin.Context) {
		var user User
		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid JSON data"})
			return
		}

		db.Create(&user)
		ctx.JSON(200, user)
	})

	// Retrieve users from db
	router.GET("/users", func(ctx *gin.Context) {
		var user []User

		db.Find(&user)
		ctx.JSON(200, user)
	})

	// Route to create a new Todo
	router.POST("/todos", func(ctx *gin.Context) {
		var todo Todo
		if err := ctx.ShouldBindJSON(&todo); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid JSON data"})
			return
		}

		// Save the TOdo to the database
		db.Create(&todo)
		ctx.JSON(200, todo)
	})

	// Route to get all Todos
	router.GET("/todos", func(ctx *gin.Context) {
		var todos []Todo

		// Retrieve all Todos from the database
		db.Find(&todos)

		ctx.JSON(200, todos)
	})

	// Route to get a specific Todo by ID
	router.GET("/todos/:id", func(ctx *gin.Context) {
		var todo Todo
		todoID := ctx.Param("id")

		// Retrieve the Todo from database
		result := db.First(&todo, todoID)
		if result.Error != nil {
			ctx.JSON(404, gin.H{"error": "Todo not found"})
			return
		}

		ctx.JSON(200, todo)
	})

	// Route to update a Todo by ID
	router.PUT("/todos/:id", func(ctx *gin.Context) {
		var todo Todo
		todoID := ctx.Param("id")

		// Retrieve the Todo from the database
		result := db.First(&todo, todoID)
		if result.Error != nil {
			ctx.JSON(404, gin.H{"error": "Todo not found"})
			return
		}

		var updatedTodo Todo
		if err := ctx.ShouldBindJSON(&updatedTodo); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid JSON data"})
		}

		// Update the Todo in the database
		todo.Title = updatedTodo.Title
		todo.Description = updatedTodo.Description
		db.Save(&todo)

		ctx.JSON(200, todo)
	})

	// Route to delete a Todo by ID
	router.DELETE("/todos/:id", func(ctx *gin.Context) {
		var todo Todo
		todoID := ctx.Param("id")

		// Retrieve the Todo from the database
		result := db.First(&todo, todoID)
		if result.Error != nil {
			ctx.JSON(404, gin.H{"error": "Todo not found"})
			return
		}

		// Delete the Todo from the database
		db.Delete(&todo)

		ctx.JSON(200, gin.H{"message": fmt.Sprintf("Todo with ID %s deleted", todoID)})
	})

	// userController := &UserController{}
	productController := &ProductController{}

	// Route using the UserController
	// router.GET("/users/:id", userController.GetUserInfo)

	// Route using the ProductController
	router.GET("/product/:id", productController.GetProductList)

	router.Use(LoggerMiddleware())

	// Basic route
	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Hello Gin!")
	})

	//  Route with URL parameters
	// router.GET("/users/:id", func(ctx *gin.Context) {
	// 	id := ctx.Param("id")
	// 	ctx.String(200, "UserID:"+id)
	// })

	// Public routes (no authentication required)
	public := router.Group("/public")
	{
		public.GET("/info", func(ctx *gin.Context) {
			ctx.String(200, "Public information")
		})
	}

	// Private routes (require authentication)
	private := router.Group("/private")
	private.Use(AuthMiddleware())
	{
		private.GET("/data", func(ctx *gin.Context) {
			ctx.String(200, "Private data accessible after authentication")
		})
		private.POST("/create", func(ctx *gin.Context) {
			ctx.String(200, "Create a new resource")
		})
	}
	router.Run(":8080")
}
