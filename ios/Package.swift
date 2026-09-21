// swift-tools-version: 5.10

import PackageDescription

let package = Package(
    name: "VedhCore",
    platforms: [
        .iOS(.v17),
        .macOS(.v13),
    ],
    products: [
        .library(name: "VedhCore", targets: ["VedhCore"]),
        .executable(name: "VedhCoreCheck", targets: ["VedhCoreCheck"]),
    ],
    targets: [
        .target(
            name: "VedhCore",
            path: "Sources/VedhCore"
        ),
        .executableTarget(
            name: "VedhCoreCheck",
            dependencies: ["VedhCore"],
            path: "Checks"
        ),
        .testTarget(
            name: "VedhCoreTests",
            dependencies: ["VedhCore"],
            path: "Tests/VedhCoreTests"
        ),
    ],
    swiftLanguageVersions: [.v5]
)
