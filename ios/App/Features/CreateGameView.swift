import SwiftUI
import VedhCore

struct CreateGameView: View {
    @EnvironmentObject private var model: AppModel
    @Environment(\.dismiss) private var dismiss
    @State private var decklist = ""
    @State private var commander: GameCard?

    var body: some View {
        ZStack {
            VedhBackground()
            Form {
                Section {
                    Text("A Commander table starts at 40 life with you holding the first priority.")
                        .foregroundStyle(Color.vedhMuted)
                }

                Section {
                    CommanderPicker(selection: $commander)
                }

                Section("Decklist (optional)") {
                    TextEditor(text: $decklist)
                        .font(.body.monospaced())
                        .frame(minHeight: 180)
                        .accessibilityIdentifier("create.decklist")
                    Text("Paste one card per line or CSV in the same format used by the web app.")
                        .font(.caption)
                        .foregroundStyle(Color.vedhMuted)
                }
            }
            .scrollContentBackground(.hidden)
        }
        .navigationTitle("New game")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") { dismiss() }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button {
                    Task {
                        if await model.createGame(decklist: decklist, commander: commander) != nil {
                            dismiss()
                        }
                    }
                } label: {
                    if model.isWorking { ProgressView() } else { Text("Create") }
                }
                .disabled(model.isWorking)
                .accessibilityIdentifier("create.submit")
            }
        }
    }
}
