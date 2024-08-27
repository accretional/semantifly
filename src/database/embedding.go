package database

import (
    "context"
    "fmt"
    "os"

    "github.com/sashabaranov/go-openai"
)

type OpenAIEmbeddingGenerator struct {
    client *openai.Client
}

func NewEmbeddingGenerator() (*OpenAIEmbeddingGenerator, error) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
    }

    client := openai.NewClient(apiKey)
    return &OpenAIEmbeddingGenerator{client: client}, nil
}

func (g *OpenAIEmbeddingGenerator) GenerateEmbedding(text string) ([]float32, error) {
    resp, err := g.client.CreateEmbeddings(
        context.Background(),
        openai.EmbeddingRequest{
            Model: openai.AdaEmbeddingV2,
            Input: []string{text},
        },
    )
    if err != nil {
        return nil, fmt.Errorf("failed to generate embedding: %w", err)
    }

    if len(resp.Data) == 0 {
        return nil, fmt.Errorf("no embedding data returned")
    }

    return resp.Data[0].Embedding, nil
}