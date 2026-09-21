import XCTest
@testable import VedhCore

final class InviteParserTests: XCTestCase {
    func testParsesRawGameID() {
        XCTAssertEqual(InviteParser.gameID(from: "d53118e8-17ff-4e22-a779-1932943c3851"), "d53118e8-17ff-4e22-a779-1932943c3851")
    }

    func testParsesJoinAndGameURLs() {
        XCTAssertEqual(
            InviteParser.gameID(from: "https://vedh.xyz/join/abc123-def"),
            "abc123-def"
        )
        XCTAssertEqual(
            InviteParser.gameID(from: "https://vedh.xyz/games/abc123-def"),
            "abc123-def"
        )
    }

    func testRejectsInvalidInvite() {
        XCTAssertNil(InviteParser.gameID(from: "short"))
        XCTAssertNil(InviteParser.gameID(from: "not a game id"))
        XCTAssertNil(InviteParser.gameID(from: ""))
    }
}
