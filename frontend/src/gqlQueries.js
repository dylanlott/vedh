import gql from 'graphql-tag';

export const cardFragment = gql`
fragment CardFields on Card {
  __typename
  ID
  Name
  Quantity
  Tapped
  Flipped
  Counters {
    Name
    Value
  }
  Colors
  ColorIdentity
  CMC
  FaceName
  FaceManaValue
  FaceConvertedManaCost
  ManaCost
  UUID
  Power
  Toughness
  Types
  Subtypes
  Supertypes
  Text
  TCGID
  ScryfallID
}
`

// gameUpdateQuery powers the TurnTracker and Opponents components.
export const gameUpdatedSubscription = gql`subscription($gameID: String!, $userID: String!) {
  gameUpdated(gameID: $gameID, userID: $userID) {
    ID
    Rules {
      Name
      Value
    }
    Turn {
      Player
      Phase
      Number
    }
    Players {
      Username
      ID
      Boardstate {
      	User
        UserID
        Life
        GameID
        Library {
          ...CardFields
        }
        Exiled {
          ...CardFields
        }
        Graveyard {
          ...CardFields
        }
        Revealed {
          ...CardFields
        }
        Hand {
          ...CardFields
        }
        Commander {
          ...CardFields
        }
        Field {
          ...CardFields
        }
        Controlled {
          ...CardFields
        }
        Counters {
          Name
          Value
        }
      }
    }
  }
}${cardFragment}
`

export const getGameQuery = gql`query($gameID: String!){
	getGame(gameID: $gameID) {
    ID
    Players {
      Username
      ID
      Boardstate {
        User
        UserID
        Life
        GameID
        Library {
          ...CardFields
        }
        Exiled {
          ...CardFields
        }
        Graveyard {
          ...CardFields
        }
        Revealed {
          ...CardFields
        }
        Hand {
          ...CardFields
        }
        Commander {
          ...CardFields
        }
        Field {
          ...CardFields
        }
        Controlled {
          ...CardFields
        }
        Counters {
          Name
          Value
        }
      }
    }
  }
}${cardFragment}`

export const gameQuery = gql`
query gameQuery($limit: Int!, $offset: Int!) {
  games(limit: $limit, offset: $offset) {
    ID
    Players {
      Username
      ID
      Boardstate {
        Username
        UserID
        Boardstate {
          User
          UserID
          Life
          GameID
          Library {
            ...CardFields
          }
          Exiled {
            ...CardFields
          }
          Graveyard {
            ...CardFields
          }
          Revealed {
            ...CardFields
          }
          Hand {
            ...CardFields
          }
          Commander {
            ...CardFields
          }
          Field {
            ...CardFields
          }
          Controlled {
            ...CardFields
          }
          Counters {
            Name
            Value
          }
        }
      }
    }
  }
}${cardFragment}
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
        GameID
        Library {
          ...CardFields
        }
        Exiled {
          ...CardFields
        }
        Graveyard {
          ...CardFields
        }
        Revealed {
          ...CardFields
        }
        Hand {
          ...CardFields
        }
        Commander {
          ...CardFields
        }
        Field {
          ...CardFields
        }
        Controlled {
          ...CardFields
        }
        Counters {
          Name
          Value
        }
      }
    }
    Turn {
      Phase
      Player
      Number
    }
  }
}${cardFragment}
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