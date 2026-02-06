package llm

import "fmt"

// Registry は複数の LLM クライアントを管理する
type Registry struct {
	clients map[Provider]Client
}

// NewRegistry は空の Registry を生成する
func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[Provider]Client),
	}
}

// Register はプロバイダに対応するクライアントを登録する
func (r *Registry) Register(provider Provider, client Client) {
	r.clients[provider] = client
}

// Get は指定されたプロバイダのクライアントを返す
//
// 未登録の場合はエラーを返す
func (r *Registry) Get(provider Provider) (Client, error) {
	client, ok := r.clients[provider]
	if !ok {
		return nil, fmt.Errorf("LLM provider %q is not registered", provider)
	}
	return client, nil
}

// Has は指定されたプロバイダが登録済みかどうかを返す
func (r *Registry) Has(provider Provider) bool {
	_, ok := r.clients[provider]
	return ok
}

// GetModelInfo は指定されたプロバイダのモデル情報を返す
func (r *Registry) GetModelInfo(provider Provider) string {
	client, ok := r.clients[provider]
	if !ok {
		return string(provider)
	}
	return client.ModelInfo()
}
