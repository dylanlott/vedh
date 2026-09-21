import Foundation

public enum InviteParser {
    public static func gameID(from input: String) -> String? {
        let trimmed = input.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        if let url = URL(string: trimmed) {
            let parts = url.pathComponents.filter { $0 != "/" }
            if let marker = parts.lastIndex(where: { $0 == "join" || $0 == "games" }), marker + 1 < parts.count {
                return validate(parts[marker + 1])
            }
        }

        return validate(trimmed)
    }

    private static func validate(_ candidate: String) -> String? {
        guard candidate.count >= 6 else { return nil }
        let allowed = CharacterSet.alphanumerics.union(CharacterSet(charactersIn: "-"))
        return candidate.unicodeScalars.allSatisfy(allowed.contains) ? candidate : nil
    }
}
