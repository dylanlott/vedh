DESTINATION ?= platform=iOS Simulator,name=iPhone 16 Pro

.PHONY: check project simulator-test

check: project
	swift build
	swift run VedhCoreCheck
	swiftc -frontend -parse App/*.swift App/Features/*.swift App/Components/*.swift
	plutil -lint Vedh.xcodeproj/project.pbxproj App/Info.plist

project:
	xcodegen generate

simulator-test: project
	xcodebuild -project Vedh.xcodeproj -scheme Vedh -destination '$(DESTINATION)' test
