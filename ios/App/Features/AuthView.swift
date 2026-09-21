import SwiftUI

struct AuthView: View {
    @EnvironmentObject private var model: AppModel
    @State private var creatingAccount = false
    @State private var username = ""
    @State private var password = ""

    private var canSubmit: Bool {
        !username.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
            password.count >= 8 &&
            !model.isWorking
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 6) {
                    Text("vedh")
                        .font(.system(size: 48, weight: .black, design: .rounded))
                        .foregroundStyle(Color.vedhParchment)
                    Text("Keep multiplayer board state legible.")
                        .foregroundStyle(Color.vedhMuted)
                }

                VStack(spacing: 16) {
                    Picker("Authentication mode", selection: $creatingAccount) {
                        Text("Sign in").tag(false)
                        Text("Create account").tag(true)
                    }
                    .pickerStyle(.segmented)

                    TextField("Username", text: $username)
                        .textContentType(.username)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .accessibilityIdentifier("auth.username")

                    SecureField("Password", text: $password)
                        .textContentType(creatingAccount ? .newPassword : .password)
                        .accessibilityIdentifier("auth.password")

                    if creatingAccount {
                        Text("Use at least 8 characters.")
                            .font(.caption)
                            .foregroundStyle(Color.vedhMuted)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }

                    Button {
                        Task {
                            await model.authenticate(
                                username: username.trimmingCharacters(in: .whitespacesAndNewlines),
                                password: password,
                                creatingAccount: creatingAccount
                            )
                        }
                    } label: {
                        HStack {
                            if model.isWorking { ProgressView() }
                            Text(creatingAccount ? "Create account" : "Sign in")
                                .fontWeight(.semibold)
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 6)
                    }
                    .buttonStyle(.borderedProminent)
                    .tint(.vedhBrass)
                    .foregroundStyle(Color.vedhTable)
                    .disabled(!canSubmit)
                    .accessibilityIdentifier("auth.submit")
                }
                .textFieldStyle(.roundedBorder)
                .vedhPanel()

                Text("Tokens are stored in the iOS Keychain and sent only as Bearer authorization to the configured vedh GraphQL endpoint.")
                    .font(.caption)
                    .foregroundStyle(Color.vedhMuted)
            }
            .frame(maxWidth: 520)
            .padding(24)
            .frame(maxWidth: .infinity)
        }
        .scrollDismissesKeyboard(.interactively)
    }
}
