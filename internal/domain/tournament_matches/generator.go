package matches

// Generator builds the matches of a stage for a given format and seed order.
// New formats (double elimination, GSL, round robin...) implement this
// interface and register themselves in generators.
type Generator interface {
	// Generate returns the matches to persist for the stage's first seeding.
	Generate(stage StageInfo, seeds []int64) ([]Match, error)
}

var generators = map[string]Generator{
	rulesFormatSingleElimination: SingleEliminationGenerator{},
	rulesFormatDoubleElimination: DoubleEliminationGenerator{},
	rulesFormatSwiss:             SwissGenerator{},
}

// generatorFor resolves the generator for a stage format code.
func generatorFor(formatCode string) (Generator, bool) {
	g, ok := generators[formatCode]
	return g, ok
}

// Format codes mirrored from tournament_stages rules to keep the registry
// self-contained.
const (
	rulesFormatSingleElimination = "SINGLE_ELIMINATION"
	rulesFormatDoubleElimination = "DOUBLE_ELIMINATION"
	rulesFormatSwiss             = "SWISS"
	rulesFormatGroupStage        = "GROUP_STAGE"
)
