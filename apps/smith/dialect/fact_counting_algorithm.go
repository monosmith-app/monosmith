package dialect

import "sort"

// MinAgreeing is the fewest sibling projects that must share a value for it to
// be followed: two projects are a coincidence, three are a habit.
const MinAgreeing = 3

// Count is how many siblings have one value of a fact.
type Count struct {
	Value string
	N     int
}

// Fact is what the siblings say about one shape key.
type Fact struct {
	Key string
	// Total is the number of siblings that report the key.
	Total int
	// Counts are the values, the most common first, ties by value.
	Counts []Count
}

// Dominant is the value to follow: shared by at least MinAgreeing siblings
// and by at least two thirds of those that report the key. ok is false when
// no value is.
func (f Fact) Dominant() (c Count, ok bool) {
	if len(f.Counts) == 0 {
		return Count{}, false
	}
	top := f.Counts[0]
	return top, top.N >= MinAgreeing && top.N*3 >= f.Total*2
}

// Split reports that there are enough siblings to have a habit and they do
// not share one: the case that is asked, never guessed.
func (f Fact) Split() bool {
	_, dominant := f.Dominant()
	return f.Total >= MinAgreeing && !dominant
}

// CountFacts counts, for each key the shapes report, how many shapes have each
// value. The facts are sorted by key.
func CountFacts(shapes []map[string]string) []Fact {
	counts := map[string]map[string]int{}
	for _, shape := range shapes {
		for key, value := range shape {
			if counts[key] == nil {
				counts[key] = map[string]int{}
			}
			counts[key][value]++
		}
	}
	facts := make([]Fact, 0, len(counts))
	for key, values := range counts {
		f := Fact{Key: key}
		for value, n := range values {
			f.Counts = append(f.Counts, Count{Value: value, N: n})
			f.Total += n
		}
		sort.Slice(f.Counts, func(i, j int) bool {
			if f.Counts[i].N != f.Counts[j].N {
				return f.Counts[i].N > f.Counts[j].N
			}
			return f.Counts[i].Value < f.Counts[j].Value
		})
		facts = append(facts, f)
	}
	sort.Slice(facts, func(i, j int) bool { return facts[i].Key < facts[j].Key })
	return facts
}
