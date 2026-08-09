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

export const TRACK_PRODUCT_EVENT_MUTATION = gql`
  mutation TrackProductEvent($input: InputProductEvent!) {
    trackProductEvent(input: $input)
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

export const CLAIM_WIN_MUTATION = gql`
  mutation ClaimWin($gameID: String!, $condition: String) {
    claimWin(gameID: $gameID, condition: $condition) {
      ID
      Status
      Result
      WinnerIDs
      WinCondition
      PendingWinClaim { ClaimedBy Condition Remaining }
      Turn { Player Phase Number Priority }
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

// --- Guest activation (Phase 2, plan 02-01) ---
// guestSession and previewDeck are Mutation fields per the locked
// api-contract in .planning/intel/constraints.md.

export const GUEST_SESSION_MUTATION = gql`
  mutation GuestSession($displayName: String, $sessionID: String!) {
    guestSession(displayName: $displayName, sessionID: $sessionID) {
      ID
      Username
      DisplayName
      IsGuest
      Token
      GuestCredential
    }
  }
`;

export const PREVIEW_DECK_MUTATION = gql`
  mutation PreviewDeck($input: InputDeckImport!) {
    previewDeck(input: $input) {
      SourceType
      CardCount
      CanContinue
      Warnings
      BlockingErrors
      Entries {
        Quantity
        Name
        Section
        SourceLine
        Resolved
      }
      CommanderCandidates {
        ID
        Name
      }
      Unresolved {
        SourceLine
        RawLine
        Name
        Reason
        Candidates {
          Name
          Score
          LowConfidence
        }
      }
    }
  }
`;
