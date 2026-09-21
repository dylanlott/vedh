import Foundation

public enum GameRules {
    public static let commanderPhases = [
        "UNTAP",
        "UPKEEP",
        "DRAW",
        "MAIN PHASE 1",
        "COMBAT",
        "MAIN PHASE 2",
        "END STEP",
        "DISCARD",
    ]

    public static func normalizedPhase(_ phase: String?) -> String? {
        guard let phase else { return nil }
        let cleaned = phase
            .uppercased()
            .split(whereSeparator: \.isWhitespace)
            .joined(separator: " ")

        switch cleaned {
        case "MAIN", "MAIN PHASE", "MAIN1", "MAIN PHASE 1":
            return "MAIN PHASE 1"
        case "MAIN2", "MAIN PHASE 2":
            return "MAIN PHASE 2"
        case "END", "END STEP":
            return "END STEP"
        default:
            return cleaned
        }
    }

    public static func nextCommanderPhase(after phase: String?) -> String {
        guard
            let normalized = normalizedPhase(phase),
            let index = commanderPhases.firstIndex(of: normalized)
        else {
            return commanderPhases[0]
        }
        return commanderPhases[(index + 1) % commanderPhases.count]
    }

    public static func nextPriorityPlayer(current: String?, players: [GamePlayer]) -> String? {
        let names = players.map(\.username).filter { !$0.isEmpty }
        guard !names.isEmpty else { return nil }
        guard let current, let index = names.firstIndex(of: current) else {
            return names[0]
        }
        return names[(index + 1) % names.count]
    }
}
