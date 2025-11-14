package samples

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GLM45VRequest GLM-4.5v API 请求结构
type GLM45VRequest struct {
	Model    string          `json:"model"`
	Messages []GLM45VMessage `json:"messages"`
}

type GLM45VMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ImageContent struct {
	Type     string   `json:"type"`
	ImageURL ImageURL `json:"image_url"`
}

type ImageURL struct {
	URL string `json:"url"`
}

// GLM45VResponse GLM-4.5v API 响应结构
type GLM45VResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ProcessVideoWithGLM45V 处理视频输入文件,调用GLM-4.5v API
// 参数:
//   - inputFilePath: Realtime SDK格式的输入文件路径(包含video_frame事件)
//   - prompt: 用户提示词
//   - outputFilePath: 输出文件路径(可选,为空则不写入文件)
//
// 返回:
//   - *GLM45VResponse: API响应
//   - error: 错误信息
//
// 注意: API Key 从环境变量 ZHIPU_API_KEY 读取
func ProcessVideoWithGLM45V(inputFilePath string, prompt string, outputFilePath string) (*GLM45VResponse, error) {
	// 从环境变量读取 API Key
	apiKey := os.Getenv("ZHIPU_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ZHIPU_API_KEY environment variable not set")
	}

	// 从输入文件提取视频帧
	frames, err := ExtractVideoFramesFromRealtimeFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("extract video frames failed: %v", err)
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("no video frames found in input file")
	}

	log.Printf("Extracted %d video frames from input file\n", len(frames))

	// 调用 GLM-4.5v API
	response, err := CallGLM45V(apiKey, frames, prompt)
	if err != nil {
		return nil, fmt.Errorf("call GLM-4.5v API failed: %v", err)
	}

	// 写入输出文件(如果指定)
	if outputFilePath != "" {
		if err := WriteResponseToFile(response, outputFilePath); err != nil {
			log.Printf("Warning: failed to write output file: %v\n", err)
		}
	}

	return response, nil
}

// ExtractVideoFramesFromRealtimeFile 从Realtime SDK输入文件中提取视频帧
// 输入文件格式: 每行一个JSON事件,包含 input_audio_buffer.append_video_frame 类型的事件
// 视频帧存储在 video_frame 字段中,为base64编码的JPEG数据
func ExtractVideoFramesFromRealtimeFile(inputFilePath string) ([][]byte, error) {
	file, err := os.Open(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()

	var frames [][]byte
	scanner := bufio.NewScanner(file)
	// 增大缓冲区以处理大的base64编码数据
	scanner.Buffer(make([]byte, 0, 10*1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行和注释
		if !strings.HasPrefix(line, "{") {
			continue
		}

		// 解析JSON事件
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// 提取视频帧事件
		if eventType, ok := event["type"].(string); ok && eventType == "input_audio_buffer.append_video_frame" {
			if videoFrameStr, ok := event["video_frame"].(string); ok {
				// 解码base64
				frameData, err := base64.StdEncoding.DecodeString(videoFrameStr)
				if err != nil {
					log.Printf("Warning: decode video frame failed: %v\n", err)
					continue
				}
				frames = append(frames, frameData)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read file failed: %v", err)
	}

	return frames, nil
}

// CallGLM45V 调用 GLM-4.5v API
func CallGLM45V(apiKey string, frames [][]byte, prompt string) (*GLM45VResponse, error) {
	// 构建消息内容
	var contents []interface{}

	// 添加文本提示
	if prompt != "" {
		contents = append(contents, TextContent{
			Type: "text",
			Text: prompt,
		})
	}

	// 添加图片(JPEG帧)
	for _, frame := range frames {
		base64Img := base64.StdEncoding.EncodeToString(frame)
		contents = append(contents, ImageContent{
			Type: "image_url",
			ImageURL: ImageURL{
				URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Img),
			},
		})
	}

	// 构建请求
	req := GLM45VRequest{
		Model: "glm-4.5v",
		Messages: []GLM45VMessage{
			{
				Role:    "user",
				Content: contents,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", "https://open.bigmodel.cn/api/paas/v4/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer httpResp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	// 检查HTTP状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	// 解析响应
	var response GLM45VResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %v", err)
	}

	return &response, nil
}

// WriteResponseToFile 将GLM-4.5v响应写入文件（追加模式，每行一个JSON事件）
func WriteResponseToFile(response *GLM45VResponse, outputPath string) error {
	// 以追加模式打开文件
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()

	// 构建响应事件
	outputEvent := map[string]interface{}{
		"type":      "glm4.5v.response",
		"content":   "",
		"timestamp": time.Now().Unix(),
		"usage": map[string]int{
			"prompt_tokens":     response.Usage.PromptTokens,
			"completion_tokens": response.Usage.CompletionTokens,
			"total_tokens":      response.Usage.TotalTokens,
		},
	}

	if len(response.Choices) > 0 {
		outputEvent["content"] = response.Choices[0].Message.Content
	}

	outputJSON, err := json.Marshal(outputEvent)
	if err != nil {
		return fmt.Errorf("marshal output failed: %v", err)
	}

	// 写入文件，追加换行符
	if _, err := file.WriteString(string(outputJSON) + "\n"); err != nil {
		return fmt.Errorf("write to file failed: %v", err)
	}

	log.Printf("Response written to file: %s\n", outputPath)
	return nil
}
