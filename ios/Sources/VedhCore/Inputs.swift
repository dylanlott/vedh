import Foundation

public struct BoardStateInput: Encodable, Equatable, Sendable {
    public let userID: String
    public let user: String
    public let gameID: String
    public let life: Int
    public let decklist: String?
    public let commander: [GameCard]
    public let library: [GameCard]
    public let graveyard: [GameCard]
    public let exiled: [GameCard]
    public let battlefield: [GameCard]
    public let hand: [GameCard]
    public let revealed: [GameCard]
    public let controlled: [GameCard]
    public let counters: [CardCounter]

    public init(
        userID: String,
        user: String,
        gameID: String,
        life: Int,
        decklist: String? = nil,
        commander: [GameCard] = [],
        library: [GameCard] = [],
        graveyard: [GameCard] = [],
        exiled: [GameCard] = [],
        battlefield: [GameCard] = [],
        hand: [GameCard] = [],
        revealed: [GameCard] = [],
        controlled: [GameCard] = [],
        counters: [CardCounter] = []
    ) {
        self.userID = userID
        self.user = user
        self.gameID = gameID
        self.life = life
        self.decklist = decklist
        self.commander = commander
        self.library = library
        self.graveyard = graveyard
        self.exiled = exiled
        self.battlefield = battlefield
        self.hand = hand
        self.revealed = revealed
        self.controlled = controlled
        self.counters = counters
    }

    public init(boardState: BoardState, life: Int? = nil) {
        self.init(
            userID: boardState.userID,
            user: boardState.user,
            gameID: boardState.gameID,
            life: life ?? boardState.life,
            commander: boardState.commander,
            library: boardState.library,
            graveyard: boardState.graveyard,
            exiled: boardState.exiled,
            battlefield: boardState.battlefield,
            hand: boardState.hand,
            revealed: boardState.revealed,
            controlled: boardState.controlled,
            counters: boardState.counters ?? []
        )
    }

    enum CodingKeys: String, CodingKey {
        case userID = "UserID"
        case user = "User"
        case gameID = "GameID"
        case life = "Life"
        case decklist = "Decklist"
        case commander = "Commander"
        case library = "Library"
        case graveyard = "Graveyard"
        case exiled = "Exiled"
        case battlefield = "Battlefield"
        case hand = "Hand"
        case revealed = "Revealed"
        case controlled = "Controlled"
        case counters = "Counters"
    }
}

public struct CreateGameInput: Encodable, Equatable, Sendable {
    public let id: String
    public let turn: GameTurn
    public let handle: String?
    public let formatID: String?
    public let players: [BoardStateInput]

    public init(
        id: String,
        turn: GameTurn,
        handle: String? = nil,
        formatID: String? = nil,
        players: [BoardStateInput]
    ) {
        self.id = id
        self.turn = turn
        self.handle = handle
        self.formatID = formatID
        self.players = players
    }

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case turn = "Turn"
        case handle = "Handle"
        case formatID = "FormatID"
        case players = "Players"
    }
}

public struct JoinGameInput: Encodable, Equatable, Sendable {
    public let id: String
    public let decklist: String?
    public let boardState: BoardStateInput

    public init(id: String, decklist: String? = nil, boardState: BoardStateInput) {
        self.id = id
        self.decklist = decklist
        self.boardState = boardState
    }

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case decklist = "Decklist"
        case boardState = "BoardState"
    }
}
