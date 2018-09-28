package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"

	"github.com/NaySoftware/go-fcm"
	"github.com/rathvong/talentmob_server/system"
)

const (
	OBJECT_USER          = "user"
	OBJECT_VIDEO         = "video"
	OBJECT_COMMENT       = "comment"
	OBJECT_EVENT         = "event"
	OBJECT_EVENT_RANKING = "event_ranking"
	OBJECT_COMPETITION   = "competition"
	VERB_FAVOURITED      = "favourited"
	VERB_IMPORTED        = "imported"
	VERB_VIEWED          = "viewed"
	VERB_WON             = "won"
	VERB_COMMENTED       = "commented"
	VERB_JOINED          = "joined"
	VERB_VOTING_BEGAN    = "voting_began"
	VERB_UPVOTED         = "upvoted"
	VERB_FOLLOWED        = "followed"
	VERB_VOTING_ENDED    = "voting_ended"
	VERB_BOOST           = "boost"
	PUSHSERVER_GOOGLE    = "google"
	PUSHSEVER_APPLE      = "apple"
)

//Server key to perform all push notifications
var (
	FCMServerKey = os.Getenv("FCM_SERVER_KEY")
	Object       = []string{OBJECT_COMMENT, OBJECT_VIDEO, OBJECT_USER, OBJECT_EVENT, OBJECT_COMPETITION, OBJECT_EVENT_RANKING}

	Verb = []string{VERB_FAVOURITED, VERB_COMMENTED, VERB_FOLLOWED, VERB_IMPORTED, VERB_JOINED, VERB_VOTING_BEGAN, VERB_UPVOTED, VERB_VIEWED, VERB_WON, VERB_VOTING_ENDED, VERB_BOOST}
)

//Apple push notification format
type AlertDictionary struct {
	Title       string `json:"title"`
	LaunchImage string `json:"launch-image"`
	Body        string `json:"body"`
}

// All notifications will be returned as in
// Apple push notification json format
// This could be parsed in Android and IOS apps.
// Apple allow custom Key/Values and
// would be easily parsed when the notification is received.
type Aps struct {
	Alert                   AlertDictionary `json:"alert"`
	Badge                   int             `json:"badge"`
	Sound                   string          `json:"sound"`
	SenderID                uint64          `json:"sender_id"`
	SenderName              string          `json:"sender_name"`
	Notification            Notification    `json:"notification"`
	Object                  interface{}     `json:"object"`
	UnreadNotificationCount int             `json:"unread_notification_count"`
	UrlImage                string          `json:"url_image"`
}

type AlertMessage struct {
	Aps Aps `json:"aps"`
}

// Build push notification and send out to all
// active mobile devices registered by the user.
func (n *Notification) SendPushNotification(db *system.DB) (err error) {
	log.Println("Notification.SendPushNotification()")

	if err = n.validateErrors(); err != nil {
		log.Println("Notification.SendPushNotification() Error -> ", err)
		return
	}

	sender := User{}

	receiver := User{}

	if err = sender.Get(db, n.SenderID); err != nil {
		log.Println("Notification.SendPushNotification() Could not retrieve sender info.")
		return
	}

	if err = receiver.Get(db, n.ReceiverID); err != nil {
		log.Println("Notification.SendPushNotification() Could not retrieve receiver info.")
		return
	}

	alertMessage := AlertMessage{}

	alertMessage.Aps.Notification = *n
	alertMessage.Aps.Alert.Title = "Talent Mob"
	alertMessage.Aps.SenderID = n.SenderID
	alertMessage.Aps.SenderName = sender.Name
	alertMessage.Aps.UrlImage = sender.Avatar

	if alertMessage.Aps.UnreadNotificationCount, err = n.GetUnreadCount(db, receiver.ID); err != nil {
		log.Println("Notification.SendPushNotification() Could not retrieve unreadcount.", err)
		return
	}

	if alertMessage.Aps.Object, err = n.GetObject(db); err != nil {
		log.Println("Notification.SendPushNotification() Could not retrieve object.", err)

		return
	}

	alertMessage.Aps.Alert.Body = n.buildBodyText(sender, receiver, alertMessage.Aps.Object, db)

	apis, err := receiver.Api.GetAllActiveAPIs(db, receiver.ID)

	if err != nil {
		log.Println("Notification.SendPushNotification() Could not retrieve apis.", err)
		return err
	}

	for _, api := range apis {
		switch api.PushNotificationService {
		case PUSHSERVER_GOOGLE:
			if api.PushNotificationToken != "" {
				if err = n.SendFCMPushToClient(api.PushNotificationToken, alertMessage); err != nil {
					log.Println("Notification.SendPushNotification() SendFCMPushToClient() Error -> ", err)
					continue
				}
			} else {
				log.Println("Notification.SendPushNotification() Error -> Missing push_notification_token for api -> ", api.ID)
			}
		case PUSHSEVER_APPLE:
			log.Println("Notification.SendPushNotification() apple service")

		default:
			log.Printf("Notification.SendPushNotification() unknown service  -> %v", api)
		}
	}

	return
}

func (n *Notification) GetObject(db *system.DB) (object interface{}, err error) {

	switch n.ObjectType {
	case OBJECT_VIDEO:
		video := Video{}

		if err = video.GetVideoByID2(db, n.ObjectID); err != nil {
			panic(err)

		}

		object = video
	case OBJECT_COMMENT:
		comment := Comment{}

		if err = comment.Get(db, n.ObjectID); err != nil {
			panic(err)
		}

		object = comment

	case OBJECT_COMPETITION:
		compete := Competitor{}

		if err = compete.Get(db, n.ObjectID); err != nil {
			panic(err)
		}

		object = compete

	case OBJECT_EVENT:
		event := Event{}

		if err = event.Get(db, n.ObjectID); err != nil {
			panic(err)
		}

		object = event

	case OBJECT_EVENT_RANKING:
		eventRanking := EventRanking{}

		if err = eventRanking.Get(db, n.ObjectID); err != nil {
			panic(err)
		}

		object = eventRanking
	}

	return
}

func (n *Notification) buildBodyText(sender User, receiver User, object interface{}, db *system.DB) (body string) {
	body += sender.Name

	switch n.Verb {
	case VERB_COMMENTED:
		body += " has commented"
	case VERB_UPVOTED:
		body += " has up voted"
	case VERB_WON:
		body = "Congratulations! You have won"
	case VERB_VIEWED:
		body += " has viewed"
	case VERB_VOTING_BEGAN:

	case VERB_VOTING_ENDED:

		body = "Congratulations! You finished #"

	case VERB_JOINED:
		body += " has joined"
	case VERB_IMPORTED:
		body += " has imported"
	case VERB_FOLLOWED:
		body += " has followed you"
	case VERB_BOOST:
		body += " has boosted "
	}

	switch n.ObjectType {
	case OBJECT_EVENT:
		body += " the event for week: "

		//todo:: complete notification for competition
	case OBJECT_EVENT_RANKING:
		eventRanking := object.(EventRanking)

		body += fmt.Sprintf("%d", eventRanking.Ranking) + " in the " + eventRanking.EventTitle + " contest and collected " + fmt.Sprintf("%d", eventRanking.PayOut) + " StarPower."

	case OBJECT_COMPETITION:
		switch n.Verb {
		case VERB_VOTING_BEGAN:
		}
	case OBJECT_COMMENT:
		comment := object.(Comment)

		video, err := comment.GetVideo(db)
		if err != nil {
			panic(err)

		}

		body += " on your video: " + video.Title

	case OBJECT_VIDEO:
		video := object.(Video)

		switch n.Verb {
		case VERB_VIEWED, VERB_UPVOTED, VERB_FAVOURITED, VERB_BOOST:
			body += " your video: " + video.Title
		case VERB_IMPORTED:
			body += " a new video: " + video.Title
		default:
			body += " on your video: " + video.Title
		}

	}

	return
}

// Push notification to FCM servers
func (n *Notification) SendFCMPushToClient(pushToken string, msg AlertMessage) (err error) {
	client := fcm.NewFcmClient(FCMServerKey)

	client.NewFcmMsgTo(pushToken, msg)

	status, err := client.Send()

	if err == nil {
		status.PrintResults()
	} else {
		log.Println("SendFCMPushToClient", err)
	}

	log.Println("Push notification sent to ", n.ReceiverID)

	return

}

//This handles all notifications operations for users
// Users can receive notifcatiosn on all types of events performed
// by other users
type Notification struct {
	BaseModel
	SenderID   uint64      `json:"sender_id"`
	ReceiverID uint64      `json:"receiver_id"`
	Sender     User        `json:"sender"`
	ObjectID   uint64      `json:"object_id"`
	ObjectType string      `json:"object_type"`
	Verb       string      `json:"verb"`
	IsRead     bool        `json:"is_read"`
	IsActive   bool        `json:"is_active"`
	Object     interface{} `json:"object"`
}

func (n *Notification) isValidObject(object string) (valid bool) {

	for _, value := range Object {
		if value == object {
			return true
		}
	}

	return false
}

func (n *Notification) isValidVerb(verb string) (valid bool) {
	for _, value := range Verb {
		if value == verb {
			return true
		}
	}

	return false
}

func (n *Notification) queryCreate() (qry string) {
	return `INSERT INTO notifications
					(sender_id,
					receiver_id,
					object_id,
					object_type,
					verb,
					is_read,
					is_active,
					created_at,
					updated_at)
			VALUES
					($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`
}

func (n *Notification) queryUpdate() (qry string) {
	return `UPDATE notifications SET
					object_id = $2,
					object_type = $3,
					verb = $4,
					is_read = $5,
					is_active = $6,
					updated_at = $7
			WHERE id = $1`
}

func (n *Notification) queryGet() (qry string) {
	return `SELECT
					id,
					sender_id,
					receiver_id,
					object_id,
					object_type,
					verb,
					is_read,
					is_active,
					created_at,
					updated_at
			FROM notifications
			WHERE
					id = $1`
}

func (n *Notification) queryGetUnread() (qry string) {
	return `SELECT
					id,
					sender_id,
					receiver_id,
					object_id,
					object_type,
					verb,
					is_read,
					is_active,
					created_at,
					updated_at
			FROM notifications
			WHERE
					receiver_id = $1
			AND
					is_read = false
			ORDER BY
					created_at DESC
			LIMIT $2
			OFFSET $3`
}

func (n *Notification) queryGetUnreadCount() (qry string) {
	return `SELECT COUNT(*) FROM notifications WHERE receiver_id = $1 AND is_read = false`
}

func (n *Notification) validateErrors() (err error) {

	if !n.isValidObject(n.ObjectType) {
		return n.Errors(ErrorIncorrectValue, "object_type")
	}

	if !n.isValidVerb(n.Verb) {
		return n.Errors(ErrorIncorrectValue, "verb")
	}

	if n.ReceiverID == 0 {
		return n.Errors(ErrorMissingValue, "receiver_id")
	}

	if n.SenderID == 0 {
		return n.Errors(ErrorMissingValue, "sender_id")
	}

	if n.ObjectID == 0 {
		return n.Errors(ErrorMissingValue, "object_id")
	}

	return
}

func (n *Notification) Create(db *system.DB) (err error) {

	if err = n.validateErrors(); err != nil {
		return
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

		go n.SendPushNotification(db)
	}()

	if err != nil {
		log.Println("Notification.Create() Begin() Error -> ", err)
		return
	}

	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	n.IsActive = true

	err = tx.QueryRow(n.queryCreate(),
		n.SenderID,
		n.ReceiverID,
		n.ObjectID,
		n.ObjectType,
		n.Verb,
		n.IsRead,
		n.IsActive,
		n.CreatedAt,
		n.UpdatedAt).Scan(&n.ID)

	if err != nil {
		log.Printf("Notification.create() QueryRow() -> %v Error -> %v", n.queryCreate(), err)
		return
	}

	return
}

func (n *Notification) Update(db *system.DB) (err error) {

	if err = n.validateErrors(); err != nil {
		return
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
		log.Println("Notification.Update() Begin() Error -> ", err)
		return
	}

	n.UpdatedAt = time.Now()

	err = tx.QueryRow(n.queryUpdate(),
		n.ID,
		n.ObjectID,
		n.ObjectType,
		n.Verb,
		n.IsRead,
		n.IsActive,
		n.UpdatedAt).Scan(&n.ID)

	if err != nil {
		log.Printf("Notification.Update() QueryRow() -> %v Error -> %v", n.queryUpdate(), err)
		return
	}

	return
}

func (n *Notification) Get(db *system.DB, notificationID uint64) (err error) {

	if notificationID == 0 {
		return n.Errors(ErrorMissingValue, "notificationID")
	}

	err = db.QueryRow(n.queryGet(), notificationID).Scan(
		&n.ID,
		&n.SenderID,
		&n.ReceiverID,
		&n.ObjectID,
		&n.ObjectType,
		&n.Verb,
		&n.IsRead,
		&n.IsActive,
		&n.CreatedAt,
		&n.UpdatedAt,
	)

	if err != nil && sql.ErrNoRows != err {
		log.Printf("Notification.Get() NotificationID -> %v QueryRow() -> %v Error -> %v", notificationID, n.queryGet(), err)
		return
	}

	return
}

func (n *Notification) GetUnreadCount(db *system.DB, receiverID uint64) (count int, err error) {

	if receiverID == 0 {
		return 0, n.Errors(ErrorMissingValue, "receiverID")
	}

	err = db.QueryRow(n.queryGetUnreadCount(), receiverID).Scan(&count)

	if err != nil {
		log.Printf("Notification.GetUnreadCount() receiverID -> %v Query() -> %v Error -> %v", receiverID, n.queryGetUnreadCount(), err)
		return
	}

	return
}

func (n *Notification) GetUnread(db *system.DB, receiverID uint64) (notifications []Notification, err error) {

	if receiverID == 0 {
		return notifications, n.Errors(ErrorMissingValue, "receiverID")
	}

	rows, err := db.Query(n.queryGetUnread(), receiverID)

	defer rows.Close()

	if err != nil {
		log.Printf("Notification.GetUnread() receiverID -> %v Query() -> %v Error -> %v", receiverID, n.queryGetUnread(), err)
		return
	}

	return n.parseRows(rows)
}

func (n *Notification) parseRows(rows *sql.Rows) (notifications []Notification, err error) {

	for rows.Next() {

		notification := Notification{}

		err = rows.Scan(
			&notification.ID,
			&notification.SenderID,
			&notification.ReceiverID,
			&notification.ObjectID,
			&notification.ObjectType,
			&notification.Verb,
			&notification.IsRead,
			&notification.IsActive,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			log.Println("Notification.parseRows() Scan() Error -> ", err)
			return
		}

		notifications = append(notifications, notification)
	}

	return
}

func (n *Notification) GetNotifications(db *system.DB, userID uint64, page int) ([]Notification, error) {

	if userID == 0 {
		return nil, n.Errors(ErrorMissingValue, "GetNotifications() : user_id")
	}

	sql := `SELECT notifications.id,
				   notifications.sender_id,
				   notifications.receiver_id,
				   notifications.object_id,
				   notifications.object_type,
				   notifications.verb,
				   notifications.is_read,
				   notifications.is_active,
				   notifications.created_at,
				   notifications.updated_at,
		    	   videos.id,
				   videos.user_id,
				   videos.categories,
				   videos.downvotes,
				   videos.upvotes,
				   videos.shares,
				   videos.views,
				   videos.comments,
				   videos.thumbnail,
				   videos.key,
				   videos.title,
			       videos.created_at,
				   videos.updated_at,
				   videos.is_active,
				   videos.upvote_trending_count,
				   vu.id,
				   vu.avatar,
				   vu.name,
				   vu.account_type,
				   vu.created_at,
				   vu.updated_at,
				   boosts.id,
				   boosts.user_id,
				   boosts.video_id,
				   boosts.start_time,
				   boosts.end_time,
				   boosts.is_active,
				   boosts.created_at,
				   boosts.updated_at,
				   competitors.vote_end_date,		
				   comments.id,
				   comments.user_id,
				   comments.video_id,
				   comments.title,
				   comments.content,
				   comments.is_active,
				   comments.created_at,
				   comments.updated_at,
				   cv.id,
				   cv.user_id,
				   cv.categories,
				   cv.downvotes,
				   cv.upvotes,
				   cv.shares,
				   cv.views,
				   cv.comments,
				   cv.thumbnail,
				   cv.key,
				   cv.title,
			       cv.created_at,
				   cv.updated_at,
				   cv.is_active,
				   cv.upvote_trending_count,
				   vu.id,
				   vu.avatar,
				   vu.name,
				   vu.account_type,
				   vu.created_at,
				   vu.updated_at,
				   cvb.id,
				   cvb.user_id,
				   cvb.video_id,
				   cvb.start_time,
				   cvb.end_time,
				   cvb.is_active,
				   cvb.created_at,
				   cvb.updated_at,
				   cvc.vote_end_date,	
				   er.id,
				   er.event_id,
				   er.competitor_id,
				   er.user_id,
				   er.ranking,
				   er.pay_out,
				   er.total_upvotes,
				   er.video_title,
				   er.video_thumbnail,
				   er.is_active,
				   er.created_at,
				   er.updated_at,
				   er.is_paid, 
				   er.video_id,
				   er.event_title,
				   sender.id,
				   sender.avatar,
				   sender.name,
				   sender.account_type,
				   sender.created_at,
				   sender.updated_at	
			FROM notifications
			LEFT JOIN videos
			ON notifications.object_type = 'video'
			AND notifications.object_id = videos.id
			LEFT JOIN users vu
			ON vu.id = videos.user_id
			LEFT JOIN boosts
			ON boosts.video_id = videos.id
			AND boosts.is_active = true
			AND boosts.end_time > now()	
			LEFT JOIN competitors
			ON competitors.video_id = videos.id
			LEFT JOIN comments
			ON notifications.object_type = 'comment'
			AND notifications.object_id = comments.id
			LEFT JOIN videos cv
			ON cv.id = comments.id
			LEFT JOIN users cvu
			ON cvu.id = cv.user_id
			AND cvu.is_active = true
			LEFT JOIN boosts cvb
			ON cvb.video_id = cv.id
			AND cvb.is_active = true
			AND cvb.end_time > now()	
			LEFT JOIN competitors cvc
			ON cvc.video_id = cv.id
			LEFT JOIN event_rankings er
			ON notifications.object_type = 'event_ranking'
			AND er.id = notifications.object_id
			LEFT JOIN users sender
			ON sender.id = notifications.sender_id
			WHERE notifications.is_active = true
			AND notifications.receiver_id = $1
			ORDER BY notifications.created_at DESC
			LIMIT $2
			OFFSET $3
					`

	rows, err := db.Query(sql, userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Println(" Query() -> %v Error() -> %v", sql, err)
		return nil, err
	}

	return n.ParseRowsForFeed(rows)
}

func (n *Notification) ParseRowsForFeed(rows *sql.Rows) ([]Notification, error) {

	var notifications []Notification

	for rows.Next() {
		var notification Notification

		var vID sql.NullInt64
		var vUserID sql.NullInt64
		var vCategories sql.NullString
		var vDownvotes sql.NullInt64
		var vUpvotes sql.NullInt64
		var vShares sql.NullInt64
		var vViews sql.NullInt64
		var vComments sql.NullInt64
		var vThumbnail sql.NullString
		var vKey sql.NullString
		var vTitle sql.NullString
		var vCreatedAt pq.NullTime
		var vUpdatedAt pq.NullTime
		var vIsActive sql.NullBool
		var vUpvoteTrendingCount sql.NullInt64

		var vuUserID sql.NullInt64
		var vuAvatar sql.NullString
		var vuName sql.NullString
		var vuAccountType sql.NullInt64
		var vuCreatedAt pq.NullTime
		var vuUpdatedAt pq.NullTime

		var vBoostID sql.NullInt64
		var vBoostUserId sql.NullInt64
		var vBoostVideoID sql.NullInt64
		var vBoostStartTime pq.NullTime
		var vBoostEndTime pq.NullTime
		var vBoostIsActive sql.NullBool
		var vBoostCreatedAt pq.NullTime
		var vBoostUpdatedAt pq.NullTime
		var vCompetitorsEndDate pq.NullTime

		var cID sql.NullInt64
		var cUserID sql.NullInt64
		var cVideoID sql.NullInt64
		var cTitle sql.NullString
		var cContent sql.NullString
		var cIsActive sql.NullBool
		var cCreatedAt pq.NullTime
		var cUpdatedAt pq.NullTime

		var cvID sql.NullInt64
		var cvUserID sql.NullInt64
		var cvCategories sql.NullString
		var cvDownvotes sql.NullInt64
		var cvUpvotes sql.NullInt64
		var cvShares sql.NullInt64
		var cvViews sql.NullInt64
		var cvComments sql.NullInt64
		var cvThumbnail sql.NullString
		var cvKey sql.NullString
		var cvTitle sql.NullString
		var cvCreatedAt pq.NullTime
		var cvUpdatedAt pq.NullTime
		var cvIsActive sql.NullBool
		var cvUpvoteTrendingCount sql.NullInt64

		var cvuUserID sql.NullInt64
		var cvuAvatar sql.NullString
		var cvuName sql.NullString
		var cvuAccountType sql.NullInt64
		var cvuCreatedAt pq.NullTime
		var cvuUpdatedAt pq.NullTime

		var cvBoostID sql.NullInt64
		var cvBoostUserId sql.NullInt64
		var cvBoostVideoID sql.NullInt64
		var cvBoostStartTime pq.NullTime
		var cvBoostEndTime pq.NullTime
		var cvBoostIsActive sql.NullBool
		var cvBoostCreatedAt pq.NullTime
		var cvBoostUpdatedAt pq.NullTime
		var cvCompetitorsEndDate pq.NullTime

		var erID sql.NullInt64
		var erEventID sql.NullInt64
		var erCompetitorID sql.NullInt64
		var erUserID sql.NullInt64
		var erRanking sql.NullInt64
		var erPayOut sql.NullInt64
		var erTotalUpVotes sql.NullInt64
		var erTitle sql.NullString
		var erThumbnail sql.NullString
		var erIsActive sql.NullBool
		var erCreatedAt pq.NullTime
		var erUpdatedAt pq.NullTime
		var erIsPaid sql.NullBool
		var erVideoID sql.NullInt64
		var erEventTitle sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.SenderID,
			&notification.ReceiverID,
			&notification.ObjectID,
			&notification.ObjectType,
			&notification.Verb,
			&notification.IsRead,
			&notification.IsActive,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&vID,
			&vUserID,
			&vCategories,
			&vDownvotes,
			&vUpvotes,
			&vShares,
			&vViews,
			&vComments,
			&vThumbnail,
			&vKey,
			&vTitle,
			&vCreatedAt,
			&vUpdatedAt,
			&vIsActive,
			&vUpvoteTrendingCount,
			&vuUserID,
			&vuAvatar,
			&vuName,
			&vuAccountType,
			&vuCreatedAt,
			&vuUpdatedAt,
			&vBoostID,
			&vBoostUserId,
			&vBoostVideoID,
			&vBoostStartTime,
			&vBoostEndTime,
			&vBoostIsActive,
			&vBoostCreatedAt,
			&vBoostUpdatedAt,
			&vCompetitorsEndDate,
			&cID,
			&cUserID,
			&cVideoID,
			&cTitle,
			&cContent,
			&cIsActive,
			&cCreatedAt,
			&cUpdatedAt,
			&cvID,
			&cvUserID,
			&cvCategories,
			&cvDownvotes,
			&cvUpvotes,
			&cvShares,
			&cvViews,
			&cvComments,
			&cvThumbnail,
			&cvKey,
			&cvTitle,
			&cvCreatedAt,
			&cvUpdatedAt,
			&cvIsActive,
			&cvUpvoteTrendingCount,
			&cvuUserID,
			&cvuAvatar,
			&cvuName,
			&cvuAccountType,
			&cvuCreatedAt,
			&cvuUpdatedAt,
			&cvBoostID,
			&cvBoostUserId,
			&cvBoostVideoID,
			&cvBoostStartTime,
			&cvBoostEndTime,
			&cvBoostIsActive,
			&cvBoostCreatedAt,
			&cvBoostUpdatedAt,
			&cvCompetitorsEndDate,
			&erID,
			&erEventID,
			&erCompetitorID,
			&erUserID,
			&erRanking,
			&erPayOut,
			&erTotalUpVotes,
			&erTitle,
			&erThumbnail,
			&erIsActive,
			&erCreatedAt,
			&erUpdatedAt,
			&erIsPaid,
			&erVideoID,
			&erEventTitle,
			&notification.Sender.ID,
			&notification.Sender.Avatar,
			&notification.Sender.Name,
			&notification.Sender.AccountType,
			&notification.Sender.CreatedAt,
			&notification.Sender.UpdatedAt,
		)

		if err != nil {
			log.Println("Notification.ParseRowsForFeed() Error: ", err)
			return nil, err
		}

		switch notification.ObjectType {
		case "comment":

			var c Comment
			if cID.Valid {
				c.ID = uint64(cID.Int64)
			}

			if cUserID.Valid {
				c.UserID = uint64(cUserID.Int64)
			}

			if cVideoID.Valid {
				c.VideoID = uint64(cVideoID.Int64)
			}

			if cTitle.Valid {
				c.Title = cTitle.String
			}

			if cContent.Valid {
				c.Content = cContent.String
			}

			if cIsActive.Valid {
				c.IsActive = cIsActive.Bool
			}

			if cCreatedAt.Valid {
				c.CreatedAt = cCreatedAt.Time
			}

			if cUpdatedAt.Valid {
				c.UpdatedAt = cUpdatedAt.Time
			}

			var v Video
			if cvID.Valid {
				v.ID = uint64(cvID.Int64)
			}

			if cvUserID.Valid {
				v.UserID = uint64(cvUserID.Int64)
			}

			if cvCategories.Valid {
				v.Categories = cvCategories.String
			}

			if cvDownvotes.Valid {
				v.Downvotes = uint64(cvDownvotes.Int64)
			}

			if cvUpvotes.Valid {
				v.Upvotes = uint64(cvUpvotes.Int64)
			}

			if cvShares.Valid {
				v.Shares = uint64(cvShares.Int64)
			}

			if cvViews.Valid {
				v.Views = uint64(cvViews.Int64)
			}

			if cvComments.Valid {
				v.Comments = uint64(cvComments.Int64)
			}

			if cvThumbnail.Valid {
				v.Thumbnail = cvThumbnail.String
			}

			if cvKey.Valid {
				v.Key = cvKey.String
			}

			if cvTitle.Valid {
				v.Title = cvTitle.String
			}

			if cvCreatedAt.Valid {
				v.CreatedAt = cvCreatedAt.Time
			}

			if cvUpdatedAt.Valid {
				v.UpdatedAt = cvUpdatedAt.Time
			}

			if cvIsActive.Valid {
				v.IsActive = cvIsActive.Bool
			}

			if cvUpvoteTrendingCount.Valid {
				v.UpVoteTrendingCount = uint(cvUpvoteTrendingCount.Int64)
			}

			if cvuUserID.Valid {
				v.Publisher.ID = uint64(cvuUserID.Int64)
			}

			if cvuAvatar.Valid {
				v.Publisher.Avatar = cvuAvatar.String
			}

			if cvuName.Valid {
				v.Publisher.Name = cvuName.String
			}

			if cvuAccountType.Valid {
				v.Publisher.AccountType = int(cvuAccountType.Int64)
			}

			if cvuCreatedAt.Valid {
				v.Publisher.CreatedAt = cvuCreatedAt.Time
			}

			if cvuUpdatedAt.Valid {
				v.Publisher.UpdatedAt = cvuUpdatedAt.Time
			}

			if cvBoostID.Valid {
				v.Boost.ID = uint64(cvBoostID.Int64)
			}

			if cvBoostUserId.Valid {
				v.Boost.UserID = uint64(cvBoostUserId.Int64)
			}

			if cvBoostVideoID.Valid {
				v.Boost.VideoID = uint64(cvBoostVideoID.Int64)
			}

			if cvBoostStartTime.Valid {
				v.Boost.StartTime = cvBoostStartTime.Time
			}

			if cvBoostEndTime.Valid {
				v.Boost.EndTime = cvBoostEndTime.Time
				v.Boost.EndTimeUnix = v.Boost.EndTime.UnixNano() / 1000000
			}

			if cvBoostIsActive.Valid {
				v.Boost.IsActive = cvBoostIsActive.Bool
			}

			if cvBoostCreatedAt.Valid {
				v.Boost.CreatedAt = cvBoostCreatedAt.Time
			}

			if cvBoostUpdatedAt.Valid {
				v.Boost.UpdatedAt = cvBoostUpdatedAt.Time
			}

			if cvCompetitorsEndDate.Valid {
				v.CompetitionEndDate = cvCompetitorsEndDate.Time.UnixNano() / 1000000
			}

			c.Object = v
			c.ObjectType = "video"

			notification.Object = c

		case "video":
			var v Video
			if vID.Valid {
				v.ID = uint64(vID.Int64)
			}

			if vUserID.Valid {
				v.UserID = uint64(vUserID.Int64)
			}

			if vCategories.Valid {
				v.Categories = vCategories.String
			}

			if vDownvotes.Valid {
				v.Downvotes = uint64(vDownvotes.Int64)
			}

			if vUpvotes.Valid {
				v.Upvotes = uint64(vUpvotes.Int64)
			}

			if vShares.Valid {
				v.Shares = uint64(vShares.Int64)
			}

			if vViews.Valid {
				v.Views = uint64(vViews.Int64)
			}

			if vComments.Valid {
				v.Comments = uint64(vComments.Int64)
			}

			if vThumbnail.Valid {
				v.Thumbnail = vThumbnail.String
			}

			if vKey.Valid {
				v.Key = vKey.String
			}

			if vTitle.Valid {
				v.Title = vTitle.String
			}

			if vCreatedAt.Valid {
				v.CreatedAt = vCreatedAt.Time
			}

			if vUpdatedAt.Valid {
				v.UpdatedAt = vUpdatedAt.Time
			}

			if vIsActive.Valid {
				v.IsActive = vIsActive.Bool
			}

			if vUpvoteTrendingCount.Valid {
				v.UpVoteTrendingCount = uint(vUpvoteTrendingCount.Int64)
			}

			if vuUserID.Valid {
				v.Publisher.ID = uint64(vuUserID.Int64)
			}

			if vuAvatar.Valid {
				v.Publisher.Avatar = vuAvatar.String
			}

			if vuName.Valid {
				v.Publisher.Name = vuName.String
			}

			if vuAccountType.Valid {
				v.Publisher.AccountType = int(vuAccountType.Int64)
			}

			if vuCreatedAt.Valid {
				v.Publisher.CreatedAt = vuCreatedAt.Time
			}

			if vuUpdatedAt.Valid {
				v.Publisher.UpdatedAt = vuUpdatedAt.Time
			}

			if vBoostID.Valid {
				v.Boost.ID = uint64(vBoostID.Int64)
			}

			if vBoostUserId.Valid {
				v.Boost.UserID = uint64(vBoostUserId.Int64)
			}

			if vBoostVideoID.Valid {
				v.Boost.VideoID = uint64(vBoostVideoID.Int64)
			}

			if vBoostStartTime.Valid {
				v.Boost.StartTime = vBoostStartTime.Time
			}

			if vBoostEndTime.Valid {
				v.Boost.EndTime = vBoostEndTime.Time
				v.Boost.EndTimeUnix = v.Boost.EndTime.UnixNano() / 1000000
			}

			if vBoostIsActive.Valid {
				v.Boost.IsActive = vBoostIsActive.Bool
			}

			if vBoostCreatedAt.Valid {
				v.Boost.CreatedAt = vBoostCreatedAt.Time
			}

			if vBoostUpdatedAt.Valid {
				v.Boost.UpdatedAt = vBoostUpdatedAt.Time
			}

			if vCompetitorsEndDate.Valid {
				v.CompetitionEndDate = vCompetitorsEndDate.Time.UnixNano() / 1000000
			}

			notification.Object = v

		case "event_ranking":

			var e EventRanking
			if erID.Valid {
				e.ID = uint64(erID.Int64)
			}

			if erEventID.Valid {
				e.EventID = uint64(erEventID.Int64)
			}

			if erCompetitorID.Valid {
				e.CompetitorID = uint64(erCompetitorID.Int64)
			}

			if erUserID.Valid {
				e.UserID = uint64(erUserID.Int64)

			}

			if erRanking.Valid {
				e.Ranking = uint(erRanking.Int64)
			}

			if erPayOut.Valid {
				e.PayOut = uint(erPayOut.Int64)
			}

			if erTotalUpVotes.Valid {
				e.TotalVotes = uint(erTotalUpVotes.Int64)
			}

			if erTitle.Valid {
				e.VideoTitle = erTitle.String
			}

			if erThumbnail.Valid {
				e.VideoThumbnail = erThumbnail.String
			}

			if erIsActive.Valid {
				e.IsActive = erIsActive.Bool
			}

			if erCreatedAt.Valid {
				e.CreatedAt = erCreatedAt.Time
			}

			if erUpdatedAt.Valid {
				e.UpdatedAt = erUpdatedAt.Time
			}

			if erIsPaid.Valid {
				e.IsPaid = erIsPaid.Bool
			}

			if erVideoID.Valid {
				e.VideoID = uint64(erVideoID.Int64)
			}

			if erEventTitle.Valid {
				e.EventTitle = erEventTitle.String
			}

			notification.Object = e

		}

		notifications = append(notifications, notification)

	}

	return notifications, nil
}

//Create and send a push notification to a target user
func Notify(db *system.DB, senderID uint64, receiverID uint64, verb string, objectID uint64, objectType string) (err error) {
	notification := Notification{}

	notification.SenderID = senderID
	notification.ReceiverID = receiverID
	notification.Verb = verb
	notification.ObjectID = objectID
	notification.ObjectType = objectType

	return notification.Create(db)
}
