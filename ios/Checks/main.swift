import Foundation
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif
import VedhCore

enum CheckFailure: Error, CustomStringConvertible {
    case failed(String)

    var description: String {
        switch self {
        case let .failed(message): return message
        }
    }
}

@main
struct VedhCoreCheck {
    static func main() async throws {
        try checkInviteParsing()
        try checkGameRules()
        try checkServerModelDecoding()
        try checkLifeUpdatePreservesZones()
        try await checkGraphQLClient()
        print("VedhCoreCheck: 5 checks passed")
    }

    private static func checkInviteParsing() throws {
        try expect(
            InviteParser.gameID(from: "https://vedh.xyz/join/abc123-def") == "abc123-def",
            "join URL did not yield its game ID"
        )
        try expect(
            InviteParser.gameID(from: "d53118e8-17ff-4e22-a779-1932943c3851") == "d53118e8-17ff-4e22-a779-1932943c3851",
            "raw game ID was rejected"
        )
        try expect(InviteParser.gameID(from: "not a game") == nil, "invalid invite was accepted")
    }

    private static func checkGameRules() throws {
        try expect(GameRules.normalizedPhase("main") == "MAIN PHASE 1", "MAIN normalization failed")
        try expect(GameRules.nextCommanderPhase(after: "UPKEEP") == "DRAW", "phase advance failed")
        try expect(GameRules.nextCommanderPhase(after: "DISCARD") == "UNTAP", "phase wrap failed")

        let players = [
            GamePlayer(id: "1", username: "Ari"),
            GamePlayer(id: "2", username: "Bo"),
        ]
        try expect(
            GameRules.nextPriorityPlayer(current: "Bo", players: players) == "Ari",
            "priority wrap failed"
        )
    }

    private static func checkServerModelDecoding() throws {
        let json = #"""
        {
          "ID":"game-123","CreatedAt":"2026-09-20T12:00:00Z","Status":"IN_PROGRESS",
          "Result":null,"WinnerIDs":[],"WinCondition":null,"PendingWinClaim":null,
          "Rules":[{"Name":"format","Value":"EDH"}],
          "Turn":{"Player":"Dylan","Phase":"MAIN","Number":1,"Priority":"Dylan"},
          "Stack":[],
          "Players":[{
            "ID":"user-1","Username":"Dylan",
            "Boardstate":{"UserID":"user-1","User":"Dylan","GameID":"game-123","Life":40,
            "Commander":[{"ID":"card-1","Name":"Atraxa"}],"Library":[],"Graveyard":[],
            "Exiled":[],"Battlefield":[],"Hand":[],"Revealed":[],"Controlled":[],"Counters":[]}
          }]
        }
        """#
        let game = try JSONDecoder().decode(GameDetail.self, from: Data(json.utf8))
        try expect(game.id == "game-123", "game ID decode failed")
        try expect(game.players.first?.boardState?.life == 40, "life decode failed")
        try expect(game.players.first?.boardState?.commander.first?.name == "Atraxa", "commander decode failed")
    }

    private static func checkLifeUpdatePreservesZones() throws {
        let card = GameCard(id: "card-1", name: "Sol Ring", tapped: true, types: "Artifact")
        let board = BoardState(
            userID: "user-1",
            user: "Dylan",
            gameID: "game-1",
            life: 40,
            battlefield: [card]
        )
        let data = try JSONEncoder().encode(BoardStateInput(boardState: board, life: 39))
        guard let object = try JSONSerialization.jsonObject(with: data) as? [String: Any] else {
            throw CheckFailure.failed("board-state input did not encode as an object")
        }
        try expect(object["Life"] as? Int == 39, "life override was not encoded")
        let battlefield = object["Battlefield"] as? [[String: Any]]
        try expect(battlefield?.first?["Name"] as? String == "Sol Ring", "battlefield was not preserved")
        try expect(object["Commander"] != nil && object["Hand"] != nil, "empty zones were omitted")
    }

    private static func checkGraphQLClient() async throws {
        struct Variables: Encodable { let username: String }
        struct Viewer: Decodable { let name: String }
        struct Payload: Decodable { let viewer: Viewer }

        URLProtocolCheck.handler = { request in
            try expect(
                request.value(forHTTPHeaderField: "Authorization") == "Bearer check-token",
                "GraphQL request omitted its Bearer token"
            )

            let response = HTTPURLResponse(
                url: request.url!,
                statusCode: 200,
                httpVersion: "HTTP/1.1",
                headerFields: ["Content-Type": "application/json"]
            )!
            return (response, Data(#"{"data":{"viewer":{"name":"Dylan"}}}"#.utf8))
        }

        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [URLProtocolCheck.self]
        let client = GraphQLClient(
            endpoint: URL(string: "https://api.vedh.test/graphql")!,
            session: URLSession(configuration: configuration),
            tokenProvider: { "check-token" }
        )
        let payload: Payload = try await client.execute(
            query: "query Viewer($username: String!) { viewer { name } }",
            variables: Variables(username: "dylan")
        )
        try expect(payload.viewer.name == "Dylan", "GraphQL response did not decode")
        URLProtocolCheck.handler = nil
    }

    private static func expect(_ condition: @autoclosure () -> Bool, _ message: String) throws {
        guard condition() else { throw CheckFailure.failed(message) }
    }
}

private final class URLProtocolCheck: URLProtocol {
    static var handler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = Self.handler else {
            client?.urlProtocol(self, didFailWithError: URLError(.badServerResponse))
            return
        }
        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}
