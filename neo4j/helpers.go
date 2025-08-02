package neo4j

import (
	"fmt"
	"maps"
	"slices"
)

func MergeCountryCypher(countryCode, country string, props map[string]any) (string, string, map[string]any) {
	key := countryCode
	nameKey := fmt.Sprintf("%s_name", key)
	cypher := fmt.Sprintf("MERGE (%s:Country {name: $%s})\nSET ", key, nameKey)

	rtnProps := make(map[string]any, len(props)+1)
	rtnProps[nameKey] = country
	for _, k := range slices.Sorted(maps.Keys(props)) {
		propKey := fmt.Sprintf("%s_%s", key, k)
		rtnProps[propKey] = props[k]
		cypher += fmt.Sprintf("%s.%s = $%s, ", key, k, propKey)
	}

	cypher = cypher[:len(cypher)-2]
	cypher += "\nWITH * \n"
	return key, cypher, rtnProps
}

func MergeCityWithCountryCypher(idx int64, city, countryKey string, props map[string]any) (string, string, map[string]any) {
	key := fmt.Sprintf("city_%d", idx)
	nameKey := fmt.Sprintf("%s_name", key)
	cypher := fmt.Sprintf("MERGE (%s)-[:LOCATED_IN]->(%s)\n", key, countryKey)
	cypher = fmt.Sprintf("MERGE (%s:City {name: $%s})\nSET ", key, nameKey)

	rtnProps := make(map[string]any, len(props)+1)
	rtnProps[nameKey] = city
	for _, k := range slices.Sorted(maps.Keys(props)) {
		propKey := fmt.Sprintf("%s_%s", key, k)
		rtnProps[propKey] = props[k]
		cypher += fmt.Sprintf("%s.%s = $%s, ", key, k, propKey)
	}
	cypher = cypher[:len(cypher)-2]
	cypher += "\nWITH * \n"
	return key, cypher, rtnProps
}

func MergeIPAddressCypher(idx int64, ipAddress string) (string, string) {
	key := fmt.Sprintf("ip_%d", idx)
	cypher := fmt.Sprintf(`MERGE (%s:IPAddress {address: "%s"})
WITH *
`, key, ipAddress)
	return key, cypher
}

func MergeSimpleRelationship(fromKey, relType, toKey string) string {
	return fmt.Sprintf(`MERGE (%s)-[:%s]->(%s)`, fromKey, relType, toKey)
}

func MergeWithIPAddressFromWithCountry(ipAddressKey, countryKey string) string {
	return MergeSimpleRelationship(ipAddressKey, "LOCATED_IN", countryKey)
}

func MergeWithIPAressReportedWithCountry(idx int64, ipAddressKey, countryKey string, props map[string]any) (string, string, map[string]any) {
	key := fmt.Sprintf("reported_%d", idx)
	cypher := fmt.Sprintf("MERGE (%s)-[%s:REPORTED_IN]->(%s)\nSET ", ipAddressKey, key, countryKey)

	rtnProps := make(map[string]any, len(props))
	for _, k := range slices.Sorted(maps.Keys(props)) {
		propKey := fmt.Sprintf("%s_%s", key, k)
		rtnProps[propKey] = props[k]
		cypher += fmt.Sprintf("%s.%s = $%s, ", key, k, propKey)
	}
	cypher = cypher[:len(cypher)-2]
	cypher += fmt.Sprintf("\nWITH %s \n", key)

	return key, cypher, rtnProps
}
