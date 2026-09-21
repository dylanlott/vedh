import Foundation
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif

public struct GraphQLIssue: Codable, Equatable, Sendable {
    public let message: String

    public init(message: String) {
        self.message = message
    }
}

public enum GraphQLClientError: LocalizedError, Equatable, Sendable {
    case invalidResponse
    case httpStatus(Int)
    case graphQL([String])
    case missingData
    case invalidPayload(String)

    public var errorDescription: String? {
        switch self {
        case .invalidResponse:
            return "The server returned an invalid response."
        case let .httpStatus(status):
            return "The server returned HTTP \(status)."
        case let .graphQL(messages):
            return messages.joined(separator: "\n")
        case .missingData:
            return "The server response did not contain data."
        case let .invalidPayload(message):
            return "The server response could not be decoded: \(message)"
        }
    }
}

private struct GraphQLRequestBody<Variables: Encodable>: Encodable {
    let query: String
    let variables: Variables
}

private struct GraphQLResponseBody<Payload: Decodable>: Decodable {
    let data: Payload?
    let errors: [GraphQLIssue]?
}

public final class GraphQLClient: @unchecked Sendable {
    public typealias TokenProvider = @Sendable () -> String?

    public let endpoint: URL
    private let session: URLSession
    private let tokenProvider: TokenProvider
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder

    public init(
        endpoint: URL,
        session: URLSession = .shared,
        tokenProvider: @escaping TokenProvider = { nil }
    ) {
        self.endpoint = endpoint
        self.session = session
        self.tokenProvider = tokenProvider
        self.encoder = JSONEncoder()
        self.decoder = JSONDecoder()
    }

    public func execute<Payload: Decodable, Variables: Encodable>(
        query: String,
        variables: Variables,
        as payloadType: Payload.Type = Payload.self
    ) async throws -> Payload {
        var request = URLRequest(url: endpoint)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(UUID().uuidString, forHTTPHeaderField: "X-Request-Id")

        if let token = tokenProvider()?.trimmingCharacters(in: .whitespacesAndNewlines), !token.isEmpty {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        request.httpBody = try encoder.encode(GraphQLRequestBody(query: query, variables: variables))

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw GraphQLClientError.invalidResponse
        }
        guard (200..<300).contains(http.statusCode) else {
            throw GraphQLClientError.httpStatus(http.statusCode)
        }

        let body: GraphQLResponseBody<Payload>
        do {
            body = try decoder.decode(GraphQLResponseBody<Payload>.self, from: data)
        } catch {
            throw GraphQLClientError.invalidPayload(error.localizedDescription)
        }

        if let errors = body.errors, !errors.isEmpty {
            throw GraphQLClientError.graphQL(errors.map(\.message))
        }
        guard let payload = body.data else {
            throw GraphQLClientError.missingData
        }
        return payload
    }
}
