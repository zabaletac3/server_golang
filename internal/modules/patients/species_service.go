package patients

import (
	"context"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type SpeciesService struct {
	repo SpeciesRepository
}

func NewSpeciesService(repo SpeciesRepository) *SpeciesService {
	return &SpeciesService{repo: repo}
}

// Resolve finds an existing species or creates a new one.
// Returns:
//   - (response, nil, nil): exact match or auto-merged (similarity > 0.85)
//   - (nil, conflict, ErrSpeciesConflict): ambiguous match (0.60 < sim <= 0.85)
//   - (response, nil, nil): newly created species (no close match)
func (s *SpeciesService) Resolve(ctx context.Context, tenantID primitive.ObjectID, name string) (*SpeciesResponse, *SpeciesConflictResponse, error) {
	norm := normalizeSpeciesName(name)
	if norm == "" {
		return nil, nil, ErrSpeciesNotFound
	}

	// 1. Exact match on normalized name
	existing, err := s.repo.FindByNormalizedName(ctx, tenantID, norm)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		resp := toSpeciesResponse(existing)
		return &resp, nil, nil
	}

	// 2. Load all tenant species for fuzzy comparison
	allSpecies, err := s.repo.FindAllByTenant(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}

	inputTrigrams := buildTrigrams(norm)
	var bestMatch *Species
	bestSim := 0.0
	var suggestions []Species

	for i := range allSpecies {
		sp := &allSpecies[i]
		sim := jaccardSimilarity(inputTrigrams, buildTrigrams(sp.NormalizedName))
		if sim > bestSim {
			bestSim = sim
			bestMatch = sp
		}
		if sim > 0.60 {
			suggestions = append(suggestions, *sp)
		}
	}

	// 3. Auto-merge: high similarity
	if bestSim > 0.85 && bestMatch != nil {
		resp := toSpeciesResponse(bestMatch)
		return &resp, nil, nil
	}

	// 4. Ambiguous: return suggestions
	if bestSim > 0.60 && len(suggestions) > 0 {
		sugResponses := make([]SpeciesResponse, len(suggestions))
		for i, sp := range suggestions {
			sugResponses[i] = toSpeciesResponse(&sp)
		}
		return nil, &SpeciesConflictResponse{
			Message:     "similar species found",
			Suggestions: sugResponses,
		}, ErrSpeciesConflict
	}

	// 5. No close match: create new species
	now := time.Now()
	newSpecies := &Species{
		ID:             primitive.NewObjectID(),
		TenantID:       tenantID,
		Name:           strings.TrimSpace(name),
		NormalizedName: norm,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, newSpecies); err != nil {
		return nil, nil, err
	}

	resp := toSpeciesResponse(newSpecies)
	return &resp, nil, nil
}

func (s *SpeciesService) ListByTenant(ctx context.Context, tenantID primitive.ObjectID) ([]SpeciesResponse, error) {
	species, err := s.repo.FindAllByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	responses := make([]SpeciesResponse, len(species))
	for i, sp := range species {
		responses[i] = toSpeciesResponse(&sp)
	}
	return responses, nil
}

// --- Trigram Jaccard Algorithm ---

// normalizeSpeciesName lowercases, trims, and removes diacritics.
func normalizeSpeciesName(input string) string {
	s := strings.TrimSpace(input)
	s = strings.ToLower(s)
	s = removeDiacritics(s)
	return s
}

// removeDiacritics decomposes unicode into NFD form and strips combining marks.
func removeDiacritics(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return result
}

// buildTrigrams generates a set of 3-character sliding windows with edge padding.
func buildTrigrams(s string) map[string]struct{} {
	padded := []rune("  " + s + "  ")
	trigrams := make(map[string]struct{})
	for i := 0; i <= len(padded)-3; i++ {
		trigrams[string(padded[i:i+3])] = struct{}{}
	}
	return trigrams
}

// jaccardSimilarity computes |A ∩ B| / |A ∪ B| for two trigram sets.
func jaccardSimilarity(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	if union == 0 {
		return 1.0
	}

	return float64(intersection) / float64(union)
}
