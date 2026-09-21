import Foundation
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif
import XCTest
@testable import VedhCore

final class GraphQLClientTests: XCTestCase {
    override func tearDown() {
        URLProtocolStub.handler = nil
        super.tearDown()
    }

    func testSendsBearerTokenAndDecodesPayload() async throws {
        struct Variables: Encodable { let username: String }
        struct User: Decodable, Equatable { let name: String }
        struct Payload: Decodable, Equatable { let viewer: User }

        URLProtocolStub.handler = { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer test-token")
            XCTAssertEqual(request.value(forHTTPHeaderField: "Content-Type"), "application/json")
            XCTAssertNotNil(request.value(forHTTPHeaderField: "X-Request-Id"))

            let body = try XCTUnwrap(request.capturedBody())
            let object = try XCTUnwrap(JSONSerialization.jsonObject(with: body) as? [String: Any])
            let variables = try XCTUnwrap(object["variables"] as? [String: Any])
            XCTAssertEqual(variables["username"] as? String, "dylan")

            return Self.response(for: request, status: 200, json: #"{"data":{"viewer":{"name":"Dylan"}}}"#)
        }

        let client = makeClient(token: "test-token")
        let payload: Payload = try await client.execute(
            query: "query Viewer($username: String!) { viewer { name } }",
            variables: Variables(username: "dylan")
        )
        XCTAssertEqual(payload, Payload(viewer: User(name: "Dylan")))
    }

    func testSurfacesGraphQLErrors() async throws {
        struct Variables: Encodable {}
        struct Payload: Decodable { let value: String }

        URLProtocolStub.handler = { request in
            Self.response(
                for: request,
                status: 200,
                json: #"{"errors":[{"message":"authentication required"}]}"#
            )
        }

        do {
            let _: Payload = try await makeClient().execute(query: "query { value }", variables: Variables())
            XCTFail("Expected the GraphQL error to be thrown")
        } catch let error as GraphQLClientError {
            XCTAssertEqual(error, .graphQL(["authentication required"]))
        }
    }

    func testSurfacesHTTPStatus() async throws {
        struct Variables: Encodable {}
        struct Payload: Decodable { let value: String }

        URLProtocolStub.handler = { request in
            Self.response(for: request, status: 401, json: "unauthorized")
        }

        do {
            let _: Payload = try await makeClient().execute(query: "query { value }", variables: Variables())
            XCTFail("Expected the HTTP error to be thrown")
        } catch let error as GraphQLClientError {
            XCTAssertEqual(error, .httpStatus(401))
        }
    }

    private func makeClient(token: String? = nil) -> GraphQLClient {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [URLProtocolStub.self]
        return GraphQLClient(
            endpoint: URL(string: "https://api.vedh.test/graphql")!,
            session: URLSession(configuration: configuration),
            tokenProvider: { token }
        )
    }

    private static func response(
        for request: URLRequest,
        status: Int,
        json: String
    ) -> (HTTPURLResponse, Data) {
        let response = HTTPURLResponse(
            url: request.url!,
            statusCode: status,
            httpVersion: "HTTP/1.1",
            headerFields: ["Content-Type": "application/json"]
        )!
        return (response, Data(json.utf8))
    }
}

private final class URLProtocolStub: URLProtocol {
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

private extension URLRequest {
    func capturedBody() -> Data? {
        if let httpBody { return httpBody }
        guard let httpBodyStream else { return nil }
        httpBodyStream.open()
        defer { httpBodyStream.close() }

        var data = Data()
        let bufferSize = 4_096
        let buffer = UnsafeMutablePointer<UInt8>.allocate(capacity: bufferSize)
        defer { buffer.deallocate() }
        while httpBodyStream.hasBytesAvailable {
            let count = httpBodyStream.read(buffer, maxLength: bufferSize)
            if count <= 0 { break }
            data.append(buffer, count: count)
        }
        return data
    }
}
