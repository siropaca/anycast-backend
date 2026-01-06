package storage

import "testing"

func TestGenerateAudioPath(t *testing.T) {
	client := &gcsClient{}

	tests := []struct {
		name    string
		audioID string
		want    string
	}{
		{
			name:    "UUID 形式の ID",
			audioID: "550e8400-e29b-41d4-a716-446655440000",
			want:    "audios/550e8400-e29b-41d4-a716-446655440000.mp3",
		},
		{
			name:    "短い ID",
			audioID: "abc123",
			want:    "audios/abc123.mp3",
		},
		{
			name:    "空の ID",
			audioID: "",
			want:    "audios/.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.GenerateAudioPath(tt.audioID)
			if got != tt.want {
				t.Errorf("GenerateAudioPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
