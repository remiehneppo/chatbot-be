package services

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var (
	SystemMessageInitiateMechanicalEngineer = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a mechanical engineer. You can answer questions about mechanical engineering. If you do not know the answer, you can research it the database before responding.",
	}
)

type OpenAIService struct {
	client        *openai.Client
	functionsCall map[string]FunctionHandler
	tools         []openai.Tool
}

func NewOpenAIService(baseURL string, apiKey string) *OpenAIService {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL // Set this to your local LLM server URL
	client := openai.NewClientWithConfig(config)
	return &OpenAIService{
		client: client,
	}
}

func (s *OpenAIService) Chat(ctx context.Context, messages []Message) (*Message, error) {
	// Convert our Message type to OpenAI chat messages
	openaiMessages := make([]openai.ChatCompletionMessage, 0)
	openaiMessages = append(openaiMessages, SystemMessageInitiateMechanicalEngineer)
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Content,
		})
	}

	// Create chat completion request
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Messages: openaiMessages,
			Tools:    s.tools,
			Stream:   true,
		},
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response generated")
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		resp, err = s.handleFunctionCall(ctx, openaiMessages, resp)
		if err != nil {
			return nil, err
		}

	}

	// Convert response back to our Message type
	return &Message{
		Role:    "assistant",
		Content: resp.Choices[0].Message.Content,
	}, nil
}

func (s *OpenAIService) ChatStream(ctx context.Context, messages []Message, streamHandler StreamHandler) error {
	// Convert our Message type to OpenAI chat messages
	openaiMessages := make([]openai.ChatCompletionMessage, 0)
	openaiMessages = append(openaiMessages, SystemMessageInitiateMechanicalEngineer)
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Content,
		})
	}

	// Create chat completion request
	stream, err := s.client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Messages: openaiMessages,
			Tools:    s.tools,
		},
	)
	if err != nil {
		return err
	}
	defer stream.Close()
	for {
		resp, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Println("Error receiving response from stream:", err)
		}
		streamHandler(resp.Choices[0].Delta.Content)
	}

}

func (s *OpenAIService) RegisterFunctionCall(name, description string, params jsonschema.Definition, handler FunctionHandler) {
	if s.functionsCall == nil {
		s.functionsCall = make(map[string]FunctionHandler)
	}
	f := openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters:  params,
	}
	t := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f,
	}
	s.functionsCall[name] = handler
	s.tools = append(s.tools, t)
}

func (s *OpenAIService) handleFunctionCall(ctx context.Context, openaiMessages []openai.ChatCompletionMessage, resp openai.ChatCompletionResponse) (openai.ChatCompletionResponse, error) {
	openaiMessages = append(openaiMessages, resp.Choices[0].Message)
	for _, toolCall := range resp.Choices[0].Message.ToolCalls {
		if toolCall.Type == openai.ToolTypeFunction {
			handler := s.functionsCall[toolCall.Function.Name]
			if handler == nil {
				return openai.ChatCompletionResponse{}, errors.New("no handler found for function call")
			}
			result, err := handler(ctx, []byte(toolCall.Function.Arguments))
			if err != nil {
				return openai.ChatCompletionResponse{}, err
			}
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result.(string),
				Name:       toolCall.Function.Name,
				ToolCallID: toolCall.ID,
			})
		}
	}
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Messages: openaiMessages,
			Tools:    s.tools,
		},
	)
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	if len(resp.Choices) == 0 {
		return openai.ChatCompletionResponse{}, errors.New("no response generated")
	}
	if resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		return s.handleFunctionCall(ctx, openaiMessages, resp)
	}
	return resp, nil
}
