// The C# of Protocol/ProtocolTypesContract.cs: one type per definition of the schema, the error codes, the
// methods and the roles.

import { docComment, summaryLine } from './xml-doc.algorithm.mjs'
import { GENERATED_HEADER, camel, literal, pascal, typeName } from './csharp-names.algorithm.mjs'
import { csharpDefault, csharpType, describe } from './schema-type.algorithm.mjs'

// -- reading ------------------------------------------------------------------------------------------

/** The expression that reads a value of a described node. `value` and `path` are expressions. */
function readExpression(desc, value, path, depth = 0) {
  switch (desc.kind) {
    case 'string':
      return `StringValue(${value}, ${path})`
    case 'integer':
      return desc.nullable ? `NullableIntegerValue(${value}, ${path})` : `IntegerValue(${value}, ${path})`
    case 'number':
      return `NumberValue(${value}, ${path})`
    case 'boolean':
      return `BooleanValue(${value}, ${path})`
    case 'enum':
    case 'struct':
      return `${desc.name}.Read(${value}, ${path})`
    case 'object':
      return `ObjectOf(${value}, ${path})`
    case 'array':
      return `ListOf(${value}, ${path}, ${itemReader(desc.items, depth)})`
    case 'map':
      return `MapOf(${value}, ${path}, ${itemReader(desc.values, depth)}${desc.values.kind === 'any' ? ', keepNull: true' : ''})`
    default:
      return `AnyValue(${value}, ${path})`
  }
}

/** A delegate that reads an item: a method group where one exists, a lambda otherwise. */
function itemReader(desc, depth) {
  switch (desc.kind) {
    case 'string':
      return 'StringValue'
    case 'integer':
      return desc.nullable ? 'NullableIntegerValue' : 'IntegerValue'
    case 'number':
      return 'NumberValue'
    case 'boolean':
      return 'BooleanValue'
    case 'enum':
    case 'struct':
      return `${desc.name}.Read`
    case 'object':
      return 'ObjectOf'
    default: {
      const value = `value${depth + 1}`
      const path = `path${depth + 1}`
      return `(${value}, ${path}) => ${readExpression(desc, value, path, depth + 1)}`
    }
  }
}

// -- writing ------------------------------------------------------------------------------------------

/** The statements that write a value of a described node. */
function writeStatements(desc, value, writer, depth = 0) {
  switch (desc.kind) {
    case 'string':
      return [`${writer}.String(${value});`]
    case 'integer':
      return desc.nullable
        ? [`if (${value} is { } number${depth})`, '{', `    ${writer}.Integer(number${depth});`, '}', 'else', '{', `    ${writer}.Null();`, '}']
        : [`${writer}.Integer(${value});`]
    case 'number':
      return [`${writer}.Number(${value});`]
    case 'boolean':
      return [`${writer}.Boolean(${value});`]
    case 'enum':
    case 'struct':
      return [`${value}.Write(${writer});`]
    case 'object':
    case 'any':
      return [`${writer}.Any(${value});`]
    case 'array': {
      const item = `item${depth + 1}`
      const inner = `builder${depth + 1}`
      return [`MessageWritingAlgorithm.List(${writer}, ${value}, ${lambda(inner, item, desc.items, depth + 1)});`]
    }
    default: {
      const item = `member${depth + 1}`
      const inner = `builder${depth + 1}`
      return [`MessageWritingAlgorithm.Map(${writer}, ${value}, ${lambda(inner, item, desc.values, depth + 1)});`]
    }
  }
}

function lambda(writer, item, desc, depth) {
  const statements = writeStatements(desc, item, writer, depth)
  if (statements.length === 1) return `(${writer}, ${item}) => ${statements[0].replace(/;$/, '')}`
  return `(${writer}, ${item}) =>\n            {\n${statements.map((line) => `                ${line}`).join('\n')}\n            }`
}

/** The condition under which an optional member is written: Go's `omitempty`. */
function presence(desc, value) {
  switch (desc.kind) {
    case 'string':
      return `${value}.Length > 0`
    case 'integer':
    case 'number':
      return `${value} != 0`
    case 'boolean':
      return value
    case 'enum':
      return `!${value}.IsEmpty`
    case 'struct':
      return `${value} is not null`
    case 'array':
    case 'map':
    case 'object':
      return `${value} is { Count: > 0 }`
    default:
      return `${value} is not (null or JsonNull)`
  }
}

// -- messages -----------------------------------------------------------------------------------------

function members(schema, node) {
  const required = new Set(node.required)
  return Object.entries(node.properties ?? {}).map(([json, property]) => {
    const desc = describe(schema, property)
    return { json, property, desc, required: required.has(json), name: pascal(json) }
  })
}

function propertyDeclaration(owner, member) {
  const { json, property, desc, required, name } = member
  const type = csharpType(desc, !required)
  const doc = docComment(property.description, "    ", "The `" + json + "` member of `" + owner + "`.")
  const mustSet = required && ['string', 'enum', 'struct'].includes(desc.kind)
  const initial = !mustSet && csharpDefault(desc) !== undefined && !(desc.kind === 'struct' && !required)
  const lead = mustSet ? 'required ' : ''
  const tail = initial ? ` = ${csharpDefault(desc)};` : ''
  return `${doc}    public ${lead}${type} ${name} { get; init; }${tail}\n`
}

function readerOf(name, list) {
  const lines = list.map((member) => {
    const value = `Member(map, ${literal(member.json)})`
    const path = `Join(path, ${literal(member.json)})`
    const expression = !member.required && member.desc.kind === 'struct'
      ? `OptionalStruct<${member.desc.name}>(${value}, ${path})`
      : readExpression(member.desc, value, path)
    return `            ${member.name} = ${expression},`
  })
  const initializer = list.length === 0 ? '' : `\n        {\n${lines.join('\n')}\n        }`
  return (
    '    /// <inheritdoc />\n' +
    `    public static ${name} Read(JsonAny? value, string path)\n    {\n` +
    `        ${list.length === 0 ? '_ = ObjectOf(value, path);' : 'var map = ObjectOf(value, path);'}\n` +
    `        return new ${name}${list.length === 0 ? "()" : initializer};\n    }\n`
  )
}

function writerOf(list) {
  const body = []
  for (const member of list) {
    const value = member.name
    const statements = [`writer.Property(${literal(member.json)});`, ...writeStatements(member.desc, value, 'writer')]
    if (member.required) {
      body.push(...statements.map((line) => `        ${line}`))
    } else {
      body.push(`        if (${presence(member.desc, value)})`, '        {', ...statements.map((line) => `            ${line}`), '        }')
    }
  }
  return (
    '    /// <inheritdoc />\n    public void Write(JsonBuilder writer)\n    {\n        writer.BeginObject();\n' +
    (body.length === 0 ? '' : `${body.join('\n')}\n`) +
    '        writer.EndObject();\n    }\n'
  )
}

/** Members the generator adds to one message. */
function extraMembers(definition, list) {
  if (definition !== 'Capabilities') return []
  const cases = list.map((member) => `            ${literal(member.json)} => ${member.name},`)
  return [
    '    /// <summary>Whether a capability is advertised, by its name in the protocol (<c>discover</c>, <c>cancel</c>).</summary>\n' +
      '    /// <param name="name">The capability.</param>\n' +
      '    /// <returns>Whether the plugin advertised it.</returns>\n' +
      '    public bool Advertises(string name) =>\n        name switch\n        {\n' +
      `${cases.join('\n')}\n            _ => false,\n        };\n`,
  ]
}

function messageType(schema, definition, node) {
  const name = typeName(definition)
  const list = members(schema, node)
  const doc = docComment(node.description, "", "The `" + definition + "` message.")
  return (
    `${doc}public sealed record ${name} : IProtocolValue<${name}>\n{\n` +
    [...list.map((member) => propertyDeclaration(name, member)), readerOf(name, list), writerOf(list), ...extraMembers(definition, list)].join('\n') +
    '}\n'
  )
}

// -- enumerations -------------------------------------------------------------------------------------

function enumType(definition, node) {
  const name = typeName(definition)
  const descriptions = node['x-enumDescriptions'] ?? {}
  const values = node.enum.map((value) => {
    const doc = docComment(descriptions[value], "    ", "The value `" + value + "`.")
    return `${doc}    public static ${name} ${pascal(value)} { get; } = new(${literal(value)});\n`
  })
  const doc = docComment(node.description, "", "The `" + definition + "` strings.")
  return (
    `${doc}/// <remarks>\n/// <para>A string, not a closed set: a newer protocol may add values, and one that is not listed here is kept as it came.</para>\n/// </remarks>\n` +
    `public readonly record struct ${name} : IProtocolValue<${name}>\n{\n` +
    '    private readonly string? _value;\n\n' +
    `    /// <summary>A ${name.replace(/([A-Z])/g, ' $1').trim().toLowerCase()} given as text.</summary>\n    /// <param name="value">The text.</param>\n    public ${name}(string value)\n    {\n        _value = value;\n    }\n\n` +
    `    /// <summary>The text of the value: empty for the default one.</summary>\n    public string Value => _value ?? string.Empty;\n\n` +
    '    /// <summary>Whether the value is empty, which is what an absent member reads as.</summary>\n    public bool IsEmpty => Value.Length == 0;\n\n' +
    `${values.join('\n')}\n` +
    `    /// <summary>A ${name} from text.</summary>\n    /// <param name="value">The text.</param>\n    public static implicit operator ${name}(string value) => new(value);\n\n` +
    `    /// <summary>The text of a ${name}.</summary>\n    /// <param name="value">The value.</param>\n    public static implicit operator string(${name} value) => value.Value;\n\n` +
    `    /// <inheritdoc />\n    public static ${name} Read(JsonAny? value, string path) => new(MessageReadingAlgorithm.StringValue(value, path));\n\n` +
    '    /// <inheritdoc />\n    public void Write(JsonBuilder writer) => writer.String(Value);\n\n' +
    `    /// <summary>Whether another value has the same text.</summary>\n    /// <param name="other">The other value.</param>\n    /// <returns>Whether they are equal.</returns>\n    public bool Equals(${name} other) => string.Equals(Value, other.Value, StringComparison.Ordinal);\n\n` +
    '    /// <inheritdoc />\n    public override int GetHashCode() => StringComparer.Ordinal.GetHashCode(Value);\n\n' +
    '    /// <inheritdoc />\n    public override string ToString() => Value;\n' +
    '}\n'
  )
}

// -- constants ----------------------------------------------------------------------------------------

function errorCodes(schema) {
  const rows = schema['x-errorCodes'].map(
    (row) => `${docComment(row.description, '    ', row.name)}    public const long ${row.name} = ${row.code};\n`,
  )
  return (
    '/// <summary>The error codes of the protocol: the JSON-RPC 2.0 ones, <see cref="PluginFailure"/> for a failure the user must act on and <see cref="RequestCancelled"/> for a cancelled request.</summary>\n' +
    `public static class ErrorCodes\n{\n${rows.join('\n')}}\n`
  )
}

/** The name of the C# constant of a method: `version/write` is `VersionWrite`. */
export function methodConstant(method) {
  return pascal(method)
}

function methods(schema) {
  const entries = Object.entries(schema['x-methods'])
  const constants = entries.map(
    ([method, row]) => `${docComment(row.description, '    ', method)}    public const string ${methodConstant(method)} = ${literal(method)};\n`,
  )
  const capabilities = entries.map(([method, row]) => `        [${methodConstant(method)}] = ${row.capability === null ? 'null' : literal(row.capability)},`)
  return (
    '/// <summary>The names of the methods of the protocol.</summary>\n' +
    `public static class ProtocolMethods\n{\n${constants.join('\n')}\n` +
    '    /// <summary>For each method the capability a plugin advertises to be asked it; <c>null</c> where it is always available.</summary>\n' +
    '    public static IReadOnlyDictionary<string, string?> CapabilityByMethod { get; } = new Dictionary<string, string?>(StringComparer.Ordinal)\n    {\n' +
    `${capabilities.join('\n')}\n    };\n}\n`
  )
}

function roles(schema) {
  const rows = schema['x-roles'].map(
    (row) =>
      `        new(new Role(${literal(row.name)}), ${literal(row.description)}, ${literal(row.cardinality)}, [${row.defaultFor.map(literal).join(', ')}]),`,
  )
  return (
    '/// <summary>A role of the protocol with how many plugins of a project can play it and which kinds of plugin play it by default.</summary>\n' +
    '/// <param name="Name">The role.</param>\n/// <param name="Description">What the role does.</param>\n' +
    '/// <param name="Cardinality"><c>exclusive</c>, <c>perSlot</c> or <c>additive</c>.</param>\n' +
    '/// <param name="DefaultFor">The plugin kinds that play the role when they declare none.</param>\n' +
    'public sealed record RoleInfo(Role Name, string Description, string Cardinality, IReadOnlyList<string> DefaultFor);\n\n' +
    '/// <summary>The roles of the protocol, in the order a project\'s make-up lists them.</summary>\n' +
    `public static class ProtocolRoles\n{\n    /// <summary>Every role.</summary>\n    public static IReadOnlyList<RoleInfo> All { get; } =\n    [\n${rows.join('\n')}\n    ];\n}\n`
  )
}

function info(schema) {
  return (
    '/// <summary>The protocol this SDK speaks.</summary>\npublic static class ProtocolInfo\n{\n' +
    '    /// <summary>The protocol major version this SDK speaks. A breaking change to the protocol raises it.</summary>\n' +
    `    public const long Version = ${schema['x-protocolVersion']};\n\n` +
    '    /// <summary>The largest message the framing accepts, in bytes.</summary>\n' +
    `    public const long MaxMessageBytes = ${schema['x-transport'].maxMessageBytes};\n}\n`
  )
}

// -- the plugin's own description ---------------------------------------------------------------------

function description(schema) {
  const node = schema.$defs.InitializeResult
  const list = members(schema, node).filter((member) => !['protocolVersion', 'capabilities'].includes(member.json))
  const assignments = list.map((member) => `            ${member.name} = ${member.name},`)
  return (
    '/// <summary>What a plugin says about itself: the <c>initialize</c> result without the members the serve loop fills in, <c>protocolVersion</c> and <c>capabilities</c> (derived from the interfaces the plugin implements).</summary>\n' +
    `public sealed record PluginDescription\n{\n${list.map((member) => propertyDeclaration('PluginDescription', member)).join('\n')}\n` +
    '    /// <summary>The <c>initialize</c> result for this description.</summary>\n    /// <param name="capabilities">The capabilities derived from the plugin.</param>\n    /// <returns>The result the serve loop answers.</returns>\n' +
    '    public InitializeResult ToInitializeResult(Capabilities capabilities) =>\n        new()\n        {\n            ProtocolVersion = ProtocolInfo.Version,\n            Capabilities = capabilities,\n' +
    `${assignments.join('\n')}\n        };\n}\n`
  )
}

/** The C# source of Protocol/ProtocolTypesContract.cs. */
export function mapSchemaToTypes(schema) {
  const types = Object.entries(schema.$defs).map(([definition, node]) =>
    node.enum !== undefined ? enumType(definition, node) : messageType(schema, definition, node),
  )
  return (
    `${GENERATED_HEADER}\n` +
    'using static MonoSmith.Plugin.Sdk.Protocol.MessageReadingAlgorithm;\n\n' +
    'namespace MonoSmith.Plugin.Sdk.Protocol;\n\n' +
    [info(schema), errorCodes(schema), methods(schema), roles(schema), ...types, description(schema)].join('\n')
  )
}

export { camel, summaryLine }
