package cloudtasks

import (
	"context"
	"encoding/json"
	"fmt"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"google.golang.org/api/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// Client は Cloud Tasks クライアントのインターフェース
type Client interface {
	EnqueueAudioJob(ctx context.Context, jobID string) error
	Close() error
}

// Config は Cloud Tasks クライアントの設定
type Config struct {
	ProjectID           string
	Location            string
	QueueName           string
	ServiceAccountEmail string
	WorkerEndpointURL   string
	CredentialsJSON     string
}

type client struct {
	tasksClient         *cloudtasks.Client
	queuePath           string
	serviceAccountEmail string
	workerEndpointURL   string
}

// NewClient は Cloud Tasks クライアントを作成する
func NewClient(ctx context.Context, cfg Config) (Client, error) {
	var opts []option.ClientOption
	if cfg.CredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON))) //nolint:staticcheck // TODO: migrate to newer auth method
	}

	tasksClient, err := cloudtasks.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud tasks client: %w", err)
	}

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		cfg.ProjectID, cfg.Location, cfg.QueueName)

	return &client{
		tasksClient:         tasksClient,
		queuePath:           queuePath,
		serviceAccountEmail: cfg.ServiceAccountEmail,
		workerEndpointURL:   cfg.WorkerEndpointURL,
	}, nil
}

// AudioJobPayload はワーカーに送信されるペイロード
type AudioJobPayload struct {
	JobID string `json:"jobId"`
}

// EnqueueAudioJob は音声生成ジョブをキューに追加する
func (c *client) EnqueueAudioJob(ctx context.Context, jobID string) error {
	log := logger.FromContext(ctx)

	payload := AudioJobPayload{JobID: jobID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error("failed to marshal payload", "error", err)
		return apperror.ErrInternal.WithMessage("ペイロードのシリアライズに失敗しました").WithError(err)
	}

	req := &taskspb.CreateTaskRequest{
		Parent: c.queuePath,
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        c.workerEndpointURL,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: payloadBytes,
					AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
						OidcToken: &taskspb.OidcToken{
							ServiceAccountEmail: c.serviceAccountEmail,
							Audience:            c.workerEndpointURL,
						},
					},
				},
			},
		},
	}

	log.Info("enqueuing audio job task", "job_id", jobID, "queue", c.queuePath)

	task, err := c.tasksClient.CreateTask(ctx, req)
	if err != nil {
		log.Error("failed to create task", "error", err, "job_id", jobID)
		return apperror.ErrInternal.WithMessage("タスクの作成に失敗しました").WithError(err)
	}

	log.Info("task created successfully", "task_name", task.Name, "job_id", jobID)
	return nil
}

// Close はクライアントを閉じる
func (c *client) Close() error {
	return c.tasksClient.Close()
}
