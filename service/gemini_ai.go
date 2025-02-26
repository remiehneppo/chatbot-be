package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiService struct {
	apiKeys       []string
	currentKey    int
	client        *genai.Client
	model         *genai.GenerativeModel
	functionsCall map[string]FunctionHandler
	mu            sync.Mutex
}

func defaultStreamHandler(response string) {
	println(response)
}

func NewGeminiService(apiKeys []string, modelName string) (*GeminiService, error) {
	if len(apiKeys) == 0 {
		return nil, errors.New("no API keys provided")
	}

	service := &GeminiService{
		apiKeys:       apiKeys,
		currentKey:    0,
		functionsCall: make(map[string]FunctionHandler),
	}

	err := service.initClient()
	if err != nil {
		return nil, err
	}

	service.model = service.client.GenerativeModel(modelName)
	return service, nil
}

func (s *GeminiService) initClient() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(s.apiKeys[s.currentKey]))
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

func (s *GeminiService) rotateAPIKey() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentKey = (s.currentKey + 1) % len(s.apiKeys)
	if err := s.client.Close(); err != nil {
		return err
	}
	return s.initClient()
}

func (s *GeminiService) Chat(ctx context.Context, prompt string, messages []Message) (string, error) {
	fmt.Printf("Chat with prompt %s and history %v\n", prompt, messages)
	// Convert messages to chat history
	history := make([]*genai.Content, 0, len(messages))
	for _, msg := range messages {
		history = append(history, &genai.Content{
			Parts: []genai.Part{genai.Text(msg.Content)},
			Role:  string(msg.Role),
		})
	}
	// Start chat
	chat := s.model.StartChat()
	chat.History = history

	resp, err := chat.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		// Try rotating API key if there's an error
		if err := s.rotateAPIKey(); err != nil {
			return "", err
		}
		chat = s.model.StartChat()
		chat.History = history
		resp, err = chat.SendMessage(ctx, genai.Text(prompt))
		if err != nil {
			return "", err
		}
	}

	if len(resp.Candidates) == 0 {
		return "", errors.New("no response generated")
	}

	candidate := resp.Candidates[0]
	if funcs := candidate.FunctionCalls(); len(funcs) > 0 {
		resp, err = s.handleFunctionCall(ctx, chat, funcs)
		if err != nil {
			return "", err
		}

	}
	content := ""
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					content += string(text)
				}
			}

		}
	}

	return content, nil
}

func (s *GeminiService) handleFunctionCall(ctx context.Context, chat *genai.ChatSession, functions []genai.FunctionCall) (*genai.GenerateContentResponse, error) {
	fmt.Printf("Handle function call with functions %v\n", functions)
	funcResults := []genai.Part{}
	for _, function := range functions {
		handler, exists := s.functionsCall[function.Name]
		if !exists {
			return nil, fmt.Errorf("unknown function: %s", function.Name)
		}

		// Convert args to JSON bytes first
		argsBytes, err := json.Marshal(function.Args)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal function args: %v", err)
		}
		// Execute the function
		result, err := handler(ctx, argsBytes)
		if err != nil {
			return nil, fmt.Errorf("function execution failed: %v", err)
		}
		funcResults = append(funcResults, genai.FunctionResponse{
			Name:     function.Name,
			Response: map[string]any{"result": result},
		})
		fmt.Printf("Function %s executed with result %v\n", function.Name, result)
	}
	// Generate final response with function result
	resp, err := chat.SendMessage(
		ctx,
		funcResults...,
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Candidates) == 0 {
		return nil, errors.New("no response generated")
	}
	candidate := resp.Candidates[0]
	if funcs := candidate.FunctionCalls(); len(funcs) > 0 {
		return s.handleFunctionCall(ctx, chat, funcs)
	}

	return resp, nil
}

// Thêm method mới để hỗ trợ streaming
func (s *GeminiService) ChatStream(ctx context.Context, prompt string, handler StreamHandler) error {
	if handler == nil {
		handler = defaultStreamHandler
	}
	iter := s.model.GenerateContentStream(ctx, genai.Text(prompt))

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Nếu lỗi, thử rotate sang API key khác
			if err := s.rotateAPIKey(); err != nil {
				return err
			}
			// Thử lại với API key mới
			iter = s.model.GenerateContentStream(ctx, genai.Text(prompt))
			resp, err = iter.Next()
			if err != nil {
				return err
			}
		}

		if len(resp.Candidates) == 0 {
			continue
		}

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					handler(string(text))
				}
			}
		}
	}
	return nil
}

// RegisterFunction adds a new function to the model's capabilities
func (s *GeminiService) RegisterFunction(name, description string, parameters map[string]*genai.Schema, handler FunctionHandler) {
	functionDeclaration := &genai.FunctionDeclaration{
		Name:        name,
		Description: description,
		Parameters: &genai.Schema{
			Type:       genai.TypeObject,
			Properties: parameters,
			Required:   make([]string, 0, len(parameters)),
		},
	}

	// Add required parameters
	for paramName := range parameters {
		functionDeclaration.Parameters.Required = append(
			functionDeclaration.Parameters.Required,
			paramName,
		)
	}

	// Create the tool with the function declaration
	tool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{functionDeclaration},
	}

	// Set the tool and function call handler
	s.model.Tools = append(s.model.Tools, tool)
	s.functionsCall[name] = handler
}
