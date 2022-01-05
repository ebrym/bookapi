package handlers

import (
	"fmt"

	"github.com/ebrym/bookapi/data"
	"github.com/ebrym/bookapi/service"
	"github.com/ebrym/bookapi/utils"
	"github.com/hashicorp/go-hclog"
)

// CategoryKey is used as a key for storing the Category object in context at middleware
type CategoryKey struct{}

// CategoryIDKey is used as a key for storing the CategoryID in context at middleware
type CategoryIDKey struct{}

// CategoryHandler wraps instances needed to perform operations on Category object
type CategoryHandler struct {
	logger      hclog.Logger
	configs     *utils.Configurations
	validator   *data.Validation
	repo        data.ICategoryRepository
	authService service.Authentication
}

// NewUserHandler returns a new UserHandler instance
func NewCategoryHandler(l hclog.Logger, c *utils.Configurations, v *data.Validation, r data.ICategoryRepository, auth service.Authentication) *CategoryHandler {
	return &CategoryHandler{
		logger:      l,
		configs:     c,
		validator:   v,
		repo:        r,
		authService: auth,
	}
}

type CategoryUpdate struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

var ErrCategoryAlreadyExists = fmt.Sprintf("Category already exists with the given email")
var ErrCategoryNotFound = fmt.Sprintf("No Category exists. Please sign in first")
var CategoryCreationFailed = fmt.Sprintf("Unable to create Category.Please try again later")
