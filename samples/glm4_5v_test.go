package samples

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

// TestGLM45VVideoProcessing 测试GLM-4.5v视频处理功能
// 使用Realtime SDK格式的输入文件,调用GLM-4.5v HTTP API
//
// 使用方法:
//
//	方式一: 配置 .env 文件
//	cp samples/.env.example samples/.env
//	编辑 samples/.env 填入你的 API Key
//	go test -v ./samples -run TestGLM45VVideoProcessing
//
//	方式二: 直接导出环境变量
//	export ZHIPU_API_KEY="your_api_key"
//	go test -v ./samples -run TestGLM45VVideoProcessing
func TestGLM45VVideoProcessing(t *testing.T) {
	// 尝试加载 .env 文件(如果存在)
	_ = godotenv.Load(".env")

	// 检查环境变量
	if os.Getenv("ZHIPU_API_KEY") == "" {
		t.Skip("ZHIPU_API_KEY environment variable not set, skipping test")
	}

	inputFile := filepath.Join("files", "Video.ClientVad.Input")
	outputFile := filepath.Join("files", "GLM45V.Output")
	prompt := "请描述这个视频的内容"

	// 确保输入文件存在
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Fatalf("Input file not found: %s", inputFile)
	}

	// 调用处理函数
	response, err := ProcessVideoWithGLM45V(inputFile, prompt, outputFile)
	if err != nil {
		t.Fatalf("ProcessVideoWithGLM45V failed: %v", err)
	}

	// 验证响应
	if len(response.Choices) == 0 {
		t.Fatal("No response choices returned")
	}

	// 打印结果
	log.Println("\n=== GLM-4.5v Response ===")
	log.Println(response.Choices[0].Message.Content)

	log.Println("\n=== Token Usage ===")
	log.Printf("Prompt: %d | Completion: %d | Total: %d\n",
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
		response.Usage.TotalTokens)

	// 验证输出文件
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file not created: %s", outputFile)
	}
}
