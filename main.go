package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/likexian/whois"
	openai "github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

func domainAvailableTimeout(domain string) bool {

	var result chan bool = make(chan bool)

	go func() {
		result <- domainAvailable(domain)
	}()
	select {
	case <-time.After(4 * time.Second):
		// fmt.Println("timed out")
		return true
	case result := <-result:
		// fmt.Println(result)
		return result
	}

}

func getChatgptResponse(req string, openaiKey string) string {

	client := openai.NewClient(openaiKey)
	messages := make([]openai.ChatCompletionMessage, 0)

	for {
		// convert CRLF to LF
		req = strings.Replace(req, "\n", "", -1)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: req,
		})

		stream, err := client.CreateChatCompletionStream(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		var oldMessage string

		if err != nil {
			fmt.Printf("ChatCompletionStream error: %v\n", err)
			return "error"
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println()

				// if the message has a newline, terminate everything
				if strings.Contains(oldMessage, "\n") == true {
					break
				}

				content := oldMessage

				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: content,
				})
				break
			}

			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				break
			}

			if response.Choices[0].Delta.Content != "" {
				oldMessage += response.Choices[0].Delta.Content
			}

			fmt.Printf(response.Choices[0].Delta.Content)

		}

		return oldMessage

	}
}

func domainAvailable(domain string) bool {

	who, err := whois.Whois(domain)

	if err != nil {
		fmt.Println(err)
		if strings.Contains(err.Error(), "timeout") == true || strings.Contains(err.Error(), "no whois server found for domain") == true {
			return true
		}

		return false
	}

	// if the tld is il
	if strings.Contains(domain, ".il") == true {
		return !strings.Contains(who, "status")
	}

	errorValues := []string{
		"but this server does not have",
		"No Object Found",
		"no whois server found for domain:",
		"No match for",
		"No entries found for the selected source(s)",
		"Domain not found",
		"No Data Found",
		"No entries found",
	}

	// check if who contains any of the error values
	for i := range errorValues {
		if strings.Contains(who, errorValues[i]) == true {
			return true
		}
	}

	return false

}

func startSearch(tlds string, key string, name string, idea bool) {

	prompt := "return a list of domain hacks that could be used with the project name " + name + ", priorities it being a cool domain rather than sticking to the project name, you can switch it up. Cool domain hack examples : [wholeso.me, moji.to, e.xplo.it, delicio.us] Reminder that a domain has to be at least 3 letters. \n Here are a list of tlds you can use as the domain name, \n " + tlds + "in a js array format, DO NOT OUTPUT ANYTHING BUT THE LIST SEPERATED BY SPACES. Try to generate at least 15 domains. output example for name \"njalla\" : njal.la njalla.com njalla.net"

	if idea == true {
		prompt = "return a list of domain hacks that could be used with the project idea " + name + ", priorities it being a cool domain rather than sticking to the project name, you can switch it up, but make sure the idea is still conveyed in the domain. Cool domain hack examples : [wholeso.me, moji.to, e.xplo.it, delicio.us] Reminder that a domain has to be at least 3 letters. \n Here are a list of tlds you can use as the domain name, \n " + tlds + "in a js array format, DO NOT OUTPUT ANYTHING BUT THE LIST SEPERATED BY SPACES. Try to generate at least 15 domains. output example for name \"njalla\" : njal.la njalla.com njalla.net"
	}

	domains := getChatgptResponse(prompt, key)

	domainsJSON := strings.Split(domains, " ")

	tldList := strings.Split(tlds, "\n")

	checkedDomains := make([]string, 0)

	for i := range domainsJSON {

		// check if the domain has a tld

		if strings.Contains(domainsJSON[i], ".") == false {
			continue
		}

		tld := strings.Split(domainsJSON[i], ".")[1]

		for j := range tldList {
			if strings.ToLower(tldList[j]) == strings.ToLower(tld) {
				fmt.Println("Valid Domain : ", domainsJSON[i])

				checkedDomains = append(checkedDomains, domainsJSON[i])
				break
			}
		}
	}

	fmt.Println(checkedDomains)

	// if the length of the checked domains is 0, recurse
	if len(checkedDomains) == 0 {
		startSearch(tlds, key, name, idea)
	}

	var goodDomains []string

	var wg sync.WaitGroup

	for i := range checkedDomains {

		// if domainAvailable(checkedDomains[i]) == true {
		// 	fmt.Println(checkedDomains[i], "is available")
		// } else {
		// 	fmt.Println(checkedDomains[i], "is not available")
		// }

		wg.Add(1)

		go func(i int) {

			defer wg.Done()

			if domainAvailableTimeout(checkedDomains[i]) == true {
				fmt.Println(checkedDomains[i], "is available")

				goodDomains = append(goodDomains, checkedDomains[i])
			} else {
				fmt.Println(checkedDomains[i], "is not available")
			}
		}(i)
	}

	wg.Wait()

	fmt.Println(goodDomains)

	if len(goodDomains) == 0 {
		fmt.Println("No domains found, retrying")

		name = name + "Do not use the domains " + strings.Join(checkedDomains, " ") + " as they are not available"

		startSearch(tlds, key, name, idea)
	}

}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}

	var key string = os.Getenv("OPENAI")

	if key == "" {
		log.Fatal("Error loading openai key from .env file, did you add OPENAI={YOUR_KEY} into the .env file ?")
	}

	fileContent, err := ioutil.ReadFile("tlds.txt")
	if err != nil {
		log.Fatal(err)
	}

	tlds := string(fileContent)

	app := &cli.App{
		Name:  "domainGPT",
		Usage: "Generate domains using the power of AI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-verify",
				Usage: "do not verify domain names before returning them",
				Aliases: []string{
					"sv",
					"s",
				},
				Value: false,
			},
		},
		DefaultCommand: "help",

		Commands: []*cli.Command{
			{
				Name:  "name",
				Usage: "generate domains based on an existing name",
				Action: func(c *cli.Context) error {
					// get arg
					name := c.Args().Get(0)

					fmt.Println(name)

					// check if the name is empty
					if name == "" {
						fmt.Println("Please enter a name")
						return nil
					}

					startSearch(tlds, key, name, false)

					return nil
				},
			},
			{
				Name:  "idea",
				Usage: "generate domains based on a project idea",
				Action: func(c *cli.Context) error {
					// get arg
					project := c.Args().Get(0)

					fmt.Println(project)

					// check if the name is empty
					if project == "" {
						fmt.Println("Please enter a project idea")
						return nil
					}

					startSearch(tlds, key, project, true)

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
