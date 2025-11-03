import { gql } from '@apollo/client/core';

export const CURRENT_USER_QUERY = gql`
  query CurrentUser {
    users {
      Username
    }
  }
`;

export const GAMES_QUERY = gql`
  query Games($offset: Int!, $limit: Int!) {
    games(offset: $offset, limit: $limit) {
      ID
      Players { Username }
    }
  }
`;

export const GET_GAME_QUERY = gql`
  query GetGame($gameID: String!) {
    getGame(gameID: $gameID) {
      ID
      Players {
        ID
        Username
        Boardstate {
          Life
          Commander {
            ID
            Name
          }
          Battlefield {
            ID
            Name
            Types
          }
          Hand {
            ID
            Name
          }
          Graveyard { ID Name }
          Exiled { ID Name }
          Revealed { ID Name }
          Library { ID Name }
          Controlled { ID Name }
        }
      }
      Stack {
        ID
        Name
      }
      Turn {
        Player
        Phase
        Number
      }
      Rules {
        Name
        Value
      }
    }
  }
`;

export const GAME_UPDATED_SUBSCRIPTION = gql`
  subscription GameUpdated($gameID: String!, $userID: String!) {
    gameUpdated(gameID: $gameID, userID: $userID) {
      ID
      Turn {
        Player
        Phase
        Number
      }
      Stack {
        ID
        Name
      }
      Players {
        ID
        Username
        Boardstate {
          Life
          Battlefield {
            ID
            Name
            Types
          }
          Hand {
            ID
            Name
          }
          Commander { ID Name }
          Graveyard { ID Name }
          Exiled { ID Name }
          Revealed { ID Name }
          Library { ID Name }
          Controlled { ID Name }
        }
      }
    }
  }
`;

export const SEARCH_CARDS_QUERY = gql`
  query SearchCards($name: String) {
    search(name: $name) {
      ID
      Name
    }
  }
`;
