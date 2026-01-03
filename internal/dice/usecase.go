package dice

import (
	"fmt"
	"math/rand"
	"time"
)

var dice = [6]string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣"}

type DiceUsecase struct{}

func NewDiceUsecase() *DiceUsecase {
	return &DiceUsecase{}
}

func (uc *DiceUsecase) Dice(diceCount int) ([]string, error) {
	if diceCount < 1 {
		return nil, fmt.Errorf("エラー")
	}

	results := make([]string, 0)
	for range diceCount {
		r, _ := random(1, 6)
		results = append(results, dice[r-1])
	}
	return results, nil
}

func random(min int, max int) (int, error) {
	if max < min {
		return 0, fmt.Errorf("max %d must be >= min %d", max, min)
	}
	if max == min {
		return min, nil
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min, nil
}
