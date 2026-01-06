package audio

import "testing"

func TestGetMP3DurationMs(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int
		want     int
	}{
		{
			name:     "1秒相当のデータ（128kbps = 16000バイト/秒）",
			dataSize: 16000,
			want:     1000,
		},
		{
			name:     "5秒相当のデータ",
			dataSize: 80000,
			want:     5000,
		},
		{
			name:     "空データ",
			dataSize: 0,
			want:     0,
		},
		{
			name:     "10秒相当のデータ",
			dataSize: 160000,
			want:     10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataSize)
			got := GetMP3DurationMs(data)
			if got != tt.want {
				t.Errorf("GetMP3DurationMs() = %v, want %v", got, tt.want)
			}
		})
	}
}
