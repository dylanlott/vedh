import Foundation
import XCTest
@testable import VedhCore

final class ModelCodingTests: XCTestCase {
    func testDecodesGameDetailUsingServerFieldNames() throws {
        let json = #"""
        {
          "ID": "game-123",
          "CreatedAt": "2026-09-20T12:00:00Z",
          "Status": "IN_PROGRESS",
          "Result": null,
          "WinnerIDs": [],
          "WinCondition": null,
          "PendingWinClaim": null,
          "Rules": [{"Name":"format","Value":"EDH"}],
          "Turn": {"Player":"Dylan","Phase":"MAIN","Number":1,"Priority":"Dylan"},
          "Stack": [],
          "Players": [{
            "ID":"user-1",
            "Username":"Dylan",
            "Boardstate": {
              "UserID":"user-1",
              "User":"Dylan",
              "GameID":"game-123",
              "Life":40,
              "Commander":[{"ID":"card-1","Name":"Atraxa"}],
              "Library":[],
              "Graveyard":[],
              "Exiled":[],
              "Battlefield":[],
              "Hand":[],
              "Revealed":[],
              "Controlled":[],
              "Counters":[]
            }
          }]
        }
        """#

        let game = try JSONDecoder().decode(GameDetail.self, from: Data(json.utf8))
        XCTAssertEqual(game.id, "game-123")
        XCTAssertEqual(game.players.first?.boardState?.life, 40)
        XCTAssertEqual(game.players.first?.boardState?.commander.first?.name, "Atraxa")
        XCTAssertEqual(GameRules.nextCommanderPhase(after: game.turn?.phase), "COMBAT")
    }

    func testBoardStateInputRetainsAllZonesWhenChangingLife() throws {
        let card = GameCard(id: "card-1", name: "Sol Ring", tapped: true, types: "Artifact")
        let board = BoardState(
            userID: "user-1",
            user: "Dylan",
            gameID: "game-1",
            life: 40,
            battlefield: [card]
        )

        let data = try JSONEncoder().encode(BoardStateInput(boardState: board, life: 39))
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        XCTAssertEqual(object["Life"] as? Int, 39)
        let battlefield = try XCTUnwrap(object["Battlefield"] as? [[String: Any]])
        XCTAssertEqual(battlefield.first?["Name"] as? String, "Sol Ring")
        XCTAssertEqual(battlefield.first?["Tapped"] as? Bool, true)
        XCTAssertNotNil(object["Commander"])
        XCTAssertNotNil(object["Library"])
        XCTAssertNotNil(object["Hand"])
    }
}
