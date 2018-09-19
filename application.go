package main

import (
	"os"

	"github.com/rathvong/talentmob_server/api"
	"github.com/rathvong/talentmob_server/models"
	"github.com/rathvong/talentmob_server/system"
)

// Key strings for environment variables
const (
	AWS_ENVIRONMENT_DATABASE_URL    = "DATABASE_AWS"
	HEROKU_ENVIRONMENT_DATABASE_URL = "DATABASE_URL"
)

// Initialized database url set in environment
var (
	//AWS DB URL
	awsDatabaseURL = os.Getenv(AWS_ENVIRONMENT_DATABASE_URL)
)

var AWS_CONFIG = awsDatabaseURL + "&sslmode=verify-full&sslrootcert=config/rds-combined-ca-bundle.pem"

var events []models.Event

func main() {

	db := system.Connect(AWS_CONFIG)
	defer db.Close()

	event := make(chan *models.Event)

	server := api.Server{Db: db, AddEventChannel: event}

	// go func() {
	// 	var e models.Event

	// 	events, err := e.GetAllEventsByRunning(db, true)
	// 	if err != nil {
	// 		panic(err)

	// 	}

	// 	for {
	// 		for i, event := range events {

	// 			if event.StartDate.Add(time.Hour*time.Duration(338)).UnixNano() > time.Now().UnixNano() {

	// 				rank, _ := leaderboardpayouts.BuildRankingPayout()
	// 				event.PrizeList = rank.GetValuesForEntireRanking(rank.DisplayForRanking(event.PrizePool, int(event.CompetitorsCount)))

	// 				qry := `SELECT
	// 								competitors.id,
	// 								competitors.user_id,
	// 								competitors.up_votes,
	// 								videos.id,
	// 								videos.title,
	// 								videos.thumbnail,
	// 								events.title
	// 						FROM competitors
	// 						INNER JOIN events
	// 						ON events.id = competitors.event_id
	// 						INNER JOIN videos
	// 						ON videos.id = competitors.video_id
	// 						WHERE event_id = $1
	// 						ORDER BY competitors.up_votes DESC, competitors.down_votes ASC
	// 						`

	// 				rows, err := db.Query(qry, event.ID)

	// 				defer rows.Close()

	// 				if err != nil {
	// 					log.Printf("sql: %v, err: ", qry, err)
	// 					continue
	// 				}

	// 				var count int

	// 				for rows.Next() {

	// 					var eventRanking models.EventRanking

	// 					err := rows.Scan(
	// 						&eventRanking.CompetitorID,
	// 						&eventRanking.UserID,
	// 						&eventRanking.TotalVotes,
	// 						&eventRanking.VideoID,
	// 						&eventRanking.VideoTitle,
	// 						&eventRanking.VideoThumbnail,
	// 						&eventRanking.EventTitle,
	// 					)

	// 					if err != nil {
	// 						log.Println("Event Payout: ", err)
	// 						return
	// 					}

	// 					var point models.Point

	// 					if err := point.GetByUserID(db, eventRanking.UserID); err != nil {
	// 						log.Println(err)
	// 						continue
	// 					}

	// 					eventRanking.Ranking = uint(count + 1)

	// 					if count < len(event.PrizeList) {
	// 						eventRanking.PayOut = event.PrizeList[count]
	// 						point.AddPayout(int64(eventRanking.PayOut))

	// 						if err := point.Update(db); err != nil {
	// 							log.Println(err)
	// 							continue
	// 						}

	// 						eventRanking.IsPaid = true
	// 					}

	// 					if err := eventRanking.Create(db); err != nil {
	// 						log.Println(err)
	// 						continue
	// 					}

	// 					models.Notify(db, 11, eventRanking.UserID, models.VERB_VOTING_ENDED, eventRanking.ID, models.OBJECT_EVENT_RANKING)

	// 				}

	// 				events = append(events[:i], events[i+1:]...)
	// 				i--

	// 			}
	// 		}
	// 	}

	// }()

	// go func() {
	// 	for {
	// 		select {
	// 		case event := <-server.AddEventChannel:
	// 			if !eventContains(*event) {
	// 				events = append(events, *event)
	// 			}
	// 		}
	// 	}
	// }()

	server.Serve()

}

func eventContains(event models.Event) bool {
	for _, e := range events {
		if e.ID == event.ID {
			return true
		}

	}

	return false
}
