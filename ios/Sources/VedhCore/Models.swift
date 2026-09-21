import Foundation

public struct AuthSession: Codable, Equatable, Sendable {
    public let id: String
    public let username: String
    public let token: String

    public init(id: String, username: String, token: String) {
        self.id = id
        self.username = username
        self.token = token
    }

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case username = "Username"
        case token = "Token"
    }
}

public struct GameRule: Codable, Equatable, Sendable {
    public let name: String
    public let value: String

    public init(name: String, value: String) {
        self.name = name
        self.value = value
    }

    enum CodingKeys: String, CodingKey {
        case name = "Name"
        case value = "Value"
    }
}

public struct GameTurn: Codable, Equatable, Sendable {
    public let player: String
    public let phase: String
    public let number: Int
    public let priority: String

    public init(player: String, phase: String, number: Int, priority: String) {
        self.player = player
        self.phase = phase
        self.number = number
        self.priority = priority
    }

    enum CodingKeys: String, CodingKey {
        case player = "Player"
        case phase = "Phase"
        case number = "Number"
        case priority = "Priority"
    }
}

public struct CardCounter: Codable, Equatable, Sendable {
    public let name: String
    public let value: String

    public init(name: String, value: String) {
        self.name = name
        self.value = value
    }

    enum CodingKeys: String, CodingKey {
        case name = "Name"
        case value = "Value"
    }
}

public struct GameCard: Codable, Identifiable, Equatable, Sendable {
    public let id: String
    public let name: String
    public let faceName: String?
    public let quantity: Int?
    public let tapped: Bool?
    public let flipped: Bool?
    public let counters: [CardCounter]?
    public let colors: String?
    public let colorIdentity: String?
    public let faceManaValue: String?
    public let faceConvertedManaCost: String?
    public let cmc: String?
    public let manaCost: String?
    public let uuid: String?
    public let power: String?
    public let toughness: String?
    public let types: String?
    public let subtypes: String?
    public let supertypes: String?
    public let text: String?
    public let tcgID: String?
    public let scryfallID: String?
    public let screenX: Double?
    public let screenY: Double?
    public let currentZone: String?

    public init(
        id: String,
        name: String,
        faceName: String? = nil,
        quantity: Int? = nil,
        tapped: Bool? = nil,
        flipped: Bool? = nil,
        counters: [CardCounter]? = nil,
        colors: String? = nil,
        colorIdentity: String? = nil,
        faceManaValue: String? = nil,
        faceConvertedManaCost: String? = nil,
        cmc: String? = nil,
        manaCost: String? = nil,
        uuid: String? = nil,
        power: String? = nil,
        toughness: String? = nil,
        types: String? = nil,
        subtypes: String? = nil,
        supertypes: String? = nil,
        text: String? = nil,
        tcgID: String? = nil,
        scryfallID: String? = nil,
        screenX: Double? = nil,
        screenY: Double? = nil,
        currentZone: String? = nil
    ) {
        self.id = id
        self.name = name
        self.faceName = faceName
        self.quantity = quantity
        self.tapped = tapped
        self.flipped = flipped
        self.counters = counters
        self.colors = colors
        self.colorIdentity = colorIdentity
        self.faceManaValue = faceManaValue
        self.faceConvertedManaCost = faceConvertedManaCost
        self.cmc = cmc
        self.manaCost = manaCost
        self.uuid = uuid
        self.power = power
        self.toughness = toughness
        self.types = types
        self.subtypes = subtypes
        self.supertypes = supertypes
        self.text = text
        self.tcgID = tcgID
        self.scryfallID = scryfallID
        self.screenX = screenX
        self.screenY = screenY
        self.currentZone = currentZone
    }

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case name = "Name"
        case faceName = "FaceName"
        case quantity = "Quantity"
        case tapped = "Tapped"
        case flipped = "Flipped"
        case counters = "Counters"
        case colors = "Colors"
        case colorIdentity = "ColorIdentity"
        case faceManaValue = "FaceManaValue"
        case faceConvertedManaCost = "FaceConvertedManaCost"
        case cmc = "CMC"
        case manaCost = "ManaCost"
        case uuid = "UUID"
        case power = "Power"
        case toughness = "Toughness"
        case types = "Types"
        case subtypes = "Subtypes"
        case supertypes = "Supertypes"
        case text = "Text"
        case tcgID = "TCGID"
        case scryfallID = "ScryfallID"
        case screenX = "ScreenX"
        case screenY = "ScreenY"
        case currentZone = "CurrentZone"
    }
}

public struct BoardState: Codable, Equatable, Sendable {
    public let userID: String
    public let user: String
    public let gameID: String
    public let life: Int
    public let commander: [GameCard]
    public let library: [GameCard]
    public let graveyard: [GameCard]
    public let exiled: [GameCard]
    public let battlefield: [GameCard]
    public let hand: [GameCard]
    public let revealed: [GameCard]
    public let controlled: [GameCard]
    public let counters: [CardCounter]?

    public init(
        userID: String,
        user: String,
        gameID: String,
        life: Int,
        commander: [GameCard] = [],
        library: [GameCard] = [],
        graveyard: [GameCard] = [],
        exiled: [GameCard] = [],
        battlefield: [GameCard] = [],
        hand: [GameCard] = [],
        revealed: [GameCard] = [],
        controlled: [GameCard] = [],
        counters: [CardCounter]? = nil
    ) {
        self.userID = userID
        self.user = user
        self.gameID = gameID
        self.life = life
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

    public func withLife(_ life: Int) -> BoardState {
        BoardState(
            userID: userID,
            user: user,
            gameID: gameID,
            life: life,
            commander: commander,
            library: library,
            graveyard: graveyard,
            exiled: exiled,
            battlefield: battlefield,
            hand: hand,
            revealed: revealed,
            controlled: controlled,
            counters: counters
        )
    }

    enum CodingKeys: String, CodingKey {
        case userID = "UserID"
        case user = "User"
        case gameID = "GameID"
        case life = "Life"
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

public struct GamePlayer: Codable, Identifiable, Equatable, Sendable {
    public let id: String
    public let username: String
    public let boardState: BoardState?

    public init(id: String, username: String, boardState: BoardState? = nil) {
        self.id = id
        self.username = username
        self.boardState = boardState
    }

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case username = "Username"
        case boardState = "Boardstate"
    }
}

public struct PendingWinClaim: Codable, Equatable, Sendable {
    public let claimedBy: String
    public let condition: String?
    public let remaining: [String]

    enum CodingKeys: String, CodingKey {
        case claimedBy = "ClaimedBy"
        case condition = "Condition"
        case remaining = "Remaining"
    }
}

public struct GameSummary: Codable, Identifiable, Equatable, Sendable {
    public let id: String
    public let createdAt: String?
    public let rules: [GameRule]?
    public let turn: GameTurn?
    public let players: [GamePlayer]
    public let status: String?

    public init(
        id: String,
        createdAt: String? = nil,
        rules: [GameRule]? = nil,
        turn: GameTurn? = nil,
        players: [GamePlayer] = [],
        status: String? = nil
    ) {
        self.id = id
        self.createdAt = createdAt
        self.rules = rules
        self.turn = turn
        self.players = players
        self.status = status
    }

    public var displayName: String {
        let names = players.map(\.username).filter { !$0.isEmpty }
        return names.isEmpty ? "Game \(id.prefix(8))" : names.joined(separator: ", ")
    }

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case createdAt = "CreatedAt"
        case rules = "Rules"
        case turn = "Turn"
        case players = "Players"
        case status = "Status"
    }
}

public struct GameDetail: Codable, Identifiable, Equatable, Sendable {
    public let id: String
    public let createdAt: String?
    public let rules: [GameRule]?
    public let turn: GameTurn?
    public let players: [GamePlayer]
    public let stack: [GameCard]
    public let status: String
    public let result: String?
    public let winnerIDs: [String]?
    public let winCondition: String?
    public let pendingWinClaim: PendingWinClaim?

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case createdAt = "CreatedAt"
        case rules = "Rules"
        case turn = "Turn"
        case players = "Players"
        case stack = "Stack"
        case status = "Status"
        case result = "Result"
        case winnerIDs = "WinnerIDs"
        case winCondition = "WinCondition"
        case pendingWinClaim = "PendingWinClaim"
    }
}

public struct GameFormatZone: Codable, Identifiable, Equatable, Sendable {
    public let id: String
    public let label: String
    public let visibility: String
    public let kind: String
    public let supportsCards: Bool

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case label = "Label"
        case visibility = "Visibility"
        case kind = "Kind"
        case supportsCards = "SupportsCards"
    }
}

public struct GameFormat: Codable, Identifiable, Equatable, Sendable {
    public let id: String
    public let name: String
    public let startingLife: Int
    public let defaultDeckSize: Int
    public let commanderEnabled: Bool
    public let phaseSequence: [String]
    public let zones: [GameFormatZone]

    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case name = "Name"
        case startingLife = "StartingLife"
        case defaultDeckSize = "DefaultDeckSize"
        case commanderEnabled = "CommanderEnabled"
        case phaseSequence = "PhaseSequence"
        case zones = "Zones"
    }
}
