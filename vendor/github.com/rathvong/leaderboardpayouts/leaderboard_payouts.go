package leaderboardpayouts

import (
	"github.com/tealeg/xlsx"
	"github.com/rathvong/util"
	"log"
)

type Entrants map[int]float32



var entrantsCategories []int
var rankingCategories []int

func buildEntrantsKeys() (keys []int){
	excelFileName := "config/leaderboardpayouts.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)


	if err != nil {
		panic(err)
	}


	for i, sheet := range xlFile.Sheets {
		if i > 0 {
			break
		}

		for j, row := range sheet.Rows {
			if j > 0 {
				break
			}

				for _, cell := range row.Cells {
					text := cell.String()

					if len(text) > 0 {

						if j == 0 {

							key,_ := util.ConvertStringToInt(text)

							if key > 0 {
								entrantsCategories = append(entrantsCategories, key)
							}
						}


					}
				}



		}
	}


	return entrantsCategories
}

func buildRankingKeys() (keys []int){

	excelFileName := "config/leaderboardpayouts.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)


	if err != nil {
		panic(err)
	}


	for i, sheet := range xlFile.Sheets {
		if i > 0 {
			break
		}

		for _, row := range sheet.Rows {

			for k, cell := range row.Cells {

				if k > 0 {
					break
				}

				text := cell.String()

				if len(text) > 0 {


						key, _ := util.ConvertStringToInt(text)

						if key > 0 {
							rankingCategories = append(rankingCategories, key)
						}



				}
			}



		}
	}

	return rankingCategories
}

func buildRankingPayout() (rankings map[int]Entrants) {

	var entrantsCategories []int
	var rankingCategories []int

	rankingCategories = buildRankingKeys()
	entrantsCategories = buildEntrantsKeys()

	rankings = make(map[int]Entrants, len(rankingCategories))


	excelFileName := "config/leaderboardpayouts.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)

	if err != nil {
		panic(err)
	}

	log.Println("buildRankingPayout()")

	for i, sheet := range xlFile.Sheets {
		if i > 0 {
			break
		}

		for j, row := range sheet.Rows {
			if j < 1 {
				continue
			}

			var entrants Entrants
			entrants = make(map[int]float32, len(entrantsCategories))

			for k, cell := range row.Cells {

				if k < 1 {
					continue
				}

				text := cell.String()

				if len(text) > 0 {

						key, _ := util.ConvertStringToFloat64(text)

						if key > 0 {

							entrants[entrantsCategories[k-1]] = float32(key)


						} else {
							  break
						}
					}


			}

			log.Println(entrants)
			rankings[rankingCategories[j-1]] = entrants

		}
	}

	return rankings
}




