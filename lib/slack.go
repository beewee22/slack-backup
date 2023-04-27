package lib

import (
	"context"
	"github.com/slack-go/slack"
	"golang.org/x/time/rate"
	"time"
)

func ReadAllMessageInChannel(ctx context.Context, slackClient *slack.Client, channelID string) (messageChan chan []slack.Message) {
	messageChan = make(chan []slack.Message)

	go func() {
		defer close(messageChan)

		limiter := rate.NewLimiter(rate.Every(time.Second), 1)
		hasMore := true
		cursor := ""
		for hasMore {
			err := limiter.Wait(ctx)
			if err != nil {
				panic(err)
			}
			history, err := slackClient.GetConversationHistory(&slack.GetConversationHistoryParameters{
				ChannelID: channelID,
				Cursor:    cursor,
			})
			if err != nil {
				panic(err)
			}

			hasMore = history.HasMore
			cursor = history.ResponseMetaData.NextCursor
			messageChan <- history.Messages
		}
	}()

	return messageChan
}

func FetchThreadInQueue(ctx context.Context, slackClient *slack.Client, channelID string, tsQueue *SafeQueue[string]) (threadChan chan []slack.Message) {
	threadChan = make(chan []slack.Message)

	go func() {
		defer close(threadChan)

		limiter := rate.NewLimiter(rate.Every(time.Second), 1)
		for {
			err := limiter.Wait(ctx)
			if err != nil {
				panic(err)
			}
			if tsQueue.IsEmpty() {
				if tsQueue.IsDone() {
					return
				}
				continue
			}

			// pop the first element
			ts := tsQueue.Pop()
			thread, _, _, err := slackClient.GetConversationReplies(&slack.GetConversationRepliesParameters{
				ChannelID: channelID,
				Timestamp: ts,
			})
			if err != nil {
				panic(err)
			}
			threadChan <- thread
		}
	}()

	return threadChan
}
