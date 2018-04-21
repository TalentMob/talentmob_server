package leaderboardpayouts

import (
	"testing"
	"log"
	"fmt"
)


func TestEvent_BuildRankingPayout(t *testing.T){
	log.Println(BuildRankingPayout())
}

func TestEvent_TestRankingValue(t *testing.T){
	ranking, err := BuildRankingPayout()

	if err != nil {
		t.Error( err)
		return
	}

	values := ranking.Data[1]

	log.Println("Values for 1 first place -> ",values)
	log.Println("Values for first place with 2 entrants -> ", values[2])
	log.Println("Values for first place with 1001 entrants -> ", values[1001])
}

func TestEvent_TestGetEntrantsKey(t *testing.T){
	log.Println("Testing Entrance Keys...")
	rank, _ := BuildRankingPayout()

	list := []int{100, 10000, 3, 25, 2, 4, 1}
	answer := []int{76, 3001, 3, 21, 2, 3, 2}

	for i, n := range list {
		k := rank.EntranceKey(n)

		log.Printf("entrance: %v column: %v response: %v", n, answer[i], k)

		if k != answer[i] {
			t.Errorf("%v entrances does not fit into %v column -> response answer was -> %v", n, answer[i], k)
		}

	}
}


func TestEvent_TestGetRankingsKey(t *testing.T){
	log.Println("Testing Ranking Keys...")
	rank, _ := BuildRankingPayout()

	list := []int{1,2,3,700,600,300, 325, 14, 15, 55, 0}
	answer := []int{1,2,3,700,600,300, 300, 10, 15, 50, 0}

	for i, n := range list {
		k := rank.RankKey(n)
		log.Printf("ranking: %v column: %v response: %v", n, answer[i], k)

		if k != answer[i] {
			t.Errorf("%v ranking does not fit into %v column -> response answer was -> %v", n, answer[i], k)
		}

	}


}

func TestEvent_TestEntranceColumn(t *testing.T){
	log.Println("Testing Entrance Column...")
	rank, err := BuildRankingPayout()

	answer := "[2800 1700 840 750 650 550 430 290 260 180 180 130]"

	if err != nil {
		t.Error(err)
		return
	}

	response := fmt.Sprint(rank.GetEntranceColumn(100))

	if response != answer {
		t.Errorf("%v does not match %v", response, answer)
	}

}

func TestRank_DisplayForRanking(t *testing.T) {
	log.Println("Testing DisplayForRanking...")
	rank, _ := BuildRankingPayout()

	log.Println(rank.DisplayForRanking(30000, 1000))
}

func TestEvent_TestPayOut(t *testing.T){
	log.Println("Testing Payouts...")
	rank, err := BuildRankingPayout()

	if err != nil {
		t.Error(err)
		return
	}

	ranks := []int{1,2,3,700,600,300, 325, 14, 15, 55, 0}
	entrances := []int{3000, 11, 3, 5, 10, 10000, 2000, 1000, 2500, 1500, 2}

	answer := []int{1600, 2500, 0, 0, 0, 5, 6, 110, 80, 18, 0}

	for i, n := range ranks {

		p := rank.GetPercentage(n, entrances[i])

		a := answer[i]

		log.Printf("rank: %v entrances: %v Answer: %v", n, entrances[i],  a)


		if p != a {
			t.Errorf("percentage returned is %v, should be %v", p, a)
		}

	}




}

func TestRank_GetValuesForEntireRanking(t *testing.T) {

	log.Println("DisplayRankWithKeyToString...")

	rank, err := BuildRankingPayout()

	if err != nil {
		t.Error(err)
		return
	}

	log.Println(rank.GetValuesForEntireRanking(rank.DisplayForRanking(30000, 100)))

}

