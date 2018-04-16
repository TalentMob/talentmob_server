package leaderboardpayouts

import (
	"testing"
	"log"
)

func TestEvent_BuildEntrantsCategories(t *testing.T){

	log.Println(buildEntrantsKeys())
}

func TestEvent_BuildRankingCategories(t *testing.T){
	log.Println(buildRankingKeys())
}

func TestEvent_BuildRankingPayout(t *testing.T){
	log.Println(buildRankingPayout())
}

func TestEvent_TestRankingValue(t *testing.T){
	ranking := buildRankingPayout()

	values := ranking[1]

	log.Println("Values for 1 first place -> ",values)
	log.Println("Values for first place with 2 entrants -> ", values[2])
	log.Println("Values for first place with 1001 entrants -> ", values[1001])
}