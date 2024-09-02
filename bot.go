package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Message struct {
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateNickname() string {
	firstNames := []string{"Cool", "Super", "Mega", "Ultra", "Epic", "Lucky", "Silent", "Mighty"}
	lastNames := []string{"Hero", "Warrior", "Ninja", "Samurai", "Wizard", "Knight", "Ranger", "Hunter"}

	return fmt.Sprintf("%s%s%d", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))], rand.Intn(1000))
}

func sendMessage(url string) {
	message := Message{
		Nickname: generateNickname(),
		Text:     generateRandomString(32),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("JSON serialization error: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Sending error: %v", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	fmt.Printf("Server response: %s\n", resp.Status)
}

func main() {
	messagePerMinute := os.Getenv("BOT_MESSAGES_PER_MINUTE")
	rate, err := strconv.Atoi(messagePerMinute)
	if err != nil {
		rate = 60
	}
	url := os.Getenv("BOT_SERVICE_URL")
	if url == "" {
		url = "http://app:3000/api/messages/"
	}
	flag.Parse()

	interval := time.Minute / time.Duration(rate)

	for {
		sendMessage(url)
		time.Sleep(interval)
	}
}
