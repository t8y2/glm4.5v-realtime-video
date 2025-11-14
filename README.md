# GLM-4.5v 视频处理 Golang SDK

本项目提供基于智谱 AI GLM-4.5v 模型的视频/图片理解功能,可从 Realtime SDK 格式的输入文件中提取视频帧并进行智能分析。

## 接口文档

最新接口文档参考 https://open.bigmodel.cn/dev/api/rtav/GLM-Realtime

## 项目结构

```Text
.
├── README.md                        # 项目说明文档
├── client                           # SDK 核心代码
│   └── client.go
├── events                           # 数据模型定义
│   ├── event.go
│   ├── items.go
│   ├── response.go
│   └── tools.go
├── go.mod
├── go.sum
└── samples                          # 示例代码目录
    ├── .env.example                 # 环境变量示例文件
    ├── glm4_5v_client.go            # GLM-4.5v 视频处理客户端
    ├── glm4_5v_test.go              # GLM-4.5v 测试
    └── files                        # 示例输入输出数据目录
        ├── Video.ClientVad.Input    # 视频输入数据(含视频帧)
        └── pics
            └── kunkun.jpg           # 示例图片
```

## 快速开始

### 1. 环境准备

首先确保您已安装 Golang 1.22.3 或更高版本。

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置 API 密钥

您需要设置 ZHIPU_API_KEY 环境变量。可以通过以下两种方式之一进行设置：

#### 方式一：直接设置环境变量

```bash
export ZHIPU_API_KEY=your_api_key
```

#### 方式二：使用 .env 文件(推荐)

在 `samples/` 目录下创建 `.env` 文件：

```bash
cd samples
cp .env.example .env
```

然后编辑 `.env` 文件，填入您的 API 密钥：

```
ZHIPU_API_KEY=your_api_key
```

> 注：API 密钥可在 [智谱 AI 开放平台](https://www.bigmodel.cn/) 注册开发者账号后创建获取

### 4. 运行示例

运行 GLM-4.5v 视频处理测试:

```bash
cd golang
go test -v ./samples -run TestGLM45VVideoProcessing
```

或者在 IDE 中直接运行测试文件 `samples/glm4_5v_test.go`。

## GLM-4.5v 视频处理功能

### 功能说明

- 从 Realtime SDK 输入文件中提取视频帧(JPEG 格式)
- 调用智谱 GLM-4.5v API 进行图像分析
- 支持多帧图像处理
- 输出 JSON 格式的分析结果

### 使用示例

```go
import "github.com/t8y2/glm4.5v-realtime-video/golang/samples"

// 处理视频文件
response, err := samples.ProcessVideoWithGLM45V(
    "files/Video.ClientVad.Input",  // 输入文件路径
    "请描述视频的内容",              // 提示词
    "files/Video.ClientVad.Output",          // 输出文件路径(可选)
)

if err != nil {
    log.Fatal(err)
}

// 打印结果
fmt.Println(response.Choices[0].Message.Content)
```

### API 函数

#### ProcessVideoWithGLM45V

主处理函数,完整流程。API Key 从环境变量 `ZHIPU_API_KEY` 读取。

```go
func ProcessVideoWithGLM45V(
    inputFilePath string,  // 输入文件路径
    prompt string,         // 提示词
    outputFilePath string  // 输出文件路径(可为空)
) (*GLM45VResponse, error)
```

#### ExtractVideoFramesFromRealtimeFile

从输入文件提取视频帧。

```go
func ExtractVideoFramesFromRealtimeFile(
    inputFilePath string
) ([][]byte, error)
```

#### CallGLM45V

直接调用 GLM-4.5v API。

```go
func CallGLM45V(
    apiKey string,
    frames [][]byte,  // JPEG图像数据
    prompt string
) (*GLM45VResponse, error)
```

## 许可证

本项目采用 [LICENSE.md](../LICENSE.md) 中规定的许可证。
