import SwiftUI
import VedhCore

struct GameView: View {
    @EnvironmentObject private var model: AppModel
    let gameID: String
    @State private var showingClaim = false
    @State private var winCondition = ""

    private var game: GameDetail? {
        guard model.activeGame?.id == gameID else { return nil }
        return model.activeGame
    }

    var body: some View {
        ZStack {
            VedhBackground()
            if let game {
                gameContent(game)
            } else {
                ProgressView("Loading game…")
            }
        }
        .navigationTitle("Game \(gameID.prefix(8))")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                ShareLink(item: URL(string: "https://vedh.xyz/join/\(gameID)")!) {
                    Image(systemName: "square.and.arrow.up")
                }
                Button {
                    Task { await model.loadGame(id: gameID) }
                } label: {
                    Image(systemName: "arrow.clockwise")
                }
                .disabled(model.isWorking)
            }
        }
        .task(id: gameID) {
            await model.loadGame(id: gameID)
            while !Task.isCancelled {
                do {
                    try await Task.sleep(nanoseconds: 3_000_000_000)
                } catch {
                    return
                }
                await model.loadGame(id: gameID, silent: true)
            }
        }
        .onDisappear { model.clearActiveGame() }
        .alert("Claim win", isPresented: $showingClaim) {
            TextField("Win condition (optional)", text: $winCondition)
            Button("Cancel", role: .cancel) { winCondition = "" }
            Button("Claim") {
                Task {
                    await model.claimWin(condition: winCondition)
                    winCondition = ""
                }
            }
        } message: {
            Text("The table will receive the claim through normal priority order.")
        }
    }

    private func gameContent(_ game: GameDetail) -> some View {
        ScrollView {
            LazyVStack(spacing: 14) {
                syncBanner
                TurnPanel(
                    game: game,
                    username: model.session?.username,
                    isWorking: model.isWorking,
                    onPassPriority: { Task { await model.passPriority() } },
                    onAdvancePhase: { Task { await model.advancePhase() } },
                    onClaimWin: { showingClaim = true }
                )

                if let claim = game.pendingWinClaim {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Win claim: \(claim.claimedBy)")
                            .font(.headline)
                        if let condition = claim.condition, !condition.isEmpty {
                            Text(condition)
                        }
                        Text(claim.remaining.isEmpty ? "Awaiting priority" : "Awaiting \(claim.remaining.joined(separator: ", "))")
                            .font(.caption)
                            .foregroundStyle(Color.vedhMuted)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .vedhPanel()
                }

                if !game.stack.isEmpty {
                    ZonePanel(title: "Stack", cards: game.stack, hidden: false)
                }

                ForEach(game.players) { player in
                    PlayerPanel(
                        player: player,
                        isCurrentUser: player.id == model.session?.id || player.username == model.session?.username,
                        isActivePlayer: player.username == game.turn?.player,
                        hasPriority: player.username == game.turn?.priority,
                        isWorking: model.isWorking,
                        onLifeChange: { delta in
                            Task { await model.changeOwnLife(by: delta) }
                        }
                    )
                }
            }
            .padding(16)
        }
        .refreshable { await model.loadGame(id: gameID) }
    }

    @ViewBuilder
    private var syncBanner: some View {
        if let issue = model.syncIssue {
            HStack(alignment: .top) {
                Image(systemName: "wifi.exclamationmark")
                VStack(alignment: .leading, spacing: 2) {
                    Text("Live refresh paused").fontWeight(.semibold)
                    Text(issue).font(.caption).lineLimit(2)
                }
                Spacer()
            }
            .foregroundStyle(Color.vedhParchment)
            .padding(12)
            .background(Color.vedhDanger.opacity(0.35))
            .clipShape(RoundedRectangle(cornerRadius: 12))
        } else if let lastSync = model.lastSync {
            HStack {
                Circle().fill(Color.vedhJewel).frame(width: 8, height: 8)
                Text("Updated \(lastSync, style: .relative)")
                    .font(.caption)
                    .foregroundStyle(Color.vedhMuted)
                Spacer()
            }
            .accessibilityLabel("Game state connected")
        }
    }
}

private struct TurnPanel: View {
    let game: GameDetail
    let username: String?
    let isWorking: Bool
    let onPassPriority: () -> Void
    let onAdvancePhase: () -> Void
    let onClaimWin: () -> Void

    private var ownsPriority: Bool {
        guard let username else { return false }
        return username == (game.turn?.priority ?? game.turn?.player)
    }

    private var ownsTurn: Bool {
        username == game.turn?.player
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(alignment: .top, spacing: 18) {
                metric("TURN", "\(game.turn?.number ?? 0)")
                metric("PHASE", game.turn?.phase ?? "—")
                metric("PRIORITY", game.turn?.priority ?? game.turn?.player ?? "—")
            }

            if game.status == "FINISHED" {
                Label(gameResult, systemImage: "checkmark.seal.fill")
                    .foregroundStyle(Color.vedhJewel)
            } else {
                HStack {
                    Button("Pass priority", action: onPassPriority)
                        .disabled(!ownsPriority || isWorking || game.players.count < 2)
                    Button("Next phase", action: onAdvancePhase)
                        .disabled(!ownsTurn || isWorking)
                    Button("Claim win", action: onClaimWin)
                        .disabled(!ownsPriority || isWorking)
                }
                .buttonStyle(.bordered)
                .font(.caption.weight(.semibold))
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .vedhPanel()
    }

    private var gameResult: String {
        if let condition = game.winCondition, !condition.isEmpty {
            return "Finished · \(condition)"
        }
        return "Finished · \(game.result ?? "result recorded")"
    }

    private func metric(_ label: String, _ value: String) -> some View {
        VStack(alignment: .leading, spacing: 3) {
            Text(label)
                .font(.caption2.weight(.bold))
                .foregroundStyle(Color.vedhBrass)
            Text(value)
                .font(label == "TURN" ? .title.bold() : .subheadline.weight(.semibold))
                .foregroundStyle(Color.vedhParchment)
                .lineLimit(2)
                .minimumScaleFactor(0.72)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

private struct PlayerPanel: View {
    let player: GamePlayer
    let isCurrentUser: Bool
    let isActivePlayer: Bool
    let hasPriority: Bool
    let isWorking: Bool
    let onLifeChange: (Int) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(alignment: .center) {
                VStack(alignment: .leading, spacing: 4) {
                    Text(player.username)
                        .font(.title3.bold())
                        .foregroundStyle(Color.vedhParchment)
                    HStack(spacing: 8) {
                        if isCurrentUser { stateBadge("You", icon: "person.fill") }
                        if isActivePlayer { stateBadge("Active", icon: "play.fill") }
                        if hasPriority { stateBadge("Priority", icon: "hand.raised.fill") }
                    }
                }
                Spacer()
                VStack(spacing: 5) {
                    Text("\(player.boardState?.life ?? 0)")
                        .font(.system(size: 42, weight: .black, design: .rounded))
                        .foregroundStyle(Color.vedhParchment)
                        .contentTransition(.numericText())
                    Text("LIFE")
                        .font(.caption2.weight(.bold))
                        .foregroundStyle(Color.vedhBrass)
                }
            }

            if isCurrentUser {
                HStack {
                    lifeButton("−5", delta: -5)
                    lifeButton("−1", delta: -1)
                    lifeButton("+1", delta: 1)
                    lifeButton("+5", delta: 5)
                }
            }

            if let board = player.boardState {
                LazyVGrid(columns: [GridItem(.adaptive(minimum: 92))], spacing: 8) {
                    zoneCount("Command", board.commander.count)
                    zoneCount("Library", board.library.count)
                    zoneCount("Hand", board.hand.count)
                    zoneCount("Field", board.battlefield.count)
                    zoneCount("Graveyard", board.graveyard.count)
                    zoneCount("Exile", board.exiled.count)
                }

                ZonePanel(title: "Commander", cards: board.commander, hidden: false)
                ZonePanel(title: "Battlefield", cards: board.battlefield, hidden: false)
                ZonePanel(title: "Graveyard", cards: board.graveyard, hidden: false)
                ZonePanel(title: "Exile", cards: board.exiled, hidden: false)
                ZonePanel(title: "Revealed", cards: board.revealed, hidden: false)
                ZonePanel(title: "Controlled", cards: board.controlled, hidden: false)
                ZonePanel(title: "Hand", cards: board.hand, hidden: !isCurrentUser)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .vedhPanel()
        .overlay(alignment: .leading) {
            if isActivePlayer {
                Rectangle()
                    .fill(Color.vedhJewel)
                    .frame(width: 4)
            }
        }
    }

    private func lifeButton(_ label: String, delta: Int) -> some View {
        Button(label) { onLifeChange(delta) }
            .buttonStyle(.bordered)
            .frame(maxWidth: .infinity)
            .disabled(isWorking)
            .accessibilityLabel(delta > 0 ? "Gain \(delta) life" : "Lose \(-delta) life")
    }

    private func stateBadge(_ label: String, icon: String) -> some View {
        Label(label, systemImage: icon)
            .font(.caption2.weight(.bold))
            .foregroundStyle(Color.vedhJewel)
    }

    private func zoneCount(_ label: String, _ count: Int) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text("\(count)").font(.headline.monospacedDigit())
            Text(label.uppercased())
                .font(.caption2)
                .foregroundStyle(Color.vedhMuted)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(10)
        .background(Color.vedhRaised)
        .clipShape(RoundedRectangle(cornerRadius: 9))
    }
}

private struct ZonePanel: View {
    let title: String
    let cards: [GameCard]
    let hidden: Bool

    var body: some View {
        if !cards.isEmpty {
            DisclosureGroup {
                if hidden {
                    Text("Hidden information · \(cards.count) cards")
                        .foregroundStyle(Color.vedhMuted)
                        .frame(maxWidth: .infinity, alignment: .leading)
                } else {
                    ForEach(Array(cards.enumerated()), id: \.offset) { _, card in
                        HStack {
                            Image(systemName: card.tapped == true ? "rectangle.landscape.rotate" : "rectangle.portrait")
                                .foregroundStyle(Color.vedhBrass)
                            Text(card.name)
                            Spacer()
                            if card.tapped == true {
                                Text("Tapped")
                                    .font(.caption)
                                    .foregroundStyle(Color.vedhMuted)
                            }
                        }
                        .font(.subheadline)
                        .padding(.vertical, 3)
                    }
                }
            } label: {
                Text("\(title) · \(cards.count)")
                    .font(.subheadline.weight(.semibold))
                    .foregroundStyle(Color.vedhParchment)
            }
        }
    }
}
