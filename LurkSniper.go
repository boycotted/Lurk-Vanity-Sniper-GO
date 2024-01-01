package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

const (
	token      = ""  // Your Discord token
	webhookURL = ""  // Your Discord webhook link
	guildID    = "123" // Your Guild ID
)

var (
	claimed bool
	mu      sync.Mutex
	vanities = []string{"ow", "smh", "jr", "care"} // Add more vanity URLs as needed
)

var lurkBanner = `
  
  ███         ██                         ██         ███
  ████        █████       ██     █████   ███       ████
   ████       █████       ████   ███████ ████     ███
    ███        ████      █████  ███  ███  ████   ████
    ▄██        ████      ████   ███   ██   ███   ███
     ██        ████     █████  ████   ██    ██  ███
    ███        ████    █████   █████████    █████████
  █████        ████   █████    ████████      █████████
 █████         ███████████     ██████        ████   ████
 ██████         ████████       ███ ███       ███     ████
 ████████████   ███████        ███  ████     ███       ████
 ███ ███████████               ██    ████    ███        █████
           █████                      ████    ██         ████
`

func notifyStart(vanities []string) {
	message := fmt.Sprintf("Vanity Sniper Started\nTarget Vanities: %s", vanitiesString(vanities))
	sendWebhook(message)
}

func notifyVanityClaimed(vanityCode string) {
	message := fmt.Sprintf("Vanity Claimed\nVanity Code: %s\nGuild ID: %s", vanityCode, guildID)
	sendWebhook(message)
}

func sendWebhook(message string) {
	payload := []byte(fmt.Sprintf(`{"content": "@everyone", "embeds": [{"title": "Vanity Sniper", "description": "%s", "color": 3060215, "footer": {"text": "Vanity sniper provided by @facilitated"}}]}`, message))
	statusCode, body, err := fasthttp.Post(nil, webhookURL, payload)
	if err != nil {
		log.Printf("Error sending webhook: %s", err)
		return
	}

	log.Printf("Webhook status code: %d, Response body: %s", statusCode, body)
}

func claimVanity(vanityCode string) {
	mu.Lock()
	defer mu.Unlock()

	if claimed {
		return
	}

	claimed = true
	url := fmt.Sprintf("https://canary.discord.com/api/v9/guilds/%s/vanity-url", guildID)
	json := []byte(fmt.Sprintf(`{"code": "%s"}`, vanityCode))

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod("PATCH")
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("X-Audit-Log-Reason", "slapped by console")
	req.Header.Set("Content-Type", "application/json")
	req.SetBody(json)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	client := &fasthttp.Client{}
	if err := client.Do(req, resp); err != nil {
		log.Printf("Error claiming vanity: %s", err)
		return
	}

	statusCode := resp.StatusCode()
	log.Printf("Vanity claimed: %s, Status code: %d", vanityCode, statusCode)

	if statusCode == fasthttp.StatusOK || statusCode == fasthttp.StatusCreated {
		notifyVanityClaimed(vanityCode)
	} else {
		log.Printf("Failed to claim vanity: %s, Status code: %d", vanityCode, statusCode)
	}
}

func fetchVanity(vanityCode string, attempt int) {
	url := "https://canary.discord.com/api/v10/invites/" + vanityCode
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.Set("Authorization", "Bot "+token)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	client := &fasthttp.Client{}

	if err := client.Do(req, resp); err != nil {
		log.Printf("Error fetching vanity %s (Attempt: %d): %s", vanityCode, attempt, err)
		return
	}

	statusCode := resp.StatusCode()
	if statusCode == fasthttp.StatusNotFound {
		claimVanity(vanityCode)
	} else if statusCode == fasthttp.StatusOK {
		log.Printf("Attempt: %d | Vanity: %s", attempt, vanityCode)
	} else if statusCode == fasthttp.StatusTooManyRequests {
		log.Println("Rate Limited")
		time.Sleep(1 * time.Second) // Adjust the sleep duration
	} else {
		log.Printf("Unknown Error: Vanity %s, Status code: %d", vanityCode, statusCode)
	}
}

func threadExecutor(vanityCode string, attempt int) {
	for !claimed {
		fetchVanity(vanityCode, attempt)
		break
	}
}

func vanitiesString(vanities []string) string {
	return "vanities: " + fmt.Sprintf("%q", vanities)
}

func main() {
	color.Set(color.FgWhite, color.BgBlack)
	fmt.Println(lurkBanner)
	color.Unset()

	fmt.Println("Starting...")

	notifyStart(vanities)

	for _, vanity := range vanities {
		for attempt := 0; attempt < 100000000; attempt++ {
			if claimed {
				break
			}
			go threadExecutor(vanity, attempt)
			time.Sleep(100 * time.Microsecond) // Adjust the sleep duration
		}
	}

	fmt.Println("Execution Completed")
	select {}
}

// Implement your input function according to your needs
func input() string {
	var input string
	fmt.Scanln(&input)
	return input
}
