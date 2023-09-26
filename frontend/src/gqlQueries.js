import gql from 'graphql-tag';

// gameUpdateQuery powers the TurnTracker and Opponents components.
export const gameUpdatedSubscription = gql`subscription($gameID: String!, $userID: String!) {
  gameUpdated(gameID: $gameID, userID: $userID) {
    ID
    Players {
      Username
      Boardstate {
        GameID
        UserID
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

export const getGameQuery = gql`query($gameID: String!){
	getGame(gameID: $gameID) {
    ID
    Players {
      Username
      ID
      Boardstate {
        UserID
        User
        Life
        GameID
      }
    }
  }
}`

export const gameQuery = gql`
query gameQuery($limit: Int!, $offset: Int!) {
  games(limit: $limit, offset: $offset) {
    ID
  }
}
`

export const updateGame = gql`
mutation updateGame($input: InputGame!) {
  updateGame(input: $input) {
    ID
    Username
    UserID
    Players {
      ID
      UserID
      Username
      Boardstate {
        User
        UserID
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
mutation signup($username: String!, $password: String!) {
  signup(username: $username, password: $password) {
    ID
    Username
    Token
  }
}
`

export const login = gql`
mutation login($username: String!, $password: String!) {
  login(username: $username, password: $password) {
    Username
    ID
    Token 
  }
}`

export const commanderQuery = gql`
  query commanderQuery($name: String!) {
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
  query cardQuery($name: String!) {
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