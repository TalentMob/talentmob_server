package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rathvong/scheduler"
	"github.com/rathvong/scheduler/storage"

	"github.com/rathvong/talentmob_server/api"
	"github.com/rathvong/talentmob_server/leaderboardpayouts"
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

var eventSchedular *scheduler.Scheduler

var AWS_CONFIG = awsDatabaseURL + "&sslmode=verify-full&sslrootcert=config/rds-combined-ca-bundle.pem"

func main() {

	db := system.Connect(AWS_CONFIG)
	defer db.Close()

	event := make(chan models.Event)
	server := api.Server{Db: db, AddEventChannel: event}

	go startSchedular(db)
	go eventHub(db, &server)

	server.Serve()

}

func startSchedular(db *system.DB) {

	s := scheduler.New(storage.NewNoOpStorage())
	eventSchedular = &s

	s.Start()
	s.Wait()

}

func HandleEventsPayout(db *system.DB, event *models.Event) {

	env := os.Getenv("env")

	log.Println("Starting event payout for ", event.Title)

	rank, _ := leaderboardpayouts.BuildRankingPayout()
	event.PrizeList = rank.GetValuesForEntireRanking(rank.DisplayForRanking(event.PrizePool, int(event.CompetitorsCount)))

	qry := `SELECT
								competitors.id,
								competitors.user_id,
								competitors.up_votes,
								videos.id,
								videos.title,
								videos.thumbnail,
								events.title
						FROM competitors
						INNER JOIN events
						ON events.id = competitors.event_id
						INNER JOIN videos
						ON videos.id = competitors.video_id
						WHERE event_id = $1
						ORDER BY competitors.up_votes DESC, competitors.down_votes ASC
						`

	rows, err := db.Query(qry, event.ID)

	defer rows.Close()

	if err != nil {
		log.Printf("sql: %v, err: ", qry, err)
		return
	}

	var count int

	for rows.Next() {

		var eventRanking models.EventRanking

		err := rows.Scan(
			&eventRanking.CompetitorID,
			&eventRanking.UserID,
			&eventRanking.TotalVotes,
			&eventRanking.VideoID,
			&eventRanking.VideoTitle,
			&eventRanking.VideoThumbnail,
			&eventRanking.EventTitle,
		)

		if err != nil {
			log.Println("Event Payout: ", err)
			return
		}

		// var point models.Point

		// if err := point.GetByUserID(db, eventRanking.UserID); err != nil {
		// 	log.Println(err)
		// 	continue
		// }

		eventRanking.EventID = event.ID
		eventRanking.Ranking = uint(count + 1)

		if count < len(event.PrizeList) {
			eventRanking.PayOut = event.PrizeList[count]
			// point.AddPayout(int64(eventRanking.PayOut))

			// if err := point.Update(db); err != nil {
			// 	log.Println(err)
			// 	continue
			// }

			eventRanking.IsPaid = true
		}

		if env == "production" {
			if err := eventRanking.Create(db); err != nil {
				panic(err)

			}
		}

		//models.Notify(db, 11, eventRanking.UserID, models.VERB_VOTING_ENDED, eventRanking.ID, models.OBJECT_EVENT_RANKING)
		count++

		log.Printf("%+v", eventRanking)

	}

	event.IsOpened = false

	if env == "production" {
		if err := event.Update(db); err != nil {
			panic(err)
			return
		}
	}

}

func eventHub(db *system.DB, server *api.Server) {
	for {
		select {
		case event := <-server.AddEventChannel:

			id, err := eventSchedular.RunAt(event.StartDate.Add(368), fmt.Sprintf("%d", event.ID), HandleEventsPayout, db, &event)

			if err != nil {
				log.Printf("eventHub() Error: %v", err)
			}

			log.Printf("eventHub() ID: %s, Event added: %+v", id, event)
		}
	}
}
