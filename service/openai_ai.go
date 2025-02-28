package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/tieubaoca/chatbot-be/database"
	"github.com/tieubaoca/chatbot-be/types"
)

var (
	SystemMessageInitiateMechanicalEngineer = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a technical assistant. You can answer questions about technical engineering. If you do not know the answer, you can research it the database before responding. You answer questions by Vietnamese and only Vietnamese.",
	}
)

type OpenAIService struct {
	client        *openai.Client
	weaviateDb    *database.WeaviateStore
	functionsCall map[string]types.FunctionHandler
	tools         []openai.Tool
	model         string
}

func NewOpenAIService(baseURL string, apiKey, model string, weaviateDb *database.WeaviateStore) *OpenAIService {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL // Set this to your local LLM server URL
	client := openai.NewClientWithConfig(config)
	return &OpenAIService{
		client:        client,
		functionsCall: make(map[string]types.FunctionHandler),
		tools:         make([]openai.Tool, 0),
		model:         model,
		weaviateDb:    weaviateDb,
	}
}

func (s *OpenAIService) Chat(ctx context.Context, messages []types.Message) (*types.Message, error) {
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
			Model:    s.model,
		},
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response generated")
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		if resp.Choices[0].Message.ToolCalls[0].Function.Name == "retrieve_augmented_graph" {
			// remove last message
			resp, err = s.retrieveDocument(ctx, openaiMessages, resp.Choices[0].Message.ToolCalls[0].Function.Arguments)
			if err != nil {
				return nil, err
			}
		} else {
			resp, err = s.handleFunctionCall(ctx, openaiMessages, resp)
			if err != nil {
				return nil, err
			}
		}

	}

	// Convert response back to our Message type
	return &types.Message{
		Role:    "assistant",
		Content: resp.Choices[0].Message.Content,
	}, nil
}

func (s *OpenAIService) ChatStream(ctx context.Context, messages []types.Message, streamHandler types.StreamHandler) error {
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
			Model:    s.model,
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

func (s *OpenAIService) RegisterFunctionCall(name, description string, params jsonschema.Definition, handler types.FunctionHandler) error {
	if name == "retrieve_augmented_graph" {
		return errors.New("function name retrieve_augmented_graph is reserved")
	}
	if s.functionsCall == nil {
		s.functionsCall = make(map[string]types.FunctionHandler)
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
	return nil
}

func (s *OpenAIService) RegisterRAGFunctionCall() error {
	f := openai.FunctionDefinition{
		Name:        "retrieve_augmented_graph",
		Description: "Retrieve the augmented graph of the documents, use the document as context to answer the question",
		Parameters: jsonschema.Definition{
			Type:        "object",
			Description: "Retrieve the augmented graph of the document",
			Properties: map[string]jsonschema.Definition{
				"queries": {
					Type:        "array",
					Description: "List of queries to retrieve the document and use as context",
				},
				"question": {
					Type:        "string",
					Description: "The question of the user",
				},
			},
		},
	}
	t := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f,
	}
	s.tools = append(s.tools, t)
	return nil
}

func (s *OpenAIService) retrieveDocument(ctx context.Context, openaiMessages []openai.ChatCompletionMessage, args string) (openai.ChatCompletionResponse, error) {
	// var question string
	// var queries []string
	openaiMessages = openaiMessages[:len(openaiMessages)-1]

	type RetrieveDocumentArgs struct {
		Queries  []string `json:"queries"`
		Question string   `json:"question"`
	}

	var retrieveDocumentArgs RetrieveDocumentArgs
	if err := json.Unmarshal([]byte(args), &retrieveDocumentArgs); err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	question := retrieveDocumentArgs.Question
	queries := retrieveDocumentArgs.Queries

	docs, _, err := s.weaviateDb.SearchSimilar(ctx, queries, 5)
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	jsonDocs, err := json.Marshal(docs)
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	var prompt string
	if len(docs) == 0 {
		prompt = createRetrieveDocumentPrompt("No documents found", question)
	} else {
		prompt = createRetrieveDocumentPrompt(string(jsonDocs), question)
	}
	openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Messages: openaiMessages,
			Tools:    s.tools,
			Model:    s.model,
		})
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	if len(resp.Choices) == 0 {
		return openai.ChatCompletionResponse{}, errors.New("no response generated")
	}
	return resp, nil
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
			Model:    s.model,
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

func createRetrieveDocumentPrompt(ctx, question string) string {
	return fmt.Sprintf(`
Use the following CONTEXT to answer the QUESTION at the end.
If you don't know the answer, just say that you don't know, don't try to make up an answer.
Use an unbiased and journalistic tone.

CONTEXT: %s

QUESTION: %s`, ctx, question)
}
