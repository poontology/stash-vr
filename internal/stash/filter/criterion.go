package filter

import (
	"stash-vr/internal/stash/gql"
	"strconv"
)

func decodeSimple[T string | bool](c interface{}, dst *T) error {
	m := c.(map[string]interface{})
	x := m["value"].(string)
	switch any(*dst).(type) {
	case string:
		d := any(dst).(*string)
		*d = x
	case bool:
		b, _ := strconv.ParseBool(x)
		d := any(dst).(*bool)
		*d = b
	}
	return nil
}

func modifier(c map[string]any) gql.CriterionModifier {
	return gql.CriterionModifier(c["modifier"].(string))
}

func parseIntCriterionInput(c map[string]any) *gql.IntCriterionInput {
	out := gql.IntCriterionInput{
		Modifier: modifier(c),
	}

	v1, _ := get[float64](c, "value.value")
	out.Value = int(v1)

	v2, err := get[float64](c, "value.value2")
	if err == nil {
		out.Value2 = int(v2)
	}

	return &out
}

func parseHierarchicalMultiCriterionInput(c map[string]any) *gql.HierarchicalMultiCriterionInput {
	out := gql.HierarchicalMultiCriterionInput{
		Modifier: modifier(c),
	}

	depth, _ := get[float64](c, "value.depth")
	out.Depth = int(depth)

	items, _ := get[[]any](c, "value.items")
	out.Value = make([]string, len(items))

	for i := range items {
		id, _ := get[float64](items[i].(map[string]any), "id")
		out.Value[i] = strconv.Itoa(int(id))
	}

	excluded, _ := get[[]any](c, "value.excluded")
	out.Excludes = make([]string, len(excluded))
	for i := range excluded {
		id, _ := get[float64](excluded[i].(map[string]any), "id")
		out.Excludes[i] = strconv.Itoa(int(id))
	}
	return &out
}

func parseMultiCriterionInput(c map[string]any) *gql.MultiCriterionInput {
	out := gql.MultiCriterionInput{
		Modifier: modifier(c),
	}

	items, err := get[[]any](c, "value.items")
	if err == nil {
		out.Value = make([]string, len(items))
		for i := range items {
			id, _ := get[float64](items[i].(map[string]any), "id")
			out.Value[i] = strconv.Itoa(int(id))
		}

		excluded, _ := get[[]any](c, "value.excluded")
		out.Excludes = make([]string, len(excluded))
		for i := range excluded {
			id, _ := get[float64](items[i].(map[string]any), "id")
			out.Excludes[i] = strconv.Itoa(int(id))
		}
		return &out
	}

	arr, err := get[[]any](c, "value")
	if err == nil {
		out.Value = make([]string, len(arr))
		for i := range arr {
			id, _ := get[float64](arr[i].(map[string]any), "id")
			out.Value[i] = strconv.Itoa(int(id))
		}
		return &out
	}

	return &out
}

func parseTimestampCriterionInput(c map[string]any) *gql.TimestampCriterionInput {
	out := gql.TimestampCriterionInput{
		Modifier: modifier(c),
	}

	out.Value, _ = get[string](c, "value.value")
	value2, err := get[string](c, "value.value2")
	if err == nil {
		out.Value2 = value2
	}

	return &out
}

func parseDateCriterionInput(c map[string]any) *gql.DateCriterionInput {
	out := gql.DateCriterionInput{
		Modifier: modifier(c),
	}

	out.Value, _ = get[string](c, "value.value")
	value2, err := get[string](c, "value.value2")
	if err == nil {
		out.Value2 = value2
	}

	return &out
}

func parsePhashDistanceCriterionInput(c map[string]any) *gql.PhashDistanceCriterionInput {
	out := gql.PhashDistanceCriterionInput{
		Modifier: modifier(c),
	}
	distance, _ := get[float64](c, "value.distance")
	out.Distance = int(distance)
	out.Value, _ = get[string](c, "value.value")
	return &out
}

func parseResolutionCriterionInput(c map[string]any) *gql.ResolutionCriterionInput {
	out := gql.ResolutionCriterionInput{
		Modifier: modifier(c),
	}

	value, _ := get[string](c, "value")
	switch value {
	case "144p":
		out.Value = gql.ResolutionEnumVeryLow
	case "240p":
		out.Value = gql.ResolutionEnumLow
	case "360p":
		out.Value = gql.ResolutionEnumR360p
	case "480p":
		out.Value = gql.ResolutionEnumStandard
	case "540p":
		out.Value = gql.ResolutionEnumWebHd
	case "720p":
		out.Value = gql.ResolutionEnumStandardHd
	case "1080p":
		out.Value = gql.ResolutionEnumFullHd
	case "1440p":
		out.Value = gql.ResolutionEnumQuadHd
	case "1920p":
		out.Value = gql.ResolutionEnumVrHd
	case "4k":
		out.Value = gql.ResolutionEnumFourK
	case "5k":
		out.Value = gql.ResolutionEnumFiveK
	case "6k":
		out.Value = gql.ResolutionEnumSixK
	case "8k":
		out.Value = gql.ResolutionEnumEightK
	}

	return &out
}

func parseStashIDCriterionInput(c map[string]any) *gql.StashIDCriterionInput {
	out := gql.StashIDCriterionInput{
		Modifier: modifier(c),
	}
	out.Endpoint, _ = get[string](c, "value.endpoint")
	out.Stash_id, _ = get[string](c, "value.stashID")
	return &out
}

func parsePHashDuplicationCriterionInput(c map[string]any) *gql.PHashDuplicationCriterionInput {
	out := gql.PHashDuplicationCriterionInput{}

	duplicated, _ := get[string](c, "value")
	out.Duplicated, _ = strconv.ParseBool(duplicated)

	return &out
}

func parseStringCriterionInput(c map[string]any) *gql.StringCriterionInput {
	out := gql.StringCriterionInput{
		Modifier: modifier(c),
	}
	out.Value, _ = get[string](c, "value")
	return &out
}
