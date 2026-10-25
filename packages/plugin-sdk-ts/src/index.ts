// The public API of @monosmith/plugin-sdk.

// The protocol: every message as a type, generated from the JSON Schema.
export * from './protocol/index.ts'

// Writing a plugin.
export * from './serve/index.ts'
export * from './concurrency/index.ts'
export * from './tasks/index.ts'
export * from './scaffold/index.ts'
export * from './ci/index.ts'
export * from './build-dir/index.ts'
export * from './plugin-roles/index.ts'
export * from './plugin-refs/index.ts'
export * from './projects/index.ts'
export * from './templating/index.ts'
export * from './jsondoc/index.ts'

// Testing a plugin.
export * from './host/index.ts'
export * from './conformance/index.ts'

// Building blocks: the wire form of the messages, for tools that speak the protocol themselves.
export { decodeParams, decodeResult, encodeParams, encodeResult, EncodeError, MessageParseError, ParamsError } from './codec/index.ts'
export { FrameReader, FramingError, frameMessage } from './framing/index.ts'
