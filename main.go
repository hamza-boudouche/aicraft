package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/joho/godotenv"
	"github.com/ktr0731/go-fuzzyfinder"
)

var path = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key="

type GeminiResp struct {
	Candidates []Candidates `json:"candidates"`
}

type Candidates struct {
	Content Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

var words = []string{
	"water",
	"earth",
	"air",
	"fire",
}

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("failed to load .env")
	}
	path = fmt.Sprintf("%s%s", path, os.Getenv("GEMINI_API_KEY"))
}

func main() {
	gameLoop()
}

func gameLoop() {
	currentHeader := ""
	for {
		first, err := pickWord(currentHeader)
		if err != nil {
			panic(err)
		}
		second, err := pickWord(first)
		if err != nil {
			panic(err)
		}
		res, err := combine(first, second)
		if err != nil {
			panic(err)
		}
		if slices.Contains(words, res) {
			currentHeader = fmt.Sprintf("You already have: %s", res)
		} else {
			currentHeader = fmt.Sprintf("You got: %s !!", res)
			words = append(words, res)
		}
	}
}

func pickWord(current string) (string, error) {
	idx, err := fuzzyfinder.Find(
		words,
		func(i int) string {
			return words[i]
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Word: %s", words[i])
		}),
		fuzzyfinder.WithHeader(current))
	fuzzyfinder.WithSelectOne()
	if err != nil {
		return "", err
	}
	return words[idx], nil
}

func combine(first, second string) (string, error) {
	prompt := "act as if you are a game that combines 2 expressions to form new expressions related to the both of the input expressions, you should ONLY output the resulting expression, no explanation or anything except the final expression result, your inputs are: %s + %s"
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{
						"text": fmt.Sprintf(prompt, first, second),
					},
				},
			},
		},
	}
	stringPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to format payload: %s", err)
	}
	resp, err := http.Post(path, "application/json", bytes.NewReader(stringPayload))
	if err != nil {
		return "", fmt.Errorf("failed to send POST request to Gemini: %s", err)
	}
	var geminiResp GeminiResp
	respSlice, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read gemini response %s", err)
	}
	json.Unmarshal(respSlice, &geminiResp)
	fmt.Println(geminiResp)
	fmt.Println(string(respSlice))

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
