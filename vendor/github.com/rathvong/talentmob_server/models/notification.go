package models

import (
	"github.com/rathvong/talentmob_server/system"
	"database/sql"
	"os"
	"log"
	"time"
	"github.com/NaySoftware/go-fcm"


)

const (
	OBJECT_USER        = "user"
	OBJECT_VIDEO       = "video"
	OBJECT_COMMENT     = "comment"
	OBJECT_EVENT       = "event"
	OBJECT_COMPETITION = "competition"
	VERB_FAVOURITED    = "favourited"
	VERB_IMPORTED      = "imported"
	VERB_VIEWED        = "viewed"
	VERB_WON           = "won"
	VERB_COMMENTED     = "commented"
	VERB_JOINED        = "joined"
	VERB_VOTING_BEGAN  = "voting_began"
	VERB_UPVOTED       = "upvoted"
	VERB_FOLLOWED      = "followed"
	VERB_VOTING_ENDED  = "voting_ended"
	PUSHSERVER_GOOGLE  = "google"
	PUSHSEVER_APPLE    = "apple"
)


//Server key to perform all push notifications
var(
	FCMServerKey = os.Getenv("FCM_SERVER_KEY")
	Object = []string{OBJECT_COMMENT,OBJECT_VIDEO, OBJECT_USER,OBJECT_EVENT,OBJECT_COMPETITION}

	Verb = []string{VERB_FAVOURITED, VERB_COMMENTED, VERB_FOLLOWED, VERB_IMPORTED, VERB_JOINED, VERB_VOTING_BEGAN, VERB_UPVOTED,VERB_VIEWED, VERB_WON, VERB_VOTING_ENDED}
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
func (n * Notification) SendPushNotification(db *system.DB) (err error) {
	log.Println("Notification.SendPushNotification()")

	if err = n.validateErrors(); err != nil {
		log.Println("Notification.SendPushNotification() Error -> ", err)
		return
	}

	sender := User{}

	receiver := User{}

	if err = sender.Get(db, n.SenderID); err != nil{
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
			log.Println("Notification.SendPushNotification() apple service" )

		default:
			log.Printf("Notification.SendPushNotification() unknown service  -> %v", api )
		}
	}

	return
}

func (n *Notification) GetObject(db *system.DB) (object interface{}, err error){

	switch n.ObjectType {
	case OBJECT_VIDEO:
		video := Video{}

		if err = video.GetVideoByID(db, n.ObjectID); err != nil {
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
	}

	return
}

func (n *Notification) buildBodyText(sender User, receiver User, object interface{}, db *system.DB) (body string){
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

	case VERB_JOINED:
		body += " has joined"
	case VERB_IMPORTED:
		body += " has imported"
	case VERB_FOLLOWED:
		body += " has followed you"
	}

	switch n.ObjectType {
	case OBJECT_EVENT:
		body += " the event for week: "

		//todo:: complete notification for competition
	case OBJECT_COMPETITION:
		switch  n.Verb {
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
		case VERB_VIEWED, VERB_UPVOTED, VERB_FAVOURITED:
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
func (n *Notification) SendFCMPushToClient(pushToken string, msg AlertMessage) (err error){
	client := fcm.NewFcmClient(FCMServerKey)

	client.NewFcmMsgTo(pushToken, msg)

	status, err := client.Send()

	if  err == nil {
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
	SenderID uint64 `json:"sender_id"`
	ReceiverID uint64 `json:"receiver_id"`
	ObjectID uint64 `json:"object_id"`
	ObjectType string `json:"object_type"`
	Verb string `json:"verb"`
	IsRead bool `json:"is_read"`
	IsActive bool `json:"is_active"`

}


func (n *Notification) isValidObject(object string) (valid bool){

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



func (n *Notification) queryCreate() (qry string){
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

func (n *Notification) queryUpdate() (qry string){
	return `UPDATE notifications SET
					object_id = $2,
					object_type = $3,
					verb = $4,
					is_read = $5,
					is_active = $6,
					updated_at = $7
			WHERE id = $1`
}


func (n *Notification) queryGet() (qry string){
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

func (n *Notification) queryGetUnread() (qry string){
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

func (n *Notification) queryGetUnreadCount() (qry string){
	return `SELECT COUNT(*) FROM notifications WHERE receiver_id = $1 AND is_read = false`
}


func (n *Notification) validateErrors() (err error) {

	if !n.isValidObject(n.ObjectType){
		return n.Errors(ErrorIncorrectValue, "object_type")
	}

	if !n.isValidVerb(n.Verb){
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

func (n *Notification) Create(db *system.DB)(err error){

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
		log.Printf("Notification.create() QueryRow() -> %v Error -> %v", n.queryCreate(), err )
		return
	}

	return
}

func (n *Notification) Update(db *system.DB) (err error){

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
		log.Printf("Notification.Update() QueryRow() -> %v Error -> %v", n.queryUpdate(), err )
		return
	}

	return
}

func (n *Notification) Get(db *system.DB, notificationID uint64) (err error){

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

	if err != nil && sql.ErrNoRows != err{
		log.Printf("Notification.Get() NotificationID -> %v QueryRow() -> %v Error -> %v", notificationID, n.queryGet(), err)
		return
	}

	return
}

func (n *Notification) GetUnreadCount(db *system.DB, receiverID uint64) (count int, err error){

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


func (n *Notification) GetUnread(db *system.DB, receiverID uint64) (notifications []Notification, err error){

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

func (n *Notification) parseRows(rows *sql.Rows) (notifications []Notification, err error){

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




//Create and send a push notification to a target user
func Notify(db *system.DB, senderID uint64, receiverID uint64, verb string, objectID uint64, objectType string) (err error){
	notification := Notification{}

	notification.SenderID = senderID
	notification.ReceiverID = receiverID
	notification.Verb = verb
	notification.ObjectID = objectID
	notification.ObjectType = objectType

	return notification.Create(db)
}