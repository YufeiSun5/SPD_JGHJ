package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AI 服务配置
const AI_SERVICE_URL = "http://127.0.0.1:8006"

// 全局流式请求管理
var (
	currentStreamCancel context.CancelFunc
	streamMutex         sync.Mutex
)

// ========================================================
// 请求和响应结构体
// ========================================================

// AIQueryRequest AI 查询请求
type AIQueryRequest struct {
	Question string `json:"question"`
}

// AIQueryResponse AI 查询响应
type AIQueryResponse struct {
	Answer       string        `json:"answer"`
	RelevantDocs []RelevantDoc `json:"relevant_docs"`
	ContextUsed  bool          `json:"context_used"`
	Error        string        `json:"error,omitempty"`
}

// RelevantDoc 相关文档
type RelevantDoc struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Source     string  `json:"source"`
	Similarity float64 `json:"similarity"`
}

// AddKnowledgeRequest 添加知识请求
type AddKnowledgeRequest struct {
	Content string  `json:"content"`
	Source  *string `json:"source,omitempty"`
}

// AddKnowledgeResponse 添加知识响应
type AddKnowledgeResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	KnowledgeID string `json:"knowledge_id,omitempty"`
}

// DeleteKnowledgeResponse 删除知识响应
type DeleteKnowledgeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// KnowledgeItem 知识条目
type KnowledgeItem struct {
	ID        string  `json:"id"`
	Content   string  `json:"content"`
	Source    *string `json:"source"`
	CreatedAt *string `json:"created_at"`
}

// KnowledgeListResponse 知识列表响应
type KnowledgeListResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    []KnowledgeItem `json:"data"`
	Total   int             `json:"total"`
}

// StreamChunk 流式数据块
type StreamChunk struct {
	Type     string                 `json:"type"`     // thinking, answer, database_result, done
	Data     interface{}            `json:"data"`     // 可能是字符串或对象
	Metadata map[string]interface{} `json:"metadata"` // 额外元数据
}

// ========================================================
// AI 服务接口实现
// ========================================================

// QueryAI 调用 AI 知识库问答（普通模式）
func (a *App) QueryAI(question string) (*AIQueryResponse, error) {
	// 构建请求
	reqBody := AIQueryRequest{Question: question}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送 HTTP 请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(
		AI_SERVICE_URL+"/api/knowledge/query",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("AI服务连接失败，请确保 FastAPI 服务运行在 %s", AI_SERVICE_URL)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI服务返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result AIQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &result, nil
}

// QueryAIStream 调用 AI 知识库问答（流式模式）
func (a *App) QueryAIStream(question string) error {
	fmt.Printf("🤖 [AI Stream] 开始流式查询: %s\n", question)

	// 取消之前的流式请求（如果有）
	streamMutex.Lock()
	if currentStreamCancel != nil {
		fmt.Printf("⚠️ [AI Stream] 取消之前的流式请求\n")
		currentStreamCancel()
	}

	// 创建新的可取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	currentStreamCancel = cancel
	streamMutex.Unlock()

	// 构建请求
	reqBody := AIQueryRequest{Question: question}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		runtime.EventsEmit(a.ctx, "ai-stream-error", fmt.Sprintf("序列化请求失败: %v", err))
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送 HTTP 请求（流式）
	fmt.Printf("🌐 [AI Stream] 正在连接 FastAPI: %s\n", AI_SERVICE_URL+"/api/knowledge/query-stream")
	client := &http.Client{Timeout: 120 * time.Second}

	// 创建带上下文的请求
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		AI_SERVICE_URL+"/api/knowledge/query-stream",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		runtime.EventsEmit(a.ctx, "ai-stream-error", fmt.Sprintf("创建请求失败: %v", err))
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// 检查是否是主动取消
		if ctx.Err() == context.Canceled {
			fmt.Printf("🛑 [AI Stream] 流式请求已被用户取消\n")
			runtime.EventsEmit(a.ctx, "ai-stream-cancelled", nil)
			return nil
		}
		errMsg := fmt.Sprintf("AI服务连接失败，请确保 FastAPI 服务运行在 %s", AI_SERVICE_URL)
		fmt.Printf("❌ [AI Stream] %s\n", errMsg)
		runtime.EventsEmit(a.ctx, "ai-stream-error", errMsg)
		return fmt.Errorf(errMsg)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("AI服务返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
		fmt.Printf("❌ [AI Stream] %s\n", errMsg)
		runtime.EventsEmit(a.ctx, "ai-stream-error", errMsg)
		return fmt.Errorf(errMsg)
	}

	fmt.Printf("✅ [AI Stream] 连接成功，开始接收流式数据...\n")

	// 读取 SSE 流
	reader := bufio.NewReader(resp.Body)
	chunkCount := 0
	for {
		// 检查是否被取消
		select {
		case <-ctx.Done():
			fmt.Printf("🛑 [AI Stream] 流式传输被取消，已接收 %d 个数据块\n", chunkCount)
			runtime.EventsEmit(a.ctx, "ai-stream-cancelled", nil)
			return nil
		default:
		}

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Printf("📝 [AI Stream] 流结束，共接收 %d 个数据块\n", chunkCount)
			break
		}
		if err != nil {
			// 检查是否是取消导致的错误
			if ctx.Err() == context.Canceled {
				fmt.Printf("🛑 [AI Stream] 流式传输被取消\n")
				runtime.EventsEmit(a.ctx, "ai-stream-cancelled", nil)
				return nil
			}
			errMsg := fmt.Sprintf("读取流失败: %v", err)
			fmt.Printf("❌ [AI Stream] %s\n", errMsg)
			runtime.EventsEmit(a.ctx, "ai-stream-error", errMsg)
			return fmt.Errorf(errMsg)
		}

		// 解析 SSE 格式: data: {...}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")

			// 解析 JSON
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
				fmt.Printf("⚠️ [AI Stream] JSON 解析失败: %s\n", jsonStr[:50])
				continue // 跳过无效的 JSON
			}

			chunkCount++
			chunkType := chunk["type"]
			fmt.Printf("📦 [AI Stream] 接收数据块 #%d, 类型: %v\n", chunkCount, chunkType)

			// 发送事件到前端
			runtime.EventsEmit(a.ctx, "ai-stream-chunk", chunk)
		}
	}

	// 流结束，清理取消函数
	streamMutex.Lock()
	currentStreamCancel = nil
	streamMutex.Unlock()

	// 流结束
	fmt.Printf("🏁 [AI Stream] 发送结束事件\n")
	runtime.EventsEmit(a.ctx, "ai-stream-end", nil)
	return nil
}

// StopAIStream 停止当前的流式查询
func (a *App) StopAIStream() bool {
	streamMutex.Lock()
	defer streamMutex.Unlock()

	if currentStreamCancel != nil {
		fmt.Printf("🛑 [AI Stream] 用户请求停止流式传输\n")
		currentStreamCancel()
		currentStreamCancel = nil
		return true
	}

	fmt.Printf("⚠️ [AI Stream] 没有正在进行的流式传输\n")
	return false
}

// AddKnowledge 添加知识到知识库
func (a *App) AddKnowledge(content string, source *string) (*AddKnowledgeResponse, error) {
	// 构建请求
	reqBody := AddKnowledgeRequest{
		Content: content,
		Source:  source,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送 HTTP 请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(
		AI_SERVICE_URL+"/api/knowledge/add",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result AddKnowledgeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return &result, nil
}

// DeleteKnowledge 删除知识
func (a *App) DeleteKnowledge(knowledgeID string) (*DeleteKnowledgeResponse, error) {
	// 发送 HTTP DELETE 请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(
		"DELETE",
		AI_SERVICE_URL+"/api/knowledge/delete/"+knowledgeID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result DeleteKnowledgeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return &result, nil
}

// DeleteKnowledgeBySource 删除指定来源的所有知识（用于替换式导入）
func (a *App) DeleteKnowledgeBySource(source string) (*DeleteKnowledgeResponse, error) {
	// 发送 HTTP DELETE 请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(
		"DELETE",
		AI_SERVICE_URL+"/api/knowledge/delete_by_source/"+source,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result DeleteKnowledgeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	return &result, nil
}

// GetKnowledgeList 获取知识库列表
func (a *App) GetKnowledgeList() (*KnowledgeListResponse, error) {
	// 发送 HTTP GET 请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(AI_SERVICE_URL + "/api/knowledge/list")
	if err != nil {
		return nil, fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result KnowledgeListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &result, nil
}

// CheckAIServiceHealth 检查 AI 服务健康状态
func (a *App) CheckAIServiceHealth() bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(AI_SERVICE_URL + "/")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ========================================================
// 点赞和排队功能
// ========================================================

// LikeAIAnswerRequest 点赞请求
type LikeAIAnswerRequest struct {
	Question     string        `json:"question"`
	Answer       string        `json:"answer"`
	RelevantDocs []RelevantDoc `json:"relevant_docs"`
}

// LikeAIAnswerResponse 点赞响应
type LikeAIAnswerResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	FeedbackID string `json:"feedback_id,omitempty"`
}

// LikeAIAnswer 点赞AI回答
func (a *App) LikeAIAnswer(question string, answer string, relevantDocs []RelevantDoc) error {
	fmt.Printf("👍 [AI Feedback] 用户点赞回答\n")

	// 构建请求
	reqBody := LikeAIAnswerRequest{
		Question:     question,
		Answer:       answer,
		RelevantDocs: relevantDocs,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送 HTTP 请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		AI_SERVICE_URL+"/api/knowledge/feedback/like",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("点赞失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result LikeAIAnswerResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf(result.Message)
	}

	fmt.Printf("✅ [AI Feedback] 点赞成功: %s\n", result.Message)
	return nil
}

// QueryAIStreamWithQueue 调用 AI 知识库问答（流式模式+排队）
func (a *App) QueryAIStreamWithQueue(question string) error {
	fmt.Printf("🤖 [AI Stream Queue] 开始流式查询（带排队）: %s\n", question)

	// 取消之前的流式请求（如果有）
	streamMutex.Lock()
	if currentStreamCancel != nil {
		fmt.Printf("⚠️ [AI Stream Queue] 取消之前的流式请求\n")
		currentStreamCancel()
	}

	// 创建新的可取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	currentStreamCancel = cancel
	streamMutex.Unlock()

	// 构建请求
	reqBody := AIQueryRequest{Question: question}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		runtime.EventsEmit(a.ctx, "ai-stream-error", fmt.Sprintf("序列化请求失败: %v", err))
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送 HTTP 请求（流式+排队）
	fmt.Printf("🌐 [AI Stream Queue] 正在连接 FastAPI: %s\n", AI_SERVICE_URL+"/api/knowledge/query-stream-with-queue")
	client := &http.Client{Timeout: 180 * time.Second} // 增加超时时间以支持排队

	// 创建带上下文的请求
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		AI_SERVICE_URL+"/api/knowledge/query-stream-with-queue",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		runtime.EventsEmit(a.ctx, "ai-stream-error", fmt.Sprintf("创建请求失败: %v", err))
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// 检查是否是主动取消
		if ctx.Err() == context.Canceled {
			fmt.Printf("🛑 [AI Stream Queue] 流式请求已被用户取消\n")
			runtime.EventsEmit(a.ctx, "ai-stream-cancelled", nil)
			return nil
		}
		errMsg := fmt.Sprintf("AI服务连接失败，请确保 FastAPI 服务运行在 %s", AI_SERVICE_URL)
		fmt.Printf("❌ [AI Stream Queue] %s\n", errMsg)
		runtime.EventsEmit(a.ctx, "ai-stream-error", errMsg)
		return fmt.Errorf(errMsg)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("AI服务返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
		fmt.Printf("❌ [AI Stream Queue] %s\n", errMsg)
		runtime.EventsEmit(a.ctx, "ai-stream-error", errMsg)
		return fmt.Errorf(errMsg)
	}

	fmt.Printf("✅ [AI Stream Queue] 连接成功，开始接收流式数据...\n")

	// 读取 SSE 流
	reader := bufio.NewReader(resp.Body)
	chunkCount := 0
	for {
		// 检查是否被取消
		select {
		case <-ctx.Done():
			fmt.Printf("🛑 [AI Stream Queue] 流式传输被取消，已接收 %d 个数据块\n", chunkCount)
			runtime.EventsEmit(a.ctx, "ai-stream-cancelled", nil)
			return nil
		default:
		}

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Printf("📝 [AI Stream Queue] 流结束，共接收 %d 个数据块\n", chunkCount)
			break
		}
		if err != nil {
			// 检查是否是取消导致的错误
			if ctx.Err() == context.Canceled {
				fmt.Printf("🛑 [AI Stream Queue] 流式传输被取消\n")
				runtime.EventsEmit(a.ctx, "ai-stream-cancelled", nil)
				return nil
			}
			errMsg := fmt.Sprintf("读取流失败: %v", err)
			fmt.Printf("❌ [AI Stream Queue] %s\n", errMsg)
			runtime.EventsEmit(a.ctx, "ai-stream-error", errMsg)
			return fmt.Errorf(errMsg)
		}

		// 解析 SSE 格式: data: {...}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")

			// 解析 JSON
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
				fmt.Printf("⚠️ [AI Stream Queue] JSON 解析失败: %s\n", jsonStr[:50])
				continue // 跳过无效的 JSON
			}

			chunkCount++
			chunkType := chunk["type"]
			fmt.Printf("📦 [AI Stream Queue] 接收数据块 #%d, 类型: %v\n", chunkCount, chunkType)

			// 发送事件到前端
			runtime.EventsEmit(a.ctx, "ai-stream-chunk", chunk)
		}
	}

	// 流结束，清理取消函数
	streamMutex.Lock()
	currentStreamCancel = nil
	streamMutex.Unlock()

	// 流结束
	fmt.Printf("🏁 [AI Stream Queue] 发送结束事件\n")
	runtime.EventsEmit(a.ctx, "ai-stream-end", nil)
	return nil
}
