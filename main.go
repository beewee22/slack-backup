package main

import (
	"context"
	"fmt"
	"github.com/beewee22/slack-backup/lib"
	"github.com/slack-go/slack"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctx := context.Background()

	app := &cli.App{
		Name:  "slack-history",
		Usage: "Fetch and save Slack channel history as a JSON file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Slack bot token. input as an argument or set as an environment variable",
				EnvVars: []string{"SLACK_BACKUP_TOKEN"},
			},
		},
		Action: func(c *cli.Context) error {
			token := c.String("token")
			if token == "" {
				println(cli.ShowAppHelp(c))
				return fmt.Errorf("please provide a bot token")
			}

			channelID := c.Args().Get(0)
			if channelID == "" {
				println(cli.ShowAppHelp(c))
				return fmt.Errorf("please provide a channel name. ")
			}

			log.Printf("Start fetching channel history for %s\n", channelID)

			err := SaveChannelHistory(ctx, token, channelID)
			if err != nil {
				return err
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func SaveChannelHistory(ctx context.Context, token string, channelID string) error {
	slackClient := slack.New(token)

	// check if the token is valid
	_, err := slackClient.AuthTest()
	log.Printf("Test auth: ")
	if err != nil {
		log.Fatal("\nAuthTest error: ", err)
		return err
	}
	log.Printf("OK\n")

	// check if the channel exists
	log.Printf("Check channel is accessable: ")
	_, err = slackClient.GetConversationInfo(&slack.GetConversationInfoInput{ChannelID: channelID})
	if err != nil {
		log.Fatal("\nGetChannelInfo error: ", err)
		return err
	}
	log.Printf("OK\n")

	// join the channel
	log.Printf("Joining channel: ")
	_, _, _, err = slackClient.JoinConversation(channelID)
	if err != nil {
		log.Fatal("\nJoinChannel error: ", err)
		return err
	}
	log.Printf("OK\n")

	log.Printf("Start fetching messages and threads in channel %s\n", channelID)
	messages, threads := ReadAllMessageAndThreadsInChannel(ctx, slackClient, channelID)

	// save the messages to a JSON file
	errOnSaveMsg := lib.SaveMessagesAsJSONFile(messages, "messages.json")
	errOnSaveThread := lib.SaveMessagesAsJSONFile(threads, "threads.json")

	if errOnSaveMsg != nil {
		return errOnSaveMsg
	}
	if errOnSaveThread != nil {
		return errOnSaveThread
	}

	return nil
}

func ReadAllMessageAndThreadsInChannel(ctx context.Context, slackClient *slack.Client, channelID string) ([]slack.Message, []slack.Message) {
	// read all messages in the channel
	var messages []slack.Message
	var threads []slack.Message
	var threadTsQueue *lib.SafeQueue[string]
	threadTsQueue = lib.NewSafeQueue[string]()

	messageChan := lib.ReadAllMessageInChannel(ctx, slackClient, channelID)
	threadChan := lib.FetchThreadInQueue(ctx, slackClient, channelID, threadTsQueue)

	for receivedMessages := range messageChan {
		log.Printf("Received %d messages\n", len(receivedMessages))
		messages = append(messages, receivedMessages...)
		for _, msg := range receivedMessages {
			if msg.ThreadTimestamp != "" {
				threadTsQueue.Add(msg.ThreadTimestamp)
			}
		}
		log.Printf("Added %d threads to queue\n", threadTsQueue.Len())
	}
	log.Printf("Done reading all messages\n")
	threadQueueMaxLen := threadTsQueue.Len()
	threadTsQueue.SetDone(true)

	for receivedThreads := range threadChan {
		log.Printf("Received %d threads\n", len(receivedThreads))
		log.Printf("%d/%d thread reply requests are done.", threadTsQueue.Len(), threadQueueMaxLen)
		threads = append(threads, receivedThreads...)
	}

	return messages, threads
}
