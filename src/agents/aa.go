package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"main/src/bus"
	"main/src/command"
)

type AAgent struct {
	logger  *slog.Logger
	bus     *OptimizedBus
	command *Command
}

type Response struct {
	ID           string `json:"id"`
	Object       string `json:"object"`
	CreatedAt    int    `json:"created_at"`
	Status       string `json:"status"`
	Background   bool   `json:"background"`
	Conversation struct {
		ID string `json:"id"`
	} `json:"conversation"`
	Error             any    `json:"error"`
	IncompleteDetails any    `json:"incomplete_details"`
	Instructions      any    `json:"instructions"`
	MaxOutputTokens   any    `json:"max_output_tokens"`
	MaxToolCalls      any    `json:"max_tool_calls"`
	Model             string `json:"model"`
	Output            []struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Summary []any  `json:"summary,omitempty"`
		Status  string `json:"status,omitempty"`
		Content []struct {
			Type        string `json:"type"`
			Annotations []any  `json:"annotations"`
			Logprobs    []any  `json:"logprobs"`
			Text        string `json:"text"`
		} `json:"content,omitempty"`
		Role string `json:"role,omitempty"`
	} `json:"output"`
	ParallelToolCalls  bool `json:"parallel_tool_calls"`
	PreviousResponseID any  `json:"previous_response_id"`
	PromptCacheKey     any  `json:"prompt_cache_key"`
	Reasoning          struct {
		Effort  string `json:"effort"`
		Summary any    `json:"summary"`
	} `json:"reasoning"`
	SafetyIdentifier any     `json:"safety_identifier"`
	ServiceTier      string  `json:"service_tier"`
	Store            bool    `json:"store"`
	Temperature      float64 `json:"temperature"`
	Text             struct {
		Format struct {
			Type string `json:"type"`
		} `json:"format"`
		Verbosity string `json:"verbosity"`
	} `json:"text"`
	ToolChoice  string  `json:"tool_choice"`
	Tools       []any   `json:"tools"`
	TopLogprobs int     `json:"top_logprobs"`
	TopP        float64 `json:"top_p"`
	Truncation  string  `json:"truncation"`
	Usage       struct {
		InputTokens        int `json:"input_tokens"`
		InputTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"input_tokens_details"`
		OutputTokens        int `json:"output_tokens"`
		OutputTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"output_tokens_details"`
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	User     any `json:"user"`
	Metadata struct {
	} `json:"metadata"`
}

type Payload struct {
	Model        string `json:"model"`
	Conversation string `json:"conversation"`
	Input        string `json:"input"`
}

func (p *Payload) GetBody() string {
	body, err := json.Marshal(Payload{
		Model:        p.Model,
		Conversation: p.Conversation,
		Input:        p.Input,
	})
	if err != nil {
		panic(err)
	}
	return string(body)
}

func (a *AAgent) Name() string { return "aa" }
func (a *AAgent) Request(body string) *Response {
	oai_url := os.Getenv("OPENAI_API_URL")
	oai_key := os.Getenv("OPENAI_API_KEY")

	client := &http.Client{}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/responses", oai_url),
		strings.NewReader(body),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oai_key))

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	response := &Response{}
	derr := json.NewDecoder(res.Body).Decode(response)
	if derr != nil {
		panic(derr)
	}

	if res.StatusCode != http.StatusOK {
		panic(res.Status)
	}

	return response
}
func (a *AAgent) Start(ctx context.Context) error {
	payload := Payload{
		Model:        "gpt-5",
		Conversation: "conv_68a78b95a9548197a40a1d7fce845d2503523f864ebccb71",
	}

	ch, unsub, err := a.bus.Subscribe(EvtMessage, 64)
	if err != nil {
		return err
	}
	defer unsub()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt, ok := <-ch:
			if !ok {
				return nil
			}
			msg, _ := evt.Data.(MessageModel)

			switch msg.Type {
			case System:
				//
			case Human:
				if ok, _ := a.command.IsCommand(msg.Text); !ok {

					payload.Input = msg.Text
					response := a.Request(payload.GetBody())

					message := MessageModel{
						Type: Assistant,
						From: a.Name(),
						Text: response.Output[len(response.Output)-1].Content[0].Text,
					}
					a.bus.Publish(EvtMessage, message)
				}
			case Assistant:
				//
			}
		}
	}
}
