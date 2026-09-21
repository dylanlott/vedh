import XCTest
@testable import VedhCore

final class GameRulesTests: XCTestCase {
    func testNormalizesLegacyPhaseNames() {
        XCTAssertEqual(GameRules.normalizedPhase("main"), "MAIN PHASE 1")
        XCTAssertEqual(GameRules.normalizedPhase(" main   phase 2 "), "MAIN PHASE 2")
        XCTAssertEqual(GameRules.normalizedPhase("end"), "END STEP")
    }

    func testAdvancesAndWrapsCommanderPhases() {
        XCTAssertEqual(GameRules.nextCommanderPhase(after: "UPKEEP"), "DRAW")
        XCTAssertEqual(GameRules.nextCommanderPhase(after: "DISCARD"), "UNTAP")
        XCTAssertEqual(GameRules.nextCommanderPhase(after: nil), "UNTAP")
    }

    func testFindsNextPriorityPlayer() {
        let players = [
            GamePlayer(id: "1", username: "Ari"),
            GamePlayer(id: "2", username: "Bo"),
            GamePlayer(id: "3", username: "Cy"),
        ]
        XCTAssertEqual(GameRules.nextPriorityPlayer(current: "Ari", players: players), "Bo")
        XCTAssertEqual(GameRules.nextPriorityPlayer(current: "Cy", players: players), "Ari")
        XCTAssertEqual(GameRules.nextPriorityPlayer(current: "Unknown", players: players), "Ari")
    }
}
