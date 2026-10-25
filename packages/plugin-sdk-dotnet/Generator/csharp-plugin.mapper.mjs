// The C# of Serve/PluginContract.cs (the interfaces a plugin implements) and Serve/PluginMethodsContract.cs
// (which interface answers which method).

import { PLUGIN_INTERFACES, SERVE_LOOP_METHODS } from './plugin-api.config.mjs'
import { GENERATED_HEADER, literal, typeName } from './csharp-names.algorithm.mjs'
import { methodConstant } from './csharp-types.mapper.mjs'
import { docComment } from './xml-doc.algorithm.mjs'

const definitionOf = (node) => typeName(node.$ref.slice('#/$defs/'.length))

/** The methods of the schema that a plugin answers, each with its interface, member and message types. */
export function pluginMethods(schema) {
  const rows = []
  for (const [method, row] of Object.entries(schema['x-methods'])) {
    if (SERVE_LOOP_METHODS.includes(method)) continue
    const entry = PLUGIN_INTERFACES.find((candidate) => method in candidate.methods)
    if (entry === undefined) {
      throw new Error(`the method ${method} of the schema has no interface in Generator/plugin-api.config.mjs`)
    }
    if (entry.capability !== row.capability) {
      throw new Error(`the method ${method} is advertised as ${row.capability} by the schema and as ${entry.capability} by the config`)
    }
    rows.push({ method, row, entry, member: entry.methods[method], params: definitionOf(row.params), result: definitionOf(row.result) })
  }
  for (const entry of PLUGIN_INTERFACES) {
    for (const method of Object.keys(entry.methods)) {
      if (schema['x-methods'][method] === undefined) throw new Error(`${entry.name} names the method ${method}, which the schema does not have`)
    }
  }
  const capabilities = new Set(Object.values(schema['x-methods']).map((row) => row.capability).filter((name) => name !== null && name !== 'cancel'))
  for (const name of capabilities) {
    if (!PLUGIN_INTERFACES.some((entry) => entry.capability === name)) throw new Error(`the capability ${name} has no interface`)
  }
  return rows
}

/** The C# source of Serve/PluginContract.cs. */
export function mapSchemaToPluginContract(schema) {
  const rows = pluginMethods(schema)
  const interfaces = PLUGIN_INTERFACES.map((entry) => {
    const members = rows
      .filter((row) => row.entry === entry)
      .map(
        (row) =>
          `${docComment(row.row.description, '    ')}    /// <param name="parameters">The params of the request.</param>\n` +
          '    /// <param name="cancellation">Cancelled when the host cancels the request (a plugin that sets <see cref="IPlugin.Cancellable"/>) or the plugin shuts down.</param>\n' +
          `    /// <returns>The result of the method.</returns>\n    Task<${row.result}> ${row.member}(${row.params} parameters, CancellationToken cancellation);\n`,
      )
    return `/// <summary>${entry.summary}</summary>\npublic interface ${entry.name}\n{\n${members.join('\n')}}\n`
  })
  return (
    `${GENERATED_HEADER}\n` +
    'using MonoSmith.Plugin.Sdk.Protocol;\n\nnamespace MonoSmith.Plugin.Sdk;\n\n' +
    '/// <summary>The one interface every plugin implements.</summary>\n' +
    '/// <remarks>\n/// <para>A plugin is a class that implements <see cref="IPlugin"/> and one interface for each thing it does; <c>PluginServer.Run</c> advertises exactly the capabilities of the interfaces it implements.</para>\n' +
    '/// </remarks>\n' +
    'public interface IPlugin\n{\n' +
    '    /// <summary>Whether the plugin stops the work of a request when the host cancels it. A plugin that says so is advertised with the <c>cancel</c> capability and sent the <c>cancel</c> notification, and the <see cref="CancellationToken"/> of the request is cancelled; one that does not finishes what it is doing.</summary>\n' +
    '    bool Cancellable => false;\n\n' +
    '    /// <summary>Answers <c>initialize</c>: the plugin\'s identity, kind, languages, manifests and roles. The serve loop fills in <c>protocolVersion</c> and <c>capabilities</c>.</summary>\n' +
    '    /// <param name="parameters">What the host said about itself.</param>\n    /// <returns>What the plugin is.</returns>\n' +
    '    PluginDescription Describe(InitializeParams parameters);\n}\n\n' +
    `${interfaces.join('\n')}`
  )
}

/** The C# source of Serve/PluginMethodsContract.cs. */
export function mapSchemaToPluginMethods(schema) {
  const rows = pluginMethods(schema)
  const entries = rows.map(
    (row) =>
      `        [ProtocolMethods.${methodConstant(row.method)}] = new(\n` +
      `            ProtocolMethods.${methodConstant(row.method)},\n` +
      `            plugin => plugin is ${row.entry.name},\n` +
      `            value => ${row.params}.Read(value, "params"),\n` +
      `            async (plugin, parameters, cancellation) => await ((${row.entry.name})plugin).${row.member}((${row.params})parameters, cancellation)),`,
  )
  const capabilities = PLUGIN_INTERFACES.map((entry) => `            ${capabilityMember(entry.capability)} = plugin is ${entry.name},`)
  return (
    `${GENERATED_HEADER}\n` +
    'using MonoSmith.Plugin.Sdk.Protocol;\n\nnamespace MonoSmith.Plugin.Sdk;\n\n' +
    'internal static class PluginMethods\n{\n' +
    '    /// <summary>The methods a plugin answers, by name.</summary>\n' +
    '    internal static IReadOnlyDictionary<string, PluginMethod> ByName { get; } = new Dictionary<string, PluginMethod>(StringComparer.Ordinal)\n    {\n' +
    `${entries.join('\n')}\n    };\n\n` +
    '    /// <summary>The capabilities a plugin advertises, derived from the interfaces it implements.</summary>\n' +
    '    internal static Capabilities CapabilitiesOf(IPlugin plugin) =>\n        new()\n        {\n' +
    `${capabilities.join('\n')}\n            Cancel = plugin.Cancellable,\n        };\n}\n`
  )
}

function capabilityMember(capability) {
  return capability[0].toUpperCase() + capability.slice(1)
}

export { literal }
