package leaderboardpayouts

import (
	"github.com/tealeg/xlsx"
	"github.com/rathvong/util"
	"log"
	"sort"
	"fmt"
)

type Entrants map[int]int



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
			entrants = make(map[int]int, len(rankings.entranceKeys))

			for k, cell := range row.Cells {

				if k < 1 {
					continue
				}

				text := cell.String()

				if len(text) > 0 {

						key, _ := util.ConvertStringToInt(text)

						if key > 0 {

							entrants[rankings.entranceKeys[k-1]] = key


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

func (r *Rank) GetPercentage(rv int, ev int) (p int){

	rk := r.RankKey(rv)
	ek := r.EntranceKey(ev)

	v := r.Data[rk]


	return v[ek]
}

func (r *Rank) GetEntranceColumn(e int) (column []int) {
	for _, rk := range r.rankKeys {
		v := r.Data[rk]

		p := v[r.EntranceKey(e)]

		if p > 0 {
			column = append(column, p)
		}

	}
	return
}

func (r *Rank) DisplayForRanking(points uint64, e int) (column []uint){

	entranceColumn := r.GetEntranceColumn(e)

	for _, v := range entranceColumn {
		if v > 0 {
			starPower := (uint(points) * uint(v)) / 10000

			column = append(column, starPower)
		}
	}

	return
}

func (r *Rank) GetValuesForEntireRanking(column []uint) (value []uint){
	size := len(column)

	for i := 0; i < size; i++ {
		rk := r.rankKeys[i]

		if i > 0 && rk > r.rankKeys[i - 1] {
			rok := rk - r.rankKeys[i - 1]
			lowerKey := r.rankKeys[i - 1] + 1
			if rok > 1 {
				for j:= 0; j < rok; j++ {

					value = append(value, column[i])
					lowerKey++
				}


			} else {
				value = append(value, column[i])

			}

		} else {
			value = append(value, column[i])

		}


	}

	return

}

func (r *Rank) DisplayRankWithKeyToString(column []uint) (s string) {

	size := len(column)

	for i := 0; i < size; i++ {
		rk := r.rankKeys[i]

		if i > 0 && rk > r.rankKeys[i - 1] {
			rok := rk - r.rankKeys[i - 1]
			lowerKey := r.rankKeys[i - 1] + 1
			if rok > 1 {
				for j:= 0; j < rok; j++ {

					s += fmt.Sprintf(" %v. %v", lowerKey, column[i])
					lowerKey++
				}


			} else {
				s += fmt.Sprintf(" %v. %v", rk, column[i])
			}

		} else {
			s += fmt.Sprintf(" %v. %v ", rk, column[i])
		}


	}


	return
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






