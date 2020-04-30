package stvc

import (
	"errors"
	"fmt"
)

var (
	ErrUnevenVotes      = errors.New("length of rankings are uneven")
	ErrUnevenCandidates = errors.New("candidates voted on are not the same across ballots")
)

type vote struct {
	Name  string
	Value float64
}

// Count calculates the winners for the given votes in a single
// transferrable-vote system.
// The list ballots is a list of ordered strings, each of which represents the
// ranking of candidates from most to least preferred by a single voter. Each
// string is the name of a candidate. The list returned is of the names of the
// winners.
func Count(seats int, ballots [][]string) ([]string, error) {
	if len(ballots) == 0 {
		return []string{}, nil
	}

	quota := len(ballots)/(seats+1) + 1

	err := ensureVoteIntegrity(ballots)
	if err != nil {
		return nil, err
	}

	if len(ballots[0]) == 0 {
		return []string{}, nil
	}

	// Translate to vote structs
	ballotsv := make([][]vote, len(ballots))
	for i, ballot := range ballots {
		ballotsv[i] = make([]vote, len(ballot))
		for j, c := range ballot {
			var value float64 = 0
			if j == 0 {
				value = 1
			}

			ballotsv[i][j] = vote{Name: c, Value: value}
		}
	}

	overallWinners := []string{}

	for len(overallWinners) < seats {

		// Find winners
		cv := make(map[string]float64, len(ballots[0]))

		// Do while
		for {
			winners := map[string]float64{}

			for _, b := range ballotsv {
				fmt.Printf("%v\n", b)
			}

			for _, b := range ballotsv {
				// Accumulates from first choices
				first := b[0].Name
				votes, ok := cv[first]
				if !ok {
					cv[first] = 0
				}
				cv[first] = votes + b[0].Value
			}
			for c, votes := range cv {
				// Winners have reached quota
				if votes >= float64(quota) {
					winners[c] = votes
				}
			}

			fmt.Printf("winners: %v\n", winners)
			println("")

			for winner := range winners {
				overallWinners = append(overallWinners, winner)
			}
			if len(overallWinners) == seats {
				return overallWinners, nil
			}

			// Give excess winner votes to second choices
			for winner, votes := range winners {
				for i := range ballotsv {
					// If winner is first choice
					if ballotsv[i][0].Name == winner {
						frac := (votes - float64(quota)) / votes * ballotsv[i][0].Value
						// Remove first choice
						ballotsv[i] = ballotsv[i][1:]
						// Give fraction vote to second choice
						ballotsv[i][0].Value = frac
						continue
					}

					// If winner is not first choice, find and remove it from the ballot ranking
					for j := 0; j < len(ballotsv[i]); j++ {
						if ballotsv[i][j].Name == winner {
							end := []vote{}
							if j+1 < len(ballotsv[i]) {
								end = ballotsv[i][j+1:]
							}
							ballotsv[i] = append(ballotsv[i][:j], end...)
							j--
						}
					}
				}
			}

			if len(winners) == 0 {
				break
			} else {
				cv = make(map[string]float64)
			}
		}

		// Find biggest loser
		var loser string
		var minv float64 = -1
		for c, v := range cv {
			if minv == -1 {
				minv = v
			}

			if v < minv {
				minv = v
				loser = c
			}
		}

		fmt.Printf("loser: %s\n", loser)

		// Eliminate biggest loser
		for i := range ballotsv {
			// If loser is first choice, give vote to second choice
			if ballotsv[i][0].Name == loser {
				vote := ballotsv[i][0].Value
				// Remove first choice
				ballotsv[i] = ballotsv[i][1:]
				// Give vote to second choice
				ballotsv[i][0].Value = vote
				continue
			}

			// If loser is not first choice, find and remove it from the ballot ranking
			for j := 0; j < len(ballotsv[i]); j++ {
				if ballotsv[i][j].Name == loser {
					end := []vote{}
					if j+1 < len(ballotsv[i]) {
						end = ballotsv[i][j+1:]
					}
					ballotsv[i] = append(ballotsv[i][:j], end...)
					j--
				}
			}
		}

	}

	return overallWinners, nil
}

func ensureVoteIntegrity(ballots [][]string) error {
	if len(ballots) == 0 {
		return nil
	}

	ln := len(ballots[0])
	candidates := map[string]struct{}{}
	for _, r := range ballots[0] {
		candidates[r] = struct{}{}
	}

	for _, b := range ballots[1:] {
		if len(b) != ln {
			return ErrUnevenVotes
		}

		for _, r := range b {
			if _, ok := candidates[r]; !ok {
				return ErrUnevenCandidates
			}
		}
	}

	return nil
}
