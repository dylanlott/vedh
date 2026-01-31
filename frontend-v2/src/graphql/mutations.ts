import { gql } from '@apollo/client/core';

export const LOGIN_MUTATION = gql`
  mutation Login($username: String!, $password: String!) {
    login(username: $username, password: $password) {
      ID
      Username
      Token
    }
  }
`;

export const SIGNUP_MUTATION = gql`
  mutation Signup($username: String!, $password: String!) {
    signup(username: $username, password: $password) {
      ID
      Username
      Token
    }
  }
`;

export const CREATE_GAME_MUTATION = gql`
  mutation CreateGame($input: InputCreateGame!) {
    createGame(input: $input) {
      ID
      CreatedAt
      Players {
        ID
        Username
      }
      Turn {
        Player
        Phase
        Number
        Priority
      }
    }
  }
`;

export const JOIN_GAME_MUTATION = gql`
  mutation JoinGame($input: InputJoinGame) {
    joinGame(input: $input) {
      ID
      Players {
        ID
        Username
      }
    }
  }
`;

export const UPDATE_GAME_MUTATION = gql`
  mutation UpdateGame($input: InputGame!) {
    updateGame(input: $input) {
      ID
      Turn {
        Player
        Phase
        Number
        Priority
      }
      Stack {
        ID
        Name
        CurrentZone
      }
    }
  }
`;

export const PASS_PRIORITY_MUTATION = gql`
  mutation PassPriority($gameID: String!, $toPlayer: String!) {
    passPriority(gameID: $gameID, toPlayer: $toPlayer) {
      ID
      Turn {
        Player
        Phase
        Number
        Priority
      }
    }
  }
`;

export const ADVANCE_PHASE_MUTATION = gql`
  mutation AdvancePhase($gameID: String!, $phase: String!, $number: Int) {
    advancePhase(gameID: $gameID, phase: $phase, number: $number) {
      ID
      Turn {
        Player
        Phase
        Number
        Priority
      }
    }
  }
`;

export const UPDATE_BOARDSTATE_MUTATION = gql`
  mutation UpdateBoardState($input: InputBoardState!) {
    updateBoardState(input: $input) {
      UserID
      User
      GameID
      Life
      Commander { ID Name }
      Battlefield { ID Name Tapped }
      Hand { ID Name Tapped }
      Graveyard { ID Name Tapped }
      Exiled { ID Name Tapped }
      Revealed { ID Name Tapped }
    }
  }
`;
