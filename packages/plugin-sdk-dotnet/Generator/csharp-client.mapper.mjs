// The C# of Host/PluginClientCallsContract.cs: one typed call on the test host for each method of the protocol.

import { GENERATED_HEADER } from './csharp-names.algorithm.mjs'
import { methodConstant } from './csharp-types.mapper.mjs'
import { pluginMethods } from './csharp-plugin.mapper.mjs'
import { docComment } from './xml-doc.algorithm.mjs'

/** The C# source of Host/PluginClientCallsContract.cs. */
export function mapSchemaToClientCalls(schema) {
  const calls = pluginMethods(schema).map(
    (row) =>
      `${docComment(row.row.description, '    ')}    /// <param name="client">The connected plugin.</param>\n` +
      '    /// <param name="parameters">The params of the request.</param>\n' +
      '    /// <param name="cancellation">Cancels the call.</param>\n' +
      '    /// <param name="timeout">How long the call may take; the default is two minutes.</param>\n' +
      '    /// <returns>The result of the method.</returns>\n' +
      `    public static Task<${row.result}> ${row.member}(\n        this PluginClient client,\n        ${row.params} parameters,\n        CancellationToken cancellation = default,\n        TimeSpan? timeout = null)\n    {\n` +
      '        ArgumentNullException.ThrowIfNull(client);\n' +
      `        return client.CallAsync<${row.params}, ${row.result}>(ProtocolMethods.${methodConstant(row.method)}, parameters, timeout, cancellation);\n    }\n`,
  )
  return (
    `${GENERATED_HEADER}\n` +
    'using MonoSmith.Plugin.Sdk.Protocol;\n\nnamespace MonoSmith.Plugin.Sdk.Testing;\n\n' +
    '/// <summary>One typed call for each method of the protocol a plugin answers.</summary>\n' +
    `public static class PluginClientCalls\n{\n${calls.join('\n')}}\n`
  )
}
