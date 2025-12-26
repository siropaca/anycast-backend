package di

import (
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/handler"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// Container は DI コンテナ
type Container struct {
	VoiceHandler *handler.VoiceHandler
}

// NewContainer は依存関係を構築して Container を返す
func NewContainer(db *gorm.DB) *Container {
	// Repository 層
	voiceRepo := repository.NewVoiceRepository(db)

	// Service 層
	voiceService := service.NewVoiceService(voiceRepo)

	// Handler 層
	voiceHandler := handler.NewVoiceHandler(voiceService)

	return &Container{
		VoiceHandler: voiceHandler,
	}
}
