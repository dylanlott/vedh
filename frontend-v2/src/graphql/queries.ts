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
      Players { ID Username }
    }
  }
`;

export const GET_GAME_QUERY = gql`
  query GetGame($gameID: String!) {
    getGame(gameID: $gameID) {
      ID
      CreatedAt
      Players {
        ID
        Username
        Boardstate {
          Life
          Commander { ID Name Tapped }
          Battlefield { ID Name Types Tapped }
          Hand { ID Name Types Tapped }
          Graveyard { ID Name Tapped }
          Exiled { ID Name Tapped }
          Revealed { ID Name Tapped }
          Library { ID Name Tapped }
          Controlled { ID Name Tapped }
        }
      }
      Stack { ID Name CurrentZone }
      Turn { Player Phase Number Priority }
      Rules { Name Value }
    }
  }
`;

export const GAME_UPDATED_SUBSCRIPTION = gql`
  subscription GameUpdated($gameID: String!, $userID: String!) {
    gameUpdated(gameID: $gameID, userID: $userID) {
      ID
      CreatedAt
      Turn { Player Phase Number Priority }
      Stack { ID Name CurrentZone }
      Rules { Name Value }
      Players {
        ID
        Username
        Boardstate {
          Life
          Battlefield { ID Name Types Tapped }
          Hand { ID Name Types Tapped }
          Commander { ID Name Tapped }
          Graveyard { ID Name Tapped }
          Exiled { ID Name Tapped }
          Revealed { ID Name Tapped }
          Library { ID Name Tapped }
          Controlled { ID Name Tapped }
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
      Text
    }
  }
`;
