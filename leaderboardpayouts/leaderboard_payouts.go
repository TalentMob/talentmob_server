package leaderboardpayouts

import (
	"github.com/tealeg/xlsx"
	"github.com/rathvong/util"
	"log"
	"sort"
)

type Entrants map[int]float32



type Rank struct {
	Data         map[int]Entrants
	rankKeys     []int
	entranceKeys []int
}



func buildColumnsTitle(xlFile *xlsx.File) (columns []int){
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
							columns = append(columns, key)
						}
					}


				}
			}



		}
	}


	return
}


func buildRowsTitle(xlFile *xlsx.File) (rows []int) {
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
						rows = append(rows, key)
					}



				}
			}



		}
	}

	return rows
}



func BuildRankingPayout() (rankings Rank, err error) {

	excelFileName := "config/leaderboardpayouts.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)

	if err != nil {
		return
	}



	rankings.rankKeys = buildRowsTitle(xlFile)
	rankings.entranceKeys = buildColumnsTitle(xlFile)

	rankings.Data = make(map[int]Entrants, len(rankings.rankKeys))




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
			entrants = make(map[int]float32, len(rankings.entranceKeys))

			for k, cell := range row.Cells {

				if k < 1 {
					continue
				}

				text := cell.String()

				if len(text) > 0 {

						key, _ := util.ConvertStringToFloat64(text)

						if key > 0 {

							entrants[rankings.entranceKeys[k-1]] = float32(key)


						} else {
							  break
						}
					}


			}

			rankings.Data[rankings.rankKeys[j-1]] = entrants

		}
	}

	sort.Ints(rankings.entranceKeys)
	sort.Ints(rankings.rankKeys)

	return
}

func (r *Rank) GetPercentage(rv int, ev int) (p float32){

	rk := r.RankKey(rv)
	ek := r.EntranceKey(ev)

	v := r.Data[rk]


	return v[ek]
}

func (r *Rank) RankKey(n int) (key int){
	return getKey(n, r.rankKeys)
}


func (r *Rank) EntranceKey(n int) (key int){
	return getKey(n, r.entranceKeys)
}


func  getKey(n int, list []int) (key int){

	if n == 0 {
		return
	}

	size := len(list)

	for i := 0; i < size - 1; i++ {
		nextKey := list[i + 1]
		currentKey := list[i]
		firstKey := list[0]

		if n < nextKey && n >= currentKey {
			key = currentKey
		} else if n >= list[size - 1] {
			key = nextKey
		} else if n < firstKey {
			key = firstKey
		}
	}
	
	return
}






