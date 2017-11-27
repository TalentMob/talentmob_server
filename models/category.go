package models

import (
	"github.com/rathvong/talentmob_server/system"
	"log"
	"database/sql"
	"fmt"
	"time"
	"strings"
)

const (
	defaultTagColor = "#000000"
)

// Handles all categories and subcategories in the app
// each category can be color coded and set with its own custom buttons
type Category struct {
	BaseModel
	CategoryID   uint64 `json:"category_id"`
	Color        string `json:"color"`
	Title        string `json:"title"`
	VideoCount   int    `json:"video_count"`
	IconActive   string `json:"icon_active"`
	IconInActive string `json:"icon_inactive"`
	Position     int    `json:"position"`
	IsActive     bool   `json:"is_active"`
}

// SQL create a new category
func (c * Category) queryCreate() (qry string){
	return `INSERT INTO categories
							(category_id,
							color,
							title,
							icon_active,
							icon_inactive,
							position,
							is_active,
							video_count,
							created_at,
							updated_at)
							VALUES
							($1, $2, $3, $4, $5, $6, $7, $8, $9)
							RETURNING id
							`
}

// SQL update a new category
func (c *Category) queryUpdate() (qry string){
	return `UPDATE categories SET
					category_id = $2,
					color = $3,
					title = $4,
					icon_active = $5,
					icon_inactive = $6,
					position = $7,
					video_count = $8,
					is_active = $9,
					updated_at = $10
				WHERE id = $1`
}

// SQL query get a category by id
func (c *Category) queryGet() (qry string) {
	return `SELECT
						id,
						category_id,
						color,
						title,
						icon_active,
						icon_inactive,
						position,
						video_count,
						is_active,
						created_at,
						updated_at
				FROM categories
				WHERE id = $1`
}

// SQL query get a category by title
func (c *Category) queryGetByTitle() (qry string){
	return `SELECT
						id,
						category_id,
						color,
						title,
						icon_active,
						icon_inactive,
						position,
						video_count,
						is_active,
						created_at,
						updated_at
				FROM categories
				WHERE title = $1`
}

// SQL query get a list by title
func (c *Category) queryGetListByName() (qry string) {
	return `SELECT
						id,
						category_id,
						color,
						title,
						icon_active,
						icon_inactive,
						position,
						video_count,
						is_active,
						created_at,
						updated_at
				FROM categories
				WHERE title IN (%v)`
}

// SQL query get a list by id
func (c *Category) queryGetListByID() (qry string){
	return `SELECT
						id,
						category_id,
						color,
						title,
						icon_active,
						icon_inactive,
						position,
						video_count,
						is_active,
						created_at,
						updated_at
				FROM categories
				WHERE id IN (%V)`
}

func (c *Category) queryMainCategories() (qry string){
	return `SELECT
						categories.id,
						categories.category_id,
						categories.color,
						categories.title,
						categories.icon_active,
						categories.icon_inactive,
						categories.position,
						categories.video_count,
						categories.is_active,
						categories.created_at,
						categories.updated_at
				FROM categories
				INNER JOIN categories main
				ON main.id = categories.category_id
				WHERE
				main.title = 'main'
				ORDER BY categories.position ASC
				`
}


func (c *Category) queryTopCategories() (qry string) {
	return `SELECT
						categories.id,
						categories.category_id,
						categories.color,
						categories.title,
						categories.icon_active,
						categories.icon_inactive,
						categories.position,
						categories.video_count,
						categories.is_active,
						categories.created_at,
						categories.updated_at
				FROM categories
				WHERE
				categories.title != 'main'
				ORDER BY categories.category_id DESC, categories.position ASC, categories.video_count DESC
				LIMIT $1
				OFFSET $2
				`
}

func (c * Category) queryExistByTag() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM categories where title = $1)`
}


// Validate insertion errors when creating a new category
func (c *Category) validateCreateErrors() (err error){

	if c.Title == "" {
		return c.Errors(ErrorMissingValue, "title")
	}


	if c.Color == ""{
		return c.Errors(ErrorMissingValue, "color")
	}

	return
}

// Validate missing requirements for updates in categories
func (c *Category) validateUpdateErrors() (err error) {
	if c.ID == 0 {
		return c.Errors(ErrorMissingValue, "id")
	}

	return c.validateCreateErrors()
}

// convert tags from video creating and saves it into the database
func (c *Category) CreateNewCategoriesFromTags(db *system.DB, tags string, video Video) {

	array := strings.Split(tags, "#")

	for _, tag := range array {
		exists, err := c.ExistsByTag(db, tag)

		if err != nil {
			continue
		}

		var category Category
		var title = strings.ToLower(tag)

		if exists {

			if err = category.GetByTitle(db, tag); err != nil {
				continue
			}

			category.VideoCount++

			if err = category.Update(db); err != nil {
				continue
			}

		} else {

			category.Title = title
			category.Color = defaultTagColor
			category.IsActive = true
			category.VideoCount = 1

			if err = category.Create(db); err != nil {
				continue
			}

		}

		// Create an associating between a video and category
		t := Tag{}
		t.Title = title
		t.VideoID = video.ID
		t.CategoryID = category.ID

		if err = t.Create(db); err != nil {
			continue
		}

	}

	return
}

// checks db is a tag exists
func (c *Category) ExistsByTag(db *system.DB, tag string) (exists bool, err error){
	if tag == "" {
		return false, c.Errors(ErrorMissingValue, "tag")
	}


	err = db.QueryRow(c.queryExistByTag(), tag).Scan(&exists)

	if err != nil {

		log.Printf("Category.ExistsByTag() tag -> %v QueryRow() -> %v Error -> %v", tag, c.queryExistByTag(), err)
		return
	}

	return
}


// Create a new category
func (c *Category) Create(db *system.DB) (err error){
	if err = c.validateCreateErrors(); err != nil {
		log.Println("Category.Create() Error -> ", err)
		return err
	}

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	if err != nil {
		log.Println("Category.Create() Begin() Error -> ", err)
		return
	}

	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	err = tx.QueryRow(c.queryCreate(),
		c.CategoryID,
		c.Color,
		c.Title,
		c.IconActive,
		c.IconInActive,
		c.Position,
		c.VideoCount,
		c.IsActive,
		c.CreatedAt,
		c.UpdatedAt,
			).Scan(&c.ID)

	if err != nil {
		log.Printf("Category.Create() QueryRow() -> %v Error -> %v", c.queryCreate(), err)
		return
	}


	return
}

// Update a new category
func (c *Category) Update(db *system.DB) (err error){

	if err = c.validateUpdateErrors(); err != nil {
		log.Println("Category.Update() Error -> ", err)
		return err
	}

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	if err != nil {
		log.Println("Category.Update() Begin() -> ", err)
		return
	}

	c.UpdatedAt = time.Now()

	_, err = tx.Exec(c.queryUpdate(),
		c.ID,
		c.CategoryID,
		c.Color,
		c.Title,
		c.IconActive,
		c.IconInActive,
		c.Position,
		c.VideoCount,
		c.IsActive,
		c.UpdatedAt,
		)


	if err != nil {
		log.Printf("Category.update() id -> %v Exec() -> %v Error -> %v", c.ID, c.queryUpdate(), err)
		return
	}

	return
}

// Get a category
func (c *Category) Get(db *system.DB, categoryID uint64) (err error){

	if categoryID == 0 {
		err  =  c.Errors(ErrorMissingValue, "id")
		log.Println("Category.get() Error -> ", err)
		return
	}

	err = db.QueryRow(c.queryGet(), categoryID).Scan(
		&c.ID,
		&c.CategoryID,
		&c.Color,
		&c.Title,
		&c.IconActive,
		&c.IconInActive,
		&c.Position,
		&c.VideoCount,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Category.Get() categoryID -> %v QueryRow() -> %v Error -> %v", categoryID, c.queryGet(), err)
		return
	}

	return
}

// Get a category by title
func (c *Category) GetByTitle(db *system.DB, title string) (err error){

	if title == "" {
		err  =  c.Errors(ErrorMissingValue, "title")
		log.Println("Category.getByTitle() Error -> ", err)
		return
	}

	err = db.QueryRow(c.queryGetByTitle(), title).Scan(
		&c.ID,
		&c.CategoryID,
		&c.Color,
		&c.Title,
		&c.IconActive,
		&c.IconInActive,
		&c.Position,
		&c.VideoCount,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Category.GetByTitle() title -> %v QueryRow() -> %v Error -> %v",title, c.queryGetByTitle(), err)
		return
	}

	return
}

// Get a list by titles
func (c *Category) GetListByTitles(db *system.DB, titleArray string) (categories []Category, err error){
	if titleArray == ""{
		err = c.Errors(ErrorMissingValue, "titleArray")
		log.Println("Category.GetListByTitles() Error -> ", err)
		return
	}

	rows, err := db.Query(fmt.Sprintf(c.queryGetListByName(), titleArray))

	defer  rows.Close()

	if err != nil {
		log.Printf("Category.GetListByTitles() titleArray -> %v, Query() -> %v Error -> %v", titleArray, c.queryGetByTitle(), err)
		return
	}

	return c.parseRows(rows)
}


// Get a list by ids
func (c *Category) GetListByIDs(db *system.DB, ids string)  (categories []Category, err error) {
	if ids == ""{
		err = c.Errors(ErrorMissingValue, "titleArray")
		log.Println("Category.GetListByIds() Error -> ", err)
		return
	}

	rows, err := db.Query(fmt.Sprintf(c.queryGetListByID(), ids))

	defer  rows.Close()

	if err != nil {
		log.Printf("Category.GetListByIds() titleArray -> %v Query() -> %v Error -> %v", ids, c.queryGetListByID(), err)
		return
	}

	return c.parseRows(rows)
}

// Retrieve all main categories
func (c *Category) GetMainCategories(db *system.DB) (categories []Category, err error){

	rows, err := db.Query(c.queryMainCategories())

	defer  rows.Close()

	if err != nil {
		log.Printf("Category.GetMainCategories() Query -> %v Error -> %v", c.queryMainCategories(), err)
		return
	}


	return c.parseRows(rows)
}


func (c *Category) GetTopCategories(db *system.DB, page int) (categories []Category, err error) {

	rows, err := db.Query(c.queryTopCategories(), LimitQueryPerRequest, offSet(page))


	defer rows.Close()

	if err != nil {
		log.Printf("Category.GetTopCategories() Query() -> %v Error -> %v", c.queryTopCategories(), err)
		return
	}

	return c.parseRows(rows)
}

// Parse each row returned from query
func (c *Category) parseRows(rows *sql.Rows) (categories []Category, err error){

	for rows.Next() {
		category := Category{}

		err = rows.Scan(
			&category.ID,
			&category.CategoryID,
			&category.Color,
			&category.Title,
			&category.IconActive,
			&category.IconInActive,
			&category.Position,
			&category.VideoCount,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt)

		if err != nil {
			log.Println("Category.parseRows() Error -> ", err)
			return
		}


		categories = append(categories, category)

	}

	return
}