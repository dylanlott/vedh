import Foundation
import VedhCore

@MainActor
final class AppModel: ObservableObject {
    @Published private(set) var session: AuthSession?
    @Published private(set) var games: [GameSummary] = []
    @Published private(set) var activeGame: GameDetail?
    @Published private(set) var formats: [GameFormat] = []
    @Published private(set) var isWorking = false
    @Published var errorMessage: String?
    @Published private(set) var syncIssue: String?
    @Published private(set) var lastSync: Date?

    private let vault: KeychainSessionStore
    private let api: VedhAPI

    init(bundle: Bundle = .main, vault: KeychainSessionStore = KeychainSessionStore()) {
        self.vault = vault

        let configuredEndpoint = bundle.object(forInfoDictionaryKey: "VEDH_GRAPHQL_ENDPOINT") as? String
        let endpoint = URL(string: configuredEndpoint ?? "https://api.vedh.xyz/graphql")!
        let client = GraphQLClient(endpoint: endpoint) { [vault] in
            try? vault.load()?.token
        }
        self.api = VedhAPI(client: client)
        self.session = try? vault.load()
    }

    func authenticate(username: String, password: String, creatingAccount: Bool) async {
        isWorking = true
        errorMessage = nil
        defer { isWorking = false }

        do {
            let session = if creatingAccount {
                try await api.signup(username: username, password: password)
            } else {
                try await api.login(username: username, password: password)
            }
            try vault.save(session)
            self.session = session
            await loadGames(silent: true)
            await loadFormats(silent: true)
        } catch {
            present(error)
        }
    }

    func logout() {
        do {
            try vault.clear()
        } catch {
            present(error)
        }
        session = nil
        games = []
        formats = []
        activeGame = nil
        syncIssue = nil
        lastSync = nil
    }

    func loadGames(silent: Bool = false) async {
        guard session != nil else { return }
        if !silent { isWorking = true }
        defer { if !silent { isWorking = false } }

        do {
            games = try await api.games()
            syncIssue = nil
            lastSync = Date()
        } catch {
            if silent {
                syncIssue = error.localizedDescription
            } else {
                present(error)
            }
        }
    }

    func loadFormats(silent: Bool = false) async {
        if !silent { isWorking = true }
        defer { if !silent { isWorking = false } }

        do {
            formats = try await api.formats()
        } catch {
            if !silent { present(error) }
        }
    }

    func loadGame(id: String, silent: Bool = false) async {
        if !silent { isWorking = true }
        defer { if !silent { isWorking = false } }

        do {
            activeGame = try await api.game(id: id)
            syncIssue = nil
            lastSync = Date()
        } catch {
            if silent {
                syncIssue = error.localizedDescription
            } else {
                present(error)
            }
        }
    }

    func clearActiveGame() {
        activeGame = nil
        syncIssue = nil
        lastSync = nil
    }

    func createGame(decklist: String, commander: GameCard?) async -> String? {
        guard let session else { return nil }
        isWorking = true
        errorMessage = nil
        defer { isWorking = false }

        do {
            let game = try await api.createCommanderGame(
                session: session,
                decklist: decklist,
                commander: commander
            )
            games.removeAll { $0.id == game.id }
            games.insert(game, at: 0)
            return game.id
        } catch {
            present(error)
            return nil
        }
    }

    func joinGame(input: String, decklist: String, commander: GameCard?) async -> String? {
        guard let session else { return nil }
        guard let gameID = InviteParser.gameID(from: input) else {
            errorMessage = "Enter a valid game ID or vedh invite link."
            return nil
        }

        isWorking = true
        errorMessage = nil
        defer { isWorking = false }

        do {
            let game = try await api.joinCommanderGame(
                id: gameID,
                session: session,
                decklist: decklist,
                commander: commander
            )
            games.removeAll { $0.id == game.id }
            games.insert(game, at: 0)
            return game.id
        } catch {
            present(error)
            return nil
        }
    }

    func changeOwnLife(by delta: Int) async {
        guard
            let session,
            let game = activeGame,
            let player = game.players.first(where: { $0.id == session.id || $0.username == session.username }),
            let boardState = player.boardState
        else { return }

        isWorking = true
        defer { isWorking = false }
        do {
            _ = try await api.updateLife(boardState: boardState, life: boardState.life + delta)
            await loadGame(id: game.id, silent: true)
        } catch {
            present(error)
        }
    }

    func passPriority() async {
        guard
            let game = activeGame,
            let next = GameRules.nextPriorityPlayer(
                current: game.turn?.priority ?? game.turn?.player,
                players: game.players
            )
        else { return }

        await performGameAction {
            try await self.api.passPriority(gameID: game.id, to: next)
        }
    }

    func advancePhase() async {
        guard let game = activeGame else { return }
        let next = GameRules.nextCommanderPhase(after: game.turn?.phase)
        await performGameAction {
            try await self.api.advancePhase(gameID: game.id, phase: next)
        }
    }

    func claimWin(condition: String) async {
        guard let game = activeGame else { return }
        await performGameAction {
            try await self.api.claimWin(gameID: game.id, condition: condition)
        }
    }

    func searchCards(name: String) async throws -> [GameCard] {
        try await api.searchCards(name: name)
    }

    private func performGameAction(_ action: () async throws -> GameDetail) async {
        isWorking = true
        errorMessage = nil
        defer { isWorking = false }
        do {
            activeGame = try await action()
            syncIssue = nil
            lastSync = Date()
        } catch {
            present(error)
        }
    }

    private func present(_ error: Error) {
        errorMessage = error.localizedDescription
    }
}
