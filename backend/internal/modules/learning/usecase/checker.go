package usecase

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
)

type checkResult struct {
	isCorrect bool
	mistakes  []string
}

func checkAnswer(challenge content_domain.Challenge, userAnswer json.RawMessage) checkResult {
	if challenge.Type == content_domain.ChallengeTypeTheory {
		return checkResult{isCorrect: true}
	}
	if len(userAnswer) == 0 {
		return incorrect("answer is empty")
	}

	expected, err := decodeAny(challenge.Answers)
	if err != nil {
		return incorrect("challenge answers are malformed")
	}
	submitted, err := decodeAny(userAnswer)
	if err != nil {
		return incorrect("answer is malformed")
	}

	switch challenge.Type {
	case content_domain.ChallengeTypeSingleChoice, content_domain.ChallengeTypeImage, content_domain.ChallengeTypeQuote, content_domain.ChallengeTypeTrueFalse:
		return checkSingle(expected, submitted)
	case content_domain.ChallengeTypeMultiple:
		return checkStringSet(expected, submitted)
	case content_domain.ChallengeTypeTimeline:
		return checkStringOrder(expected, submitted)
	case content_domain.ChallengeTypeMatchPairs, content_domain.ChallengeTypeMatchImage:
		return checkPairs(expected, submitted)
	case content_domain.ChallengeTypeMatchPhotos:
		return checkPhotoPairs(expected, submitted)
	case content_domain.ChallengeTypeFillBlank:
		return checkFillBlank(expected, submitted)
	case content_domain.ChallengeTypeMapPoint:
		return checkMapPoint(expected, submitted)
	case content_domain.ChallengeTypeMapArea:
		return checkMapArea(expected, submitted)
	default:
		return incorrect("challenge type is unsupported")
	}
}

func checkSingle(expected any, submitted any) checkResult {
	expectedItems := scalarList(expected)
	if len(expectedItems) != 1 {
		return incorrect("challenge expects exactly one answer")
	}
	if normalizeScalar(submitted) != expectedItems[0] {
		return incorrect("answer does not match")
	}
	return checkResult{isCorrect: true}
}

func checkStringSet(expected any, submitted any) checkResult {
	expectedItems := scalarList(expected)
	submittedItems := scalarList(submitted)
	sort.Strings(expectedItems)
	sort.Strings(submittedItems)
	if !sameStrings(expectedItems, submittedItems) {
		return incorrect("selected options do not match")
	}
	return checkResult{isCorrect: true}
}

func checkStringOrder(expected any, submitted any) checkResult {
	if !sameStrings(scalarList(expected), scalarList(submitted)) {
		return incorrect("order does not match")
	}
	return checkResult{isCorrect: true}
}

func checkPairs(expected any, submitted any) checkResult {
	expectedPairs := pairList(expected)
	submittedPairs := pairList(submitted)
	sort.Strings(expectedPairs)
	sort.Strings(submittedPairs)
	if len(expectedPairs) == 0 || !sameStrings(expectedPairs, submittedPairs) {
		return incorrect("pairs do not match")
	}
	return checkResult{isCorrect: true}
}

func checkPhotoPairs(expected any, submitted any) checkResult {
	expectedPairs := photoPairList(expected)
	submittedPairs := photoPairList(submitted)
	sort.Strings(expectedPairs)
	sort.Strings(submittedPairs)
	if len(expectedPairs) == 0 || !sameStrings(expectedPairs, submittedPairs) {
		return incorrect("photo pairs do not match")
	}
	return checkResult{isCorrect: true}
}

func checkFillBlank(expected any, submitted any) checkResult {
	answer := normalizeText(normalizeScalar(submitted))
	if answer == "" {
		return incorrect("answer is empty")
	}
	for _, item := range scalarList(expected) {
		if normalizeText(item) == answer {
			return checkResult{isCorrect: true}
		}
	}
	return incorrect("answer does not match")
}

func checkMapPoint(expected any, submitted any) checkResult {
	target, ok := objectValue(expected)
	if !ok {
		return incorrect("challenge map target is malformed")
	}
	answer, ok := objectValue(submitted)
	if !ok {
		return incorrect("map answer is malformed")
	}
	targetLat, ok := numberValue(target, "lat")
	if !ok {
		return incorrect("challenge map target is malformed")
	}
	targetLng, ok := numberValue(target, "lng")
	if !ok {
		return incorrect("challenge map target is malformed")
	}
	radius, ok := numberValue(target, "radius_m")
	if !ok || radius <= 0 {
		return incorrect("challenge map target radius is malformed")
	}
	answerLat, ok := numberValue(answer, "lat")
	if !ok {
		return incorrect("map answer is malformed")
	}
	answerLng, ok := numberValue(answer, "lng")
	if !ok {
		return incorrect("map answer is malformed")
	}
	if haversineMeters(targetLat, targetLng, answerLat, answerLng) > radius {
		return incorrect("map point is outside target radius")
	}
	return checkResult{isCorrect: true}
}

func checkMapArea(expected any, submitted any) checkResult {
	target, ok := objectValue(expected)
	if !ok {
		return incorrect("challenge map area is malformed")
	}
	center, ok := nestedObjectValue(target, "center")
	if !ok {
		return incorrect("challenge map area center is malformed")
	}
	centerLat, ok := numberValue(center, "lat")
	if !ok {
		return incorrect("challenge map area center is malformed")
	}
	centerLng, ok := numberValue(center, "lng")
	if !ok {
		return incorrect("challenge map area center is malformed")
	}
	expectedArea, ok := numberValue(target, "area_m2")
	if !ok || expectedArea <= 0 {
		return incorrect("challenge map area size is malformed")
	}
	centerRadius, ok := numberValue(target, "center_radius_m")
	if !ok || centerRadius <= 0 {
		return incorrect("challenge map area center radius is malformed")
	}
	tolerance, ok := numberValue(target, "area_tolerance")
	if !ok || tolerance < 0 {
		return incorrect("challenge map area tolerance is malformed")
	}

	answer, ok := objectValue(submitted)
	if !ok {
		return incorrect("map area answer is malformed")
	}
	answerCenter, ok := nestedObjectValue(answer, "center")
	var answerCenterLat, answerCenterLng, answerArea float64
	if ok {
		answerCenterLat, ok = numberValue(answerCenter, "lat")
		if !ok {
			return incorrect("map area answer center is malformed")
		}
		answerCenterLng, ok = numberValue(answerCenter, "lng")
		if !ok {
			return incorrect("map area answer center is malformed")
		}
		answerArea, ok = numberValue(answer, "area_m2")
		if !ok || answerArea <= 0 {
			return incorrect("map area answer size is malformed")
		}
	} else {
		points := latLngList(answer["points"])
		if len(points) < 3 {
			return incorrect("map area answer is malformed")
		}
		answerCenterLat, answerCenterLng, answerArea = polygonCentroidAndArea(points)
	}
	if haversineMeters(centerLat, centerLng, answerCenterLat, answerCenterLng) > centerRadius {
		return incorrect("map area center is outside target radius")
	}
	minArea := expectedArea * math.Max(0, 1-tolerance)
	maxArea := expectedArea * (1 + tolerance)
	if answerArea < minArea || answerArea > maxArea {
		return incorrect("map area size is outside tolerance")
	}
	return checkResult{isCorrect: true}
}

func decodeAny(raw json.RawMessage) (any, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func scalarList(value any) []string {
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if normalized := normalizeScalar(item); normalized != "" {
				out = append(out, normalized)
			}
		}
		return out
	default:
		if normalized := normalizeScalar(v); normalized != "" {
			return []string{normalized}
		}
		return nil
	}
}

func normalizeScalar(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func pairList(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		left := normalizeScalar(m["left_id"])
		right := normalizeScalar(m["right_id"])
		if left != "" && right != "" {
			out = append(out, left+"="+right)
		}
	}
	return out
}

func photoPairList(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		photo := normalizeScalar(m["photo_id"])
		label := normalizeScalar(m["label_id"])
		if photo != "" && label != "" {
			out = append(out, photo+"="+label)
		}
	}
	return out
}

type latLng struct {
	lat float64
	lng float64
}

func objectValue(value any) (map[string]any, bool) {
	if m, ok := value.(map[string]any); ok {
		return m, true
	}
	items, ok := value.([]any)
	if !ok || len(items) != 1 {
		return nil, false
	}
	m, ok := items[0].(map[string]any)
	return m, ok
}

func nestedObjectValue(value map[string]any, key string) (map[string]any, bool) {
	m, ok := value[key].(map[string]any)
	return m, ok
}

func numberValue(value map[string]any, key string) (float64, bool) {
	switch v := value[key].(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	default:
		return 0, false
	}
}

func latLngList(value any) []latLng {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]latLng, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		lat, ok := numberValue(m, "lat")
		if !ok {
			continue
		}
		lng, ok := numberValue(m, "lng")
		if !ok {
			continue
		}
		out = append(out, latLng{lat: lat, lng: lng})
	}
	return out
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusMeters = 6371000.0
	lat1Rad := degToRad(lat1)
	lat2Rad := degToRad(lat2)
	dLat := degToRad(lat2 - lat1)
	dLng := degToRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func polygonCentroidAndArea(points []latLng) (float64, float64, float64) {
	originLat, originLng := averageLatLng(points)
	coords := make([]struct{ x, y float64 }, 0, len(points))
	for _, point := range points {
		coords = append(coords, struct{ x, y float64 }{
			x: degToRad(point.lng-originLng) * 6371000 * math.Cos(degToRad(originLat)),
			y: degToRad(point.lat-originLat) * 6371000,
		})
	}

	var twiceArea, centroidX, centroidY float64
	for i := range coords {
		j := (i + 1) % len(coords)
		cross := coords[i].x*coords[j].y - coords[j].x*coords[i].y
		twiceArea += cross
		centroidX += (coords[i].x + coords[j].x) * cross
		centroidY += (coords[i].y + coords[j].y) * cross
	}
	if math.Abs(twiceArea) < 1e-9 {
		return originLat, originLng, 0
	}
	centroidX /= 3 * twiceArea
	centroidY /= 3 * twiceArea
	lat := originLat + radToDeg(centroidY/6371000)
	lng := originLng + radToDeg(centroidX/(6371000*math.Cos(degToRad(originLat))))
	return lat, lng, math.Abs(twiceArea) / 2
}

func averageLatLng(points []latLng) (float64, float64) {
	var latSum, lngSum float64
	for _, point := range points {
		latSum += point.lat
		lngSum += point.lng
	}
	return latSum / float64(len(points)), lngSum / float64(len(points))
}

func degToRad(value float64) float64 {
	return value * math.Pi / 180
}

func radToDeg(value float64) float64 {
	return value * 180 / math.Pi
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func normalizeText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func incorrect(reason string) checkResult {
	return checkResult{mistakes: []string{reason}}
}
