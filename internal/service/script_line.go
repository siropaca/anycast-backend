package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 台本行関連のビジネスロジックインターフェース
type ScriptLineService interface {
	ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error)
	Create(ctx context.Context, userID, channelID, episodeID string, req request.CreateScriptLineRequest) (*response.ScriptLineResponse, error)
	Update(ctx context.Context, userID, channelID, episodeID, lineID string, req request.UpdateScriptLineRequest) (*response.ScriptLineResponse, error)
	Delete(ctx context.Context, userID, channelID, episodeID, lineID string) error
	Reorder(ctx context.Context, userID, channelID, episodeID string, req request.ReorderScriptLinesRequest) (*response.ScriptLineListResponse, error)
}

type scriptLineService struct {
	db             *gorm.DB
	scriptLineRepo repository.ScriptLineRepository
	episodeRepo    repository.EpisodeRepository
	channelRepo    repository.ChannelRepository
}

// ScriptLineService の実装を返す
func NewScriptLineService(
	db *gorm.DB,
	scriptLineRepo repository.ScriptLineRepository,
	episodeRepo repository.EpisodeRepository,
	channelRepo repository.ChannelRepository,
) ScriptLineService {
	return &scriptLineService{
		db:             db,
		scriptLineRepo: scriptLineRepo,
		episodeRepo:    episodeRepo,
		channelRepo:    channelRepo,
	}
}

// 指定されたエピソードの台本行一覧を取得する
func (s *scriptLineService) ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 台本行一覧を取得
	scriptLines, err := s.scriptLineRepo.FindByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}

	// レスポンスに変換
	responses := s.toScriptLineResponses(scriptLines)

	return &response.ScriptLineListResponse{
		Data: responses,
	}, nil
}

// 新しい台本行を作成する
func (s *scriptLineService) Create(ctx context.Context, userID, channelID, episodeID string, req request.CreateScriptLineRequest) (*response.ScriptLineResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	speakerID, err := uuid.Parse(req.SpeakerID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("speakerId の形式が無効です")
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// speakerId がチャンネルに紐づいているキャラクターか確認
	var validSpeaker bool
	for _, cc := range channel.ChannelCharacters {
		if cc.CharacterID == speakerID {
			validSpeaker = true
			break
		}
	}
	if !validSpeaker {
		return nil, apperror.ErrValidation.WithMessage("指定された話者はこのチャンネルに紐づいていません")
	}

	// 挿入位置（lineOrder）を決定
	var newLineOrder int
	if req.AfterLineID != nil {
		afterLineID, err := uuid.Parse(*req.AfterLineID)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("afterLineId の形式が無効です")
		}

		// afterLineId で指定された行を取得
		afterLine, err := s.scriptLineRepo.FindByID(ctx, afterLineID)
		if err != nil {
			return nil, err
		}

		// 指定された行がこのエピソードに属しているか確認
		if afterLine.EpisodeID != eid {
			return nil, apperror.ErrNotFound.WithMessage("指定された afterLineId がこのエピソードに見つかりません")
		}

		newLineOrder = afterLine.LineOrder + 1
	} else {
		// afterLineId が指定されていない場合は先頭（lineOrder = 0）に挿入
		newLineOrder = 0
	}

	// 新しい台本行を作成
	scriptLine := &model.ScriptLine{
		EpisodeID: eid,
		LineOrder: newLineOrder,
		SpeakerID: speakerID,
		Text:      req.Text,
		Emotion:   req.Emotion,
	}

	// トランザクションで lineOrder のインクリメントと作成を実行
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txScriptLineRepo := repository.NewScriptLineRepository(tx)

		// 挿入位置以降の行の lineOrder を +1
		if err := txScriptLineRepo.IncrementLineOrderFrom(ctx, eid, newLineOrder); err != nil {
			return err
		}

		// 新しい行を作成
		if err := txScriptLineRepo.Create(ctx, scriptLine); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 作成した行を再取得（Speaker 情報を含める）
	createdLine, err := s.scriptLineRepo.FindByID(ctx, scriptLine.ID)
	if err != nil {
		return nil, err
	}

	// レスポンスに変換
	resp := s.toScriptLineResponse(createdLine)

	return &resp, nil
}

// 指定された台本行を更新する
func (s *scriptLineService) Update(ctx context.Context, userID, channelID, episodeID, lineID string, req request.UpdateScriptLineRequest) (*response.ScriptLineResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	lid, err := uuid.Parse(lineID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 台本行の存在確認とエピソードの一致チェック
	scriptLine, err := s.scriptLineRepo.FindByID(ctx, lid)
	if err != nil {
		return nil, err
	}

	if scriptLine.EpisodeID != eid {
		return nil, apperror.ErrNotFound.WithMessage("このエピソードに台本行が見つかりません")
	}

	// フィールドを更新
	if req.Text != nil {
		scriptLine.Text = *req.Text
	}

	if req.Emotion != nil {
		if *req.Emotion == "" {
			scriptLine.Emotion = nil
		} else {
			scriptLine.Emotion = req.Emotion
		}
	}

	if err := s.scriptLineRepo.Update(ctx, scriptLine); err != nil {
		return nil, err
	}

	// レスポンスに変換
	resp := s.toScriptLineResponse(scriptLine)

	return &resp, nil
}

// 指定された台本行を削除する
func (s *scriptLineService) Delete(ctx context.Context, userID, channelID, episodeID, lineID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return err
	}

	lid, err := uuid.Parse(lineID)
	if err != nil {
		return err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return err
	}

	if channel.UserID != uid {
		return apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return err
	}

	if episode.ChannelID != cid {
		return apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 台本行の存在確認とエピソードの一致チェック
	scriptLine, err := s.scriptLineRepo.FindByID(ctx, lid)
	if err != nil {
		return err
	}

	if scriptLine.EpisodeID != eid {
		return apperror.ErrNotFound.WithMessage("このエピソードに台本行が見つかりません")
	}

	// 台本行を削除
	return s.scriptLineRepo.Delete(ctx, lid)
}

// 台本行の順序を並び替える
func (s *scriptLineService) Reorder(ctx context.Context, userID, channelID, episodeID string, req request.ReorderScriptLinesRequest) (*response.ScriptLineListResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	cid, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}

	eid, err := uuid.Parse(episodeID)
	if err != nil {
		return nil, err
	}

	// チャンネルの存在確認とオーナーチェック
	channel, err := s.channelRepo.FindByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if channel.UserID != uid {
		return nil, apperror.ErrForbidden.WithMessage("このチャンネルへのアクセス権限がありません")
	}

	// エピソードの存在確認とチャンネルの一致チェック
	episode, err := s.episodeRepo.FindByID(ctx, eid)
	if err != nil {
		return nil, err
	}

	if episode.ChannelID != cid {
		return nil, apperror.ErrNotFound.WithMessage("このチャンネルにエピソードが見つかりません")
	}

	// 重複チェック
	lineIDSet := make(map[string]struct{}, len(req.LineIDs))
	for _, id := range req.LineIDs {
		if _, exists := lineIDSet[id]; exists {
			return nil, apperror.ErrValidation.WithMessage("リクエストに重複した lineId があります")
		}
		lineIDSet[id] = struct{}{}
	}

	// lineIDs を uuid.UUID に変換
	lineUUIDs := make([]uuid.UUID, len(req.LineIDs))
	for i, id := range req.LineIDs {
		lineUUID, err := uuid.Parse(id)
		if err != nil {
			return nil, apperror.ErrValidation.WithMessage("lineId の形式が無効です")
		}
		lineUUIDs[i] = lineUUID
	}

	// 指定された行を取得
	scriptLines, err := s.scriptLineRepo.FindByIDs(ctx, lineUUIDs)
	if err != nil {
		return nil, err
	}

	// 全ての行が見つかったか確認
	if len(scriptLines) != len(req.LineIDs) {
		return nil, apperror.ErrNotFound.WithMessage("一部の台本行が見つかりません")
	}

	// 全ての行が対象エピソードに属しているか確認
	for _, sl := range scriptLines {
		if sl.EpisodeID != eid {
			return nil, apperror.ErrNotFound.WithMessage("このエピソードに台本行が見つかりません")
		}
	}

	// lineOrder のマッピングを作成
	lineOrders := make(map[uuid.UUID]int, len(lineUUIDs))
	for i, id := range lineUUIDs {
		lineOrders[id] = i
	}

	// トランザクションで lineOrder を更新
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txScriptLineRepo := repository.NewScriptLineRepository(tx)
		return txScriptLineRepo.UpdateLineOrders(ctx, lineOrders)
	})

	if err != nil {
		return nil, err
	}

	// 更新後の台本行一覧を取得
	updatedLines, err := s.scriptLineRepo.FindByEpisodeID(ctx, eid)
	if err != nil {
		return nil, err
	}

	// レスポンスに変換
	responses := s.toScriptLineResponses(updatedLines)

	return &response.ScriptLineListResponse{
		Data: responses,
	}, nil
}

// ScriptLine モデルのスライスをレスポンス DTO のスライスに変換する
func (s *scriptLineService) toScriptLineResponses(scriptLines []model.ScriptLine) []response.ScriptLineResponse {
	result := make([]response.ScriptLineResponse, len(scriptLines))

	for i, sl := range scriptLines {
		result[i] = s.toScriptLineResponse(&sl)
	}

	return result
}

// ScriptLine モデルをレスポンス DTO に変換する
func (s *scriptLineService) toScriptLineResponse(sl *model.ScriptLine) response.ScriptLineResponse {
	return response.ScriptLineResponse{
		ID:        sl.ID,
		LineOrder: sl.LineOrder,
		Speaker: response.SpeakerResponse{
			ID:      sl.Speaker.ID,
			Name:    sl.Speaker.Name,
			Persona: sl.Speaker.Persona,
			Voice: response.CharacterVoiceResponse{
				ID:       sl.Speaker.Voice.ID,
				Name:     sl.Speaker.Voice.Name,
				Provider: sl.Speaker.Voice.Provider,
				Gender:   string(sl.Speaker.Voice.Gender),
			},
		},
		Text:      sl.Text,
		Emotion:   sl.Emotion,
		CreatedAt: sl.CreatedAt,
		UpdatedAt: sl.UpdatedAt,
	}
}
