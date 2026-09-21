import SwiftUI

extension Color {
    static let vedhTable = Color(red: 0.12, green: 0.105, blue: 0.095)
    static let vedhPanel = Color(red: 0.20, green: 0.18, blue: 0.17)
    static let vedhRaised = Color(red: 0.27, green: 0.235, blue: 0.215)
    static let vedhParchment = Color(red: 0.96, green: 0.90, blue: 0.79)
    static let vedhBrass = Color(red: 0.78, green: 0.58, blue: 0.25)
    static let vedhMuted = Color(red: 0.72, green: 0.68, blue: 0.63)
    static let vedhJewel = Color(red: 0.35, green: 0.62, blue: 0.56)
    static let vedhDanger = Color(red: 0.72, green: 0.25, blue: 0.23)
}

struct VedhBackground: View {
    var body: some View {
        Color.vedhTable.ignoresSafeArea()
    }
}

struct PanelModifier: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding(16)
            .background(Color.vedhPanel)
            .overlay(
                RoundedRectangle(cornerRadius: 14)
                    .stroke(Color.vedhParchment.opacity(0.12), lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: 14))
    }
}

extension View {
    func vedhPanel() -> some View {
        modifier(PanelModifier())
    }
}
