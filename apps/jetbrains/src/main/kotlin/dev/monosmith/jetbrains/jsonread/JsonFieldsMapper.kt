package dev.monosmith.jetbrains.jsonread

import com.google.gson.JsonArray
import com.google.gson.JsonElement
import com.google.gson.JsonObject
import com.google.gson.JsonParser
import com.google.gson.JsonPrimitive

/** Reads the fields of the JSON smith prints, leaving a missing or mistyped field empty instead of throwing. */
object JsonFieldsMapper {
    /**
     * Parses [json] as an object.
     *
     * @param json what smith printed
     * @return the object
     * @throws IllegalStateException when the text is not a JSON object
     * @throws com.google.gson.JsonSyntaxException when the text is not JSON
     */
    fun parseObject(json: String): JsonObject = JsonParser.parseString(json).asJsonObject

    /**
     * Parses [json] as a list.
     *
     * @param json what smith printed
     * @return the list
     * @throws IllegalStateException when the text is not a JSON array
     * @throws com.google.gson.JsonSyntaxException when the text is not JSON
     */
    fun parseArray(json: String): JsonArray = JsonParser.parseString(json).asJsonArray

    /** The string at [key], or an empty text. */
    fun text(
        record: JsonObject,
        key: String,
    ): String = primitive(record, key)?.takeIf { it.isString }?.asString.orEmpty()

    /** The boolean at [key]; false when it is missing. */
    fun flag(
        record: JsonObject,
        key: String,
    ): Boolean = primitive(record, key)?.takeIf { it.isBoolean }?.asBoolean ?: false

    /** The number at [key]; 0 when it is missing. */
    fun number(
        record: JsonObject,
        key: String,
    ): Int = primitive(record, key)?.takeIf { it.isNumber }?.asInt ?: 0

    /** The strings of the list at [key]. */
    fun words(
        record: JsonObject,
        key: String,
    ): List<String> = list(record, key).filter { it.isJsonPrimitive && it.asJsonPrimitive.isString }.map { it.asString }

    /** The objects of the list at [key]. */
    fun records(
        record: JsonObject,
        key: String,
    ): List<JsonObject> = list(record, key).filter { it.isJsonObject }.map { it.asJsonObject }

    /** The object at [key], or null. */
    fun record(
        record: JsonObject,
        key: String,
    ): JsonObject? = record.get(key)?.takeIf { it.isJsonObject }?.asJsonObject

    private fun list(
        record: JsonObject,
        key: String,
    ): List<JsonElement> =
        record
            .get(key)
            ?.takeIf { it.isJsonArray }
            ?.asJsonArray
            ?.toList()
            .orEmpty()

    private fun primitive(
        record: JsonObject,
        key: String,
    ): JsonPrimitive? = record.get(key)?.takeIf { it.isJsonPrimitive }?.asJsonPrimitive
}
