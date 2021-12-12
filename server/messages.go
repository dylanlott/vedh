package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/ksuid"
)

// PostMessage ...
func (s *graphQLServer) PostMessage(ctx context.Context, user string, text string) (*Message, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	// Create message
	m := &Message{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		Text:      text,
		User:      user,
	}
	mj, _ := json.Marshal(m)

	// TODO: Update messages to key off of `message:<game_id>` and `message:<room_id>`
	if err := s.rc.LPush("messages", mj).Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	// Notify new message
	s.mutex.Lock()
	for _, ch := range s.messageChannels {
		ch <- m
	}
	s.mutex.Unlock()
	return m, nil
}

// Messages ...
func (s *graphQLServer) Messages(ctx context.Context) ([]*Message, error) {
	cmd := s.rc.LRange("messages", 0, -1)
	if cmd.Err() != nil {
		log.Println(cmd.Err())
		return nil, cmd.Err()
	}
	res, err := cmd.Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	messages := []*Message{}
	for _, mj := range res {
		m := &Message{}
		err = json.Unmarshal([]byte(mj), &m)
		if err != nil {
			log.Printf("error unmarhsaling json Message: %s", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// MessagePosted ...
func (s *graphQLServer) MessagePosted(ctx context.Context, user string) (<-chan *Message, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	fmt.Printf("\n received messagePosted: %+v\n", user)

	// Create new channel for request
	messages := make(chan *Message, 1)
	s.mutex.Lock()
	s.messageChannels[user] = messages
	s.mutex.Unlock()

	// Delete channel when done
	go func() {
		<-ctx.Done()
		s.mutex.Lock()
		delete(s.messageChannels, user)
		s.mutex.Unlock()
	}()

	return messages, nil
}
