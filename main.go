package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	fmt.Println("What do you need counsel on today?")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	question := scanner.Text()

	c := openai.NewClient("your key here")
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an advisor. The user is seeking advice on a personal matter. Give helpful but frank advice.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	terminalWidth, _, err := getTerminalSize()
	if err != nil {
		fmt.Printf("\nError getting terminal size: %v\n", err)
		return
	}
	currentLineLength := 0
	var wordBuffer []rune
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			word := string(wordBuffer)
			if currentLineLength+len(word)+1 > terminalWidth-1 {
				fmt.Print("\n")
			}
			fmt.Print(string(wordBuffer), "\n")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		for _, r := range response.Choices[0].Delta.Content {
			if r == ' ' || r == '\n' {
				word := string(wordBuffer)
				if currentLineLength+len(word)+1 > terminalWidth-1 {
					fmt.Print("\n")
					currentLineLength = 0
				} else if r == '\n' {
					currentLineLength = 0
				} else {
					currentLineLength += 1
				}
				wordBuffer = append(wordBuffer, r)
				fmt.Print(string(wordBuffer))
				wordBuffer = wordBuffer[:0]
			} else {
				wordBuffer = append(wordBuffer, r)
				currentLineLength += 1
			}
		}
	}
}

func getTerminalSize() (int, int, error) {
	file, err := os.Open("/dev/tty")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var dimensions [4]uint16
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		file.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&dimensions)),
	)
	if errno != 0 {
		return 0, 0, errno
	}
	return int(dimensions[1]), 0, nil
}
