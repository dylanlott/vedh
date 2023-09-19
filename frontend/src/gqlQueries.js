import gql from 'graphql-tag';

// gameUpdateQuery powers the TurnTracker and Opponents components.
export const gameUpdatedSubscription = gql`
subscription($gameID: String!, $userID: String!) {
  gameUpdated(gameID: $gameID, userID: $userID) {
    ID
    Players {
      Username
      ID
      Boardstate {
        User
        Life
      }
    }
    Turn {
      Player
      Phase
      Number
    }
  }
} 
`

export const gameQuery = gql`
query ($limit: Int!, $offset: Int!) {
  games(limit: $limit, offset: $offset) {
    ID
    Players {
      Username
      ID
      Boardstate {
        User
        Life
      }
    }
    Turn {
      Player
      Phase
      Number
    }
  }
}
`

export const updateGame = gql`
mutation ($input: InputGame!) {
  updateGame(input: $input) {
    ID
    Players {
      ID
      Username
      Boardstate {
        User
        Life
      }
    }
    Turn {
      Phase
      Player
      Number
    }
  }
}
`

export const signup = gql`
mutation ($username: String!, $password: String!) {
  signup(username: $username, password: $password) {
    ID
    Username
    Token
  }
}
`

export const login = gql`
mutation($username: String!, $password: String!) {
  login(username: $username, password: $password) {
    Username
    ID
    Token 
  }
}`

export const commanderQuery = gql`
  query($name: String!) {
    search(name: $name) {
      Name
      ID
      Colors
      ColorIdentity
      CMC
      ManaCost
    }
  }
`

export const cardQuery = gql`
  query($name: String!) {
    card(name: $name) {
      Name
      ID
      Colors
      ColorIdentity
      CMC
      ManaCost
    }
  }
`