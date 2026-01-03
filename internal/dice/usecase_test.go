package dice

import (
	"testing"
)

func TestRandom(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
		wantErr bool
	}{
		{
			name:    "正常なレンジ",
			min:     1,
			max:     6,
			wantErr: false,
		},
		{
			name:    "minとmaxが同じ",
			min:     5,
			max:     5,
			wantErr: false,
		},
		{
			name:    "maxがminより小さい場合はエラー",
			min:     10,
			max:     5,
			wantErr: true,
		},
		{
			name:    "負の数を含むレンジ",
			min:     -5,
			max:     5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := random(tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("random() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.min == tt.max {
					// min と max が同じ場合、結果も同じであるべき
					if got != tt.min {
						t.Errorf("random() = %v, want %v", got, tt.min)
					}
				} else {
					// レンジ内に収まっているか確認
					if got < tt.min || got > tt.max {
						t.Errorf("random() = %v, want between %v and %v", got, tt.min, tt.max)
					}
				}
			}
		})
	}
}

func TestDiceUsecase_Dice(t *testing.T) {
	uc := NewDiceUsecase()

	tests := []struct {
		name      string
		diceCount int
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "1個のダイスを振る",
			diceCount: 1,
			wantLen:   1,
			wantErr:   false,
		},
		{
			name:      "複数のダイスを振る",
			diceCount: 5,
			wantLen:   5,
			wantErr:   false,
		},
		{
			name:      "0個の場合はエラー",
			diceCount: 0,
			wantLen:   0,
			wantErr:   true,
		},
		{
			name:      "負の数の場合はエラー",
			diceCount: -1,
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uc.Dice(tt.diceCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != tt.wantLen {
					t.Errorf("Dice() length = %v, want %v", len(got), tt.wantLen)
				}
				// 各結果が有効なダイス絵文字であることを確認
				for i, result := range got {
					valid := false
					for _, validDice := range dice {
						if result == validDice {
							valid = true
							break
						}
					}
					if !valid {
						t.Errorf("Dice()[%d] = %v, want valid dice emoji", i, result)
					}
				}
			}
		})
	}
}
