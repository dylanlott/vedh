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

func (r *queryResolver) Games(ctx context.Context, offset int, limit int) ([]*Game, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetGame(ctx context.Context, gameID string) (*Game, error) {
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

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *queryResolver) Boardstates(ctx context.Context, gameID string, userID *string) ([]*BoardState, error) {
	panic(fmt.Errorf("not implemented"))
}
