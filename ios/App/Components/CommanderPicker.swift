import SwiftUI
import VedhCore

struct CommanderPicker: View {
    @EnvironmentObject private var model: AppModel
    @Binding var selection: GameCard?
    @State private var query = ""
    @State private var results: [GameCard] = []
    @State private var isSearching = false
    @State private var searchMessage: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            if let selection {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Commander")
                            .font(.caption)
                            .foregroundStyle(Color.vedhMuted)
                        Text(selection.name)
                            .font(.headline)
                            .foregroundStyle(Color.vedhParchment)
                    }
                    Spacer()
                    Button("Remove", role: .destructive) {
                        self.selection = nil
                    }
                }
            } else {
                Text("Commander")
                    .font(.headline)

                HStack {
                    TextField("Search by card name", text: $query)
                        .textInputAutocapitalization(.words)
                        .autocorrectionDisabled()
                        .onSubmit { search() }
                        .accessibilityIdentifier("commander.search")
                    Button {
                        search()
                    } label: {
                        if isSearching {
                            ProgressView()
                        } else {
                            Image(systemName: "magnifyingglass")
                        }
                    }
                    .disabled(query.trimmingCharacters(in: .whitespacesAndNewlines).count < 2 || isSearching)
                }

                if let searchMessage {
                    Text(searchMessage)
                        .font(.caption)
                        .foregroundStyle(Color.vedhMuted)
                }

                ForEach(results.prefix(8)) { card in
                    Button {
                        selection = card
                        results = []
                        searchMessage = nil
                    } label: {
                        VStack(alignment: .leading, spacing: 3) {
                            Text(card.name)
                                .foregroundStyle(Color.vedhParchment)
                            if let text = card.text, !text.isEmpty {
                                Text(text)
                                    .font(.caption)
                                    .foregroundStyle(Color.vedhMuted)
                                    .lineLimit(2)
                            }
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                    }
                    .buttonStyle(.plain)
                    Divider().overlay(Color.vedhParchment.opacity(0.12))
                }
            }
        }
    }

    private func search() {
        let name = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard name.count >= 2 else { return }
        isSearching = true
        searchMessage = nil
        Task {
            do {
                results = try await model.searchCards(name: name)
                searchMessage = results.isEmpty ? "No cards matched that name." : nil
            } catch {
                results = []
                searchMessage = error.localizedDescription
            }
            isSearching = false
        }
    }
}
