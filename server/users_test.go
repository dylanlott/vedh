package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func uniqueUsername(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func Test_graphQLServer_Signup(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *User
		wantErr bool
	}{
		{
			name: "should create a user and persist it to the appDB.",
			args: args{
				username: "shakezula",
				password: "ohhellyeah",
			},
			want: &User{
				Username: "shakezula",
			},
			wantErr: false,
		},
		{
			name: "should return an error if no password is provided",
			args: args{
				username: "shakezula",
				password: "",
			},
			wantErr: true,
		},
		{
			name: "should return a friendly error when username already exists",
			args: args{
				username: "shakezula",
				password: "ohhellyeah",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			username := tt.args.username
			switch tt.name {
			case "should create a user and persist it to the appDB.":
				username = uniqueUsername("user")
			case "should return a friendly error when username already exists":
				username = uniqueUsername("dupe")
				_, err := s.Signup(context.Background(), username, "seedpassword")
				if err != nil {
					t.Fatalf("failed to seed duplicate username: %v", err)
				}
			}
			got, err := s.Signup(context.Background(), username, tt.args.password)
			if tt.want != nil && !tt.wantErr {
				tt.want.Username = username
			}
			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(User{}, "ID", "Token")); diff != "" {
				t.Errorf("graphQLServer.Signup: wanted: %+v - got: %+v", tt.want, got)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.Signup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.name == "should return a friendly error when username already exists" {
				if err == nil || err.Error() != "That username is already taken. Try another one." {
					t.Fatalf("expected friendly duplicate username error, got %v", err)
				}
				return
			}
			if err == nil {
				if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(User{}, "ID", "Token")); diff != "" {
					t.Errorf("failed to get correct user back")
				}
				if got.ID == "" {
					t.Errorf("failed to set UUID on user")
				}
				if got.Token == nil || *got.Token == "" {
					t.Errorf("failed to set signup token")
				}
			}
		})
	}
}

func Test_graphQLServer_Login(t *testing.T) {
	type args struct {
		ctx      context.Context
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *User
		wantErr bool
	}{
		{
			name: "should log in a user and give them a token",
			args: args{
				ctx:      context.Background(),
				username: "shakezula",
				password: "testpassword",
			},
			want: &User{
				Username: "shakezula",
			},
			wantErr: false,
		},
		{
			name: "should return an error when user not found",
			args: args{
				ctx:      context.Background(),
				username: "notfound",
				password: "testpassword",
			},
			wantErr: true,
		},
		{
			name: "should return an error if password invalid",
			args: args{
				ctx:      context.Background(),
				username: "shakezula",
				password: "invalidpass",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			createdUsername := uniqueUsername("user")
			if tt.name != "should return an error when user not found" {
				_, err := s.Signup(tt.args.ctx, createdUsername, "testpassword")
				if err != nil {
					t.Errorf("failed to create user for login tests: %s", err)
				}
			}
			username := tt.args.username
			if tt.name != "should return an error when user not found" {
				username = createdUsername
			} else {
				username = uniqueUsername("missing")
			}
			got, err := s.Login(tt.args.ctx, username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got == nil {
					t.Errorf("failed to return user")
					return
				}
				if got.ID == "" {
					t.Errorf("failed to set ID")
				}
				if got.Password != nil {
					t.Errorf("should not receive password back from login")
				}
				if got.Token == nil {
					t.Errorf("failed to set token")
				}
				if got.Username != createdUsername {
					t.Errorf("failed to set correct username")
				}
			}
		})
	}
}
