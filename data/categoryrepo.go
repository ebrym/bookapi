package data

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

// CategoryRepository has the implementation of the db methods.
type CategoryRepository struct {
	db     *sqlx.DB
	logger hclog.Logger
}

// NewCategoryRepository returns a new CategoryRepository instance
func NewCategoryRepository(db *sqlx.DB, logger hclog.Logger) *CategoryRepository {
	return &CategoryRepository{db, logger}
}

// Create inserts the given user into the database
func (repo *CategoryRepository) Create(ctx context.Context, category *Category) error {
	category.ID = uuid.NewV4().String()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	repo.logger.Info("creating Category", hclog.Fmt("%#v", category))
	query := "insert into category (id, name, code, createdat, updatedat) values ($1, $2, $3, $4, $5)"
	_, err := repo.db.ExecContext(ctx, query, category.ID, category.Name, category.Code, category.CreatedAt, category.UpdatedAt)
	return err
}

// GetUserByEmail retrieves the user object having the given email, else returns error
func (repo *CategoryRepository) GetCategoryByCode(ctx context.Context, code string) (*Category, error) {
	repo.logger.Debug("querying for category with Code", code)
	query := "select * from category where code = $1"
	var category Category
	if err := repo.db.GetContext(ctx, &category, query, code); err != nil {
		return nil, err
	}
	repo.logger.Debug("read category", hclog.Fmt("%#v", category))
	return &category, nil
}

// GetcategoryByID retrieves the category object having the given ID, else returns error
func (repo *CategoryRepository) GetCategoryByID(ctx context.Context, categoryID string) (*Category, error) {
	repo.logger.Debug("querying for category with id", categoryID)
	query := "select * from category where id = $1"
	var category Category
	if err := repo.db.GetContext(ctx, &category, query, categoryID); err != nil {
		return nil, err
	}
	return &category, nil
}

// GetcategoryByID retrieves the category object having the given ID, else returns error
func (repo *CategoryRepository) GetCategories(ctx context.Context) ([]Category, error) {
	repo.logger.Debug("querying for all category")
	query := "select * from category"

	//category := Category{}
	//category, err := repo.db.GetContext(ctx, &CategoryList, query)
	cur, err := repo.db.QueryContext(ctx, query)
	//var category CategoryList
	if err != nil {
		return nil, err
	}

	var categories []Category
	for cur.Next() {
		category := Category{}
		if err := cur.Scan(&category.ID, &category.Name, &category.Code, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

// Updatecategory updates the category of the given category
func (repo *CategoryRepository) UpdateCategory(ctx context.Context, category *Category) error {
	category.UpdatedAt = time.Now()

	query := "update category set name = $1,code = $2, updatedat = $3 where id = $4"
	if _, err := repo.db.ExecContext(ctx, query, category.Name, category.Code, category.UpdatedAt, category.ID); err != nil {
		return err
	}
	return nil
}
