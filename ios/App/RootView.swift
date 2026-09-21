import SwiftUI

struct RootView: View {
    @EnvironmentObject private var model: AppModel

    var body: some View {
        ZStack {
            VedhBackground()
            if model.session == nil {
                AuthView()
            } else {
                authenticatedApp
            }
        }
        .tint(.vedhBrass)
        .alert(
            "vedh",
            isPresented: Binding(
                get: { model.errorMessage != nil },
                set: { if !$0 { model.errorMessage = nil } }
            )
        ) {
            Button("Dismiss", role: .cancel) { model.errorMessage = nil }
        } message: {
            Text(model.errorMessage ?? "Unknown error")
        }
    }

    private var authenticatedApp: some View {
        TabView {
            NavigationStack {
                GamesView()
            }
            .tabItem { Label("Games", systemImage: "rectangle.stack") }

            NavigationStack {
                JoinGameView()
            }
            .tabItem { Label("Join", systemImage: "person.badge.plus") }

            NavigationStack {
                AccountView()
            }
            .tabItem { Label("Account", systemImage: "person.crop.circle") }
        }
        .toolbarBackground(Color.vedhPanel, for: .tabBar)
        .toolbarBackground(.visible, for: .tabBar)
    }
}

private struct AccountView: View {
    @EnvironmentObject private var model: AppModel

    var body: some View {
        ZStack {
            VedhBackground()
            Form {
                Section("Signed in") {
                    LabeledContent("Player", value: model.session?.username ?? "—")
                    LabeledContent("User ID", value: model.session?.id ?? "—")
                        .fontDesign(.monospaced)
                }
                Section {
                    Button("Sign out", role: .destructive) { model.logout() }
                }
            }
            .scrollContentBackground(.hidden)
        }
        .navigationTitle("Account")
    }
}
