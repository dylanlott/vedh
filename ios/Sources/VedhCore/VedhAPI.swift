import Foundation

public final class VedhAPI: @unchecked Sendable {
    private let client: GraphQLClient

    public init(client: GraphQLClient) {
        self.client = client
    }

    public func login(username: String, password: String) async throws -> AuthSession {
        struct Variables: Encodable { let username: String; let password: String }
        struct Payload: Decodable { let login: AuthSession }

        let payload: Payload = try await client.execute(
            query: Self.loginMutation,
            variables: Variables(username: username, password: password)
        )
        return payload.login
    }

    public func signup(username: String, password: String) async throws -> AuthSession {
        struct Variables: Encodable { let username: String; let password: String }
        struct Payload: Decodable { let signup: AuthSession }

        let payload: Payload = try await client.execute(
            query: Self.signupMutation,
            variables: Variables(username: username, password: password)
        )
        return payload.signup
    }

    public func games(offset: Int = 0, limit: Int = 50) async throws -> [GameSummary] {
        struct Variables: Encodable { let offset: Int; let limit: Int }
        struct Payload: Decodable { let games: [GameSummary] }

        let payload: Payload = try await client.execute(
            query: Self.gamesQuery,
            variables: Variables(offset: offset, limit: limit)
        )
        return payload.games
    }

    public func formats() async throws -> [GameFormat] {
        struct Variables: Encodable {}
        struct Payload: Decodable { let formats: [GameFormat] }

        let payload: Payload = try await client.execute(
            query: Self.formatsQuery,
            variables: Variables()
        )
        return payload.formats
    }

    public func game(id: String) async throws -> GameDetail {
        struct Variables: Encodable { let gameID: String }
        struct Payload: Decodable { let getGame: GameDetail }

        let payload: Payload = try await client.execute(
            query: Self.gameQuery,
            variables: Variables(gameID: id)
        )
        return payload.getGame
    }

    public func createGame(input: CreateGameInput) async throws -> GameSummary {
        struct Variables: Encodable { let input: CreateGameInput }
        struct Payload: Decodable { let createGame: GameSummary }

        let payload: Payload = try await client.execute(
            query: Self.createGameMutation,
            variables: Variables(input: input)
        )
        return payload.createGame
    }

    public func createCommanderGame(
        session: AuthSession,
        decklist: String,
        commander: GameCard?,
        formatID: String = "EDH"
    ) async throws -> GameSummary {
        let gameID = UUID().uuidString
        let boardState = BoardStateInput(
            userID: session.id,
            user: session.username,
            gameID: gameID,
            life: 40,
            decklist: decklist.nilIfEmpty,
            commander: commander.map { [$0] } ?? []
        )
        let input = CreateGameInput(
            id: gameID,
            turn: GameTurn(
                player: session.username,
                phase: "MAIN PHASE 1",
                number: 1,
                priority: session.username
            ),
            formatID: formatID,
            players: [boardState]
        )
        return try await createGame(input: input)
    }

    public func joinGame(input: JoinGameInput) async throws -> GameSummary {
        struct Variables: Encodable { let input: JoinGameInput }
        struct Payload: Decodable { let joinGame: GameSummary }

        let payload: Payload = try await client.execute(
            query: Self.joinGameMutation,
            variables: Variables(input: input)
        )
        return payload.joinGame
    }

    public func joinCommanderGame(
        id: String,
        session: AuthSession,
        decklist: String,
        commander: GameCard?
    ) async throws -> GameSummary {
        let boardState = BoardStateInput(
            userID: session.id,
            user: session.username,
            gameID: id,
            life: 40,
            decklist: decklist.nilIfEmpty,
            commander: commander.map { [$0] } ?? []
        )
        return try await joinGame(
            input: JoinGameInput(id: id, decklist: decklist.nilIfEmpty, boardState: boardState)
        )
    }

    public func updateBoardState(_ input: BoardStateInput) async throws -> BoardState {
        struct Variables: Encodable { let input: BoardStateInput }
        struct Payload: Decodable { let updateBoardState: BoardState }

        let payload: Payload = try await client.execute(
            query: Self.updateBoardStateMutation,
            variables: Variables(input: input)
        )
        return payload.updateBoardState
    }

    public func updateLife(boardState: BoardState, life: Int) async throws -> BoardState {
        try await updateBoardState(BoardStateInput(boardState: boardState, life: life))
    }

    public func passPriority(gameID: String, to player: String) async throws -> GameDetail {
        struct Variables: Encodable { let gameID: String; let toPlayer: String }
        struct Payload: Decodable { let passPriority: GameDetail }

        let payload: Payload = try await client.execute(
            query: Self.passPriorityMutation,
            variables: Variables(gameID: gameID, toPlayer: player)
        )
        return payload.passPriority
    }

    public func advancePhase(gameID: String, phase: String) async throws -> GameDetail {
        struct Variables: Encodable { let gameID: String; let phase: String }
        struct Payload: Decodable { let advancePhase: GameDetail }

        let payload: Payload = try await client.execute(
            query: Self.advancePhaseMutation,
            variables: Variables(gameID: gameID, phase: phase)
        )
        return payload.advancePhase
    }

    public func claimWin(gameID: String, condition: String?) async throws -> GameDetail {
        struct Variables: Encodable { let gameID: String; let condition: String? }
        struct Payload: Decodable { let claimWin: GameDetail }

        let payload: Payload = try await client.execute(
            query: Self.claimWinMutation,
            variables: Variables(gameID: gameID, condition: condition?.nilIfEmpty)
        )
        return payload.claimWin
    }

    public func searchCards(name: String) async throws -> [GameCard] {
        struct Variables: Encodable { let name: String }
        struct Payload: Decodable { let search: [GameCard]? }

        let payload: Payload = try await client.execute(
            query: Self.searchCardsQuery,
            variables: Variables(name: name)
        )
        return payload.search ?? []
    }
}

private extension String {
    var nilIfEmpty: String? {
        let value = trimmingCharacters(in: .whitespacesAndNewlines)
        return value.isEmpty ? nil : value
    }
}

private extension VedhAPI {
    static let loginMutation = """
    mutation Login($username: String!, $password: String!) {
      login(username: $username, password: $password) { ID Username Token }
    }
    """

    static let signupMutation = """
    mutation Signup($username: String!, $password: String!) {
      signup(username: $username, password: $password) { ID Username Token }
    }
    """

    static let gamesQuery = """
    query Games($offset: Int!, $limit: Int!) {
      games(offset: $offset, limit: $limit) {
        ID CreatedAt Status
        Rules { Name Value }
        Turn { Player Phase Number Priority }
        Players { ID Username }
      }
    }
    """

    static let formatsQuery = """
    query Formats {
      formats {
        ID Name StartingLife DefaultDeckSize CommanderEnabled PhaseSequence
        Zones { ID Label Visibility Kind SupportsCards }
      }
    }
    """

    static let createGameMutation = """
    mutation CreateGame($input: InputCreateGame!) {
      createGame(input: $input) {
        ID CreatedAt Status
        Rules { Name Value }
        Turn { Player Phase Number Priority }
        Players { ID Username }
      }
    }
    """

    static let joinGameMutation = """
    mutation JoinGame($input: InputJoinGame) {
      joinGame(input: $input) {
        ID CreatedAt Status
        Rules { Name Value }
        Turn { Player Phase Number Priority }
        Players { ID Username }
      }
    }
    """

    static let gameQuery = """
    query GetGame($gameID: String!) {
      getGame(gameID: $gameID) { ...GameDetailFields }
    }
    \(gameDetailFragment)
    """

    static let updateBoardStateMutation = """
    mutation UpdateBoardState($input: InputBoardState!) {
      updateBoardState(input: $input) { ...BoardStateFields }
    }
    \(cardFragment)
    \(boardStateFragment)
    """

    static let passPriorityMutation = """
    mutation PassPriority($gameID: String!, $toPlayer: String!) {
      passPriority(gameID: $gameID, toPlayer: $toPlayer) { ...GameDetailFields }
    }
    \(gameDetailFragment)
    """

    static let advancePhaseMutation = """
    mutation AdvancePhase($gameID: String!, $phase: String!) {
      advancePhase(gameID: $gameID, phase: $phase) { ...GameDetailFields }
    }
    \(gameDetailFragment)
    """

    static let claimWinMutation = """
    mutation ClaimWin($gameID: String!, $condition: String) {
      claimWin(gameID: $gameID, condition: $condition) { ...GameDetailFields }
    }
    \(gameDetailFragment)
    """

    static let searchCardsQuery = """
    query SearchCards($name: String) {
      search(name: $name) { ...CardFields }
    }
    \(cardFragment)
    """

    static let cardFragment = """
    fragment CardFields on Card {
      FaceName Name ID Quantity Tapped Flipped
      Counters { Name Value }
      Colors ColorIdentity FaceManaValue FaceConvertedManaCost CMC ManaCost UUID
      Power Toughness Types Subtypes Supertypes Text TCGID ScryfallID
      ScreenX ScreenY CurrentZone
    }
    """

    static let boardStateFragment = """
    fragment BoardStateFields on BoardState {
      UserID User GameID Life Counters { Name Value }
      Commander { ...CardFields }
      Library { ...CardFields }
      Graveyard { ...CardFields }
      Exiled { ...CardFields }
      Battlefield { ...CardFields }
      Hand { ...CardFields }
      Revealed { ...CardFields }
      Controlled { ...CardFields }
    }
    """

    static let gameDetailFragment = """
    \(cardFragment)
    \(boardStateFragment)
    fragment GameDetailFields on Game {
      ID CreatedAt Status Result WinnerIDs WinCondition
      PendingWinClaim { ClaimedBy Condition Remaining }
      Rules { Name Value }
      Turn { Player Phase Number Priority }
      Stack { ...CardFields }
      Players { ID Username Boardstate { ...BoardStateFields } }
    }
    """
}
