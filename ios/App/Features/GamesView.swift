import SwiftUI
import VedhCore

struct GamesView: View {
    @EnvironmentObject private var model: AppModel
    @State private var showingCreate = false

    var body: some View {
        ZStack {
            VedhBackground()
            content
        }
        .navigationTitle("Games")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingCreate = true
                } label: {
                    Label("New game", systemImage: "plus")
                }
                .accessibilityIdentifier("games.create")
            }
        }
        .sheet(isPresented: $showingCreate) {
            NavigationStack {
                CreateGameView()
            }
        }
        .task {
            if model.games.isEmpty {
                await model.loadGames()
            }
        }
        .refreshable { await model.loadGames() }
    }

    @ViewBuilder
    private var content: some View {
        if model.isWorking && model.games.isEmpty {
            ProgressView("Loading games…")
        } else if model.games.isEmpty {
            ContentUnavailableView {
                Label("No games", systemImage: "rectangle.stack.badge.plus")
            } description: {
                Text("Create a table or join with an invite.")
            } actions: {
                Button("Create game") { showingCreate = true }
                    .buttonStyle(.borderedProminent)
            }
        } else {
            List(model.games) { game in
                NavigationLink {
                    GameView(gameID: game.id)
                } label: {
                    GameRow(game: game)
                }
                .listRowBackground(Color.vedhPanel)
                .accessibilityIdentifier("games.row.\(game.id)")
            }
            .scrollContentBackground(.hidden)
        }
    }
}

private struct GameRow: View {
    let game: GameSummary

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(alignment: .firstTextBaseline) {
                Text(game.displayName)
                    .font(.headline)
                    .foregroundStyle(Color.vedhParchment)
                Spacer()
                Text(game.status == "FINISHED" ? "Finished" : "In progress")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(game.status == "FINISHED" ? Color.vedhMuted : Color.vedhJewel)
            }

            if let turn = game.turn {
                Text("Turn \(turn.number) · \(turn.phase) · priority \(turn.priority)")
                    .font(.subheadline)
                    .foregroundStyle(Color.vedhMuted)
                    .lineLimit(1)
            }

            Text(game.id)
                .font(.caption2.monospaced())
                .foregroundStyle(Color.vedhMuted.opacity(0.8))
                .lineLimit(1)
        }
        .padding(.vertical, 6)
    }
}
