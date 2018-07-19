package badgecontroller

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rathvong/talentmob_server/models"
	"github.com/rathvong/talentmob_server/system"
)

const (
	TriggerMoreThan     = "more_than"
	TriggerWithInOneDay = "with_in_one_day"
)

type Badge struct {
	models.BaseModel
	Object      string `json:"object"`
	Field       string `json:"field"`
	Trigger     string `json:"trigger"`
	Value       uint64 `json:"value"`
	Reward      int    `json:"reward"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	HasBadge    bool   `json:"has_badge"`
	IsActive    bool   `json:"is_active"`
}

func (b *Badge) List(db *system.DB, userID uint64) ([]Badge, error) {

	if userID == 0 {
		return nil, b.Errors(models.ErrorMissingID, "Badge: id")
	}

	qry := fmt.Sprintf(`SELECT  id,
								object,
								field,
								trigger,
								value,
								reward,
								title,
								description,
								icon,
								is_active,
								updated_at,
								created_at,
								(SELECT EXISTS(SELECT 1 FROM acheivements WHERE badge_id = id AND user_id = %d AND is_active=true) as has_badge
						FROM	badges
						ORDER BY title ASC`, userID)

	rows, err := db.Query(qry)

	defer rows.Close()

	if err != nil {
		log.Printf("Query -> %s Error -> %s", qry, err)
		return nil, err
	}

	return b.parseRows(rows)
}

func (b *Badge) parseRows(rows *sql.Rows) ([]Badge, error) {

	var badges []Badge

	for rows.Next() {
		var badge Badge

		err := rows.Scan(
			&badge.ID,
			&badge.Object,
			&badge.Field,
			&badge.Trigger,
			&badge.Value,
			&badge.Reward,
			&badge.Title,
			&badge.Description,
			&badge.Icon,
			&badge.IsActive,
			&badge.UpdatedAt,
			&badge.CreatedAt,
			&badge.HasBadge,
		)

		if err != nil {
			return nil, err
		}

		badges = append(badges, badge)

	}

	return badges, nil
}

type Achievement struct {
	models.BaseModel
	UserID   uint64 `json:"user_id"`
	BadgeID  uint64 `json:"badge_id"`
	IsActive bool   `json:"is_active"`
}

func (a *Achievement) Create(db *system.DB) error {

	if a.UserID == 0 {
		return a.Errors(models.ErrorMissingID, "Achievement: id")
	}

	if a.BadgeID == 0 {
		return a.Errors(models.ErrorMissingID, "Achievement: badge_id")
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
		return err
	}

	qry := `INSERT INTO achievements
						(user_id, badge_id, is_active, updated_at, created_at)
						VALUES
						($1, $2, $3, $4, $5)
						RETURNING id`

	a.UpdatedAt = time.Now()
	a.CreatedAt = time.Now()
	a.IsActive = true

	err = tx.QueryRow(
		qry,
		a.UserID,
		a.BadgeID,
		a.IsActive,
		a.UpdatedAt,
		a.CreatedAt,
	).Scan(&a.ID)

	if err != nil {
		log.Printf("UserID -> %d BadgeID -> %d Query() -> %s Error() -> %s", qry, err)
		return err
	}

	return nil
}

func (a *Achievement) HasBadge(db *system.DB, badgeID uint64, userID uint64) (bool, error) {

	if userID == 0 {
		return false, a.Errors(models.ErrorMissingID, "Achievement: id")
	}

	if badgeID == 0 {
		return false, a.Errors(models.ErrorMissingID, "Achievement: badge_id")
	}

	var exists bool

	qry := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM acheivements WHERE badge_id = %d AND user_id = %d AND is_active=true)", badgeID, userID)

	err := db.QueryRow(qry).Scan(&exists)

	if err != nil {
		log.Printf("Query() -> %s Error() -> %s", qry, err)
	}

	return exists, err
}

func CalculateReward(db *system.DB, userID uint64, object string, objectID uint64, field string) (*Achievement, error) {

	b := new(Badge)
	badges, err := b.List(db, userID)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	var achievement *Achievement

	for _, badge := range badges {

		if badge.HasBadge {
			continue
		}

		if badge.Object != object {
			continue
		}

		if badge.Field != field {
			continue
		}

		handler := BadgeHandler{db, userID, &badge, objectID}

		switch object {
		case "video":
			achievement, err = handler.handleVideoBadges()
		case "user":
			achievement, err = handler.handleUserBadges()

		default:
			return nil, errors.New("object not supported")
		}

		if err != nil {
			return nil, err
		}

		// if achievement != nil {
		// 	models.Notify(db, 0, userID, "achievement", badge.ID, "badge")
		// }

	}

	return achievement, err
}

type BadgeHandler struct {
	db       *system.DB
	userID   uint64
	badge    *Badge
	objectID uint64
}

func (b *BadgeHandler) handleVideoBadges() (*Achievement, error) {

	if b.objectID == 0 {
		return nil, errors.New("missing object_id")
	}

	switch b.badge.Field {

	case "like_count":
		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM votes WHERE video_id = %d AND upvote > 0", b.objectID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}
	case "view_count":
		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM views WHERE video_id = %d", b.objectID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}
	case "comment_count":
		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM comments WHERE video_id = %d", b.objectID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	case "share_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

		case TriggerWithInOneDay:
		}

	case "favourite_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

		case TriggerWithInOneDay:
		}

	case "boost_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM boosts WHERE video_id = %d", b.objectID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	}

	return nil, nil
}

func (b *BadgeHandler) handleUserBadges() (*Achievement, error) {

	switch b.badge.Field {

	case "import_count":
		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM videos WHERE video_id = %d", b.userID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}
	case "fan_count":
		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM relationships WHERE followed_id = %d", b.userID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}
	case "following_count":
		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM relationships WHERE follower_id = %d", b.userID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	case "total_comment_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM comments WHERE user_id = %d", b.userID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	case "total_upvote_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM votes WHERE user_id = %d AND upvote > 0", b.userID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	case "total_downvote_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM votes WHERE user_id = %d AND downvote > 0", b.objectID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	case "total_view_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM views WHERE user_id = %d", b.userID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	case "total_favourite_count":
		switch b.badge.Trigger {
		case TriggerMoreThan:

		case TriggerWithInOneDay:
		}

	case "total_boost_count":

		switch b.badge.Trigger {
		case TriggerMoreThan:

			var count uint64

			qry := fmt.Sprintf("SELECT COUNT(*) FROM boosts WHERE user_id = %d", b.userID)
			err := b.db.QueryRow(qry).Scan(&count)

			if err != nil {
				log.Printf("Query: %s Error: %s", qry, err)
				return nil, err
			}

			if count >= b.badge.Value {
				var achievement Achievement

				achievement.BadgeID = b.badge.ID
				achievement.UserID = b.userID

				if err := achievement.Create(b.db); err != nil {
					return nil, err
				}

				return &achievement, nil
			}

		case TriggerWithInOneDay:
		}

	}

	return &Achievement{}, nil
}
