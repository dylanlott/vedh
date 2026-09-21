import SwiftUI
import VedhCore

struct JoinGameView: View {
    @EnvironmentObject private var model: AppModel
    @State private var invite = ""
    @State private var decklist = ""
    @State private var commander: GameCard?
    @State private var joinedGameID = ""
    @State private var showingJoinedGame = false

    private var parsedGameID: String? {
        InviteParser.gameID(from: invite)
    }

    var body: some View {
        ZStack {
            VedhBackground()
            Form {
                Section("Invite") {
                    TextField("Game ID or invite link", text: $invite)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .accessibilityIdentifier("join.invite")

                    if !invite.isEmpty {
                        if let parsedGameID {
                            LabeledContent("Game", value: parsedGameID)
                                .fontDesign(.monospaced)
                        } else {
                            Text("Could not detect a valid game ID.")
                                .foregroundStyle(Color.vedhDanger)
                        }
                    }
                }

                Section {
                    CommanderPicker(selection: $commander)
                }

                Section("Decklist (optional)") {
                    TextEditor(text: $decklist)
                        .font(.body.monospaced())
                        .frame(minHeight: 160)
                        .accessibilityIdentifier("join.decklist")
                }

                Section {
                    Button {
                        Task {
                            if let id = await model.joinGame(
                                input: invite,
                                decklist: decklist,
                                commander: commander
                            ) {
                                joinedGameID = id
                                showingJoinedGame = true
                            }
                        }
                    } label: {
                        HStack {
                            if model.isWorking { ProgressView() }
                            Text("Join game")
                                .fontWeight(.semibold)
                        }
                        .frame(maxWidth: .infinity)
                    }
                    .disabled(parsedGameID == nil || model.isWorking)
                    .accessibilityIdentifier("join.submit")
                }
            }
            .scrollContentBackground(.hidden)
        }
        .navigationTitle("Join")
        .navigationDestination(isPresented: $showingJoinedGame) {
            GameView(gameID: joinedGameID)
        }
    }
}
