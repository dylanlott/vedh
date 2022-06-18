package server

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
)

func (r *mutationResolver) Signup(ctx context.Context, username string, password string) (*User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) Login(ctx context.Context, username string, password string) (*User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateGame(ctx context.Context, input InputCreateGame) (*Game, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) JoinGame(ctx context.Context, input *InputJoinGame) (*Game, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateGame(ctx context.Context, input InputGame) (*Game, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateBoardState(ctx context.Context, input InputBoardState) (*BoardState, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Users(ctx context.Context, userID *string) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Games(ctx context.Context, gameID *string) ([]*Game, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Boardstates(ctx context.Context, gameID string, userID *string) ([]*BoardState, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Card(ctx context.Context, name string, id *string) (*Card, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Cards(ctx context.Context, list []string) ([]*Card, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Search(ctx context.Context, name *string, colors []*string, colorIdentity []*string, keywords []*string) ([]*Card, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) GameUpdated(ctx context.Context, gameID string, userID string) (<-chan *Game, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) BoardstateUpdated(ctx context.Context, observerID string, userID string) (<-chan *BoardState, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
