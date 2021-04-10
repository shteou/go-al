package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/MaxHalford/eaopt"
	log "github.com/sirupsen/logrus"
)

var (
	corpus       = strings.Split("ABCD", "")
	genomeLength uint
)

type Genome []string

type GenomeState struct {
	Children  uint
	Energy    float64
	FoundFood bool
}

func logWithFields(g *GenomeState) *log.Entry {
	return log.WithFields(log.Fields{
		"state": g,
	})
}

func (G Genome) Evaluate() (fitness float64, err error) {
	index := 0

	g := GenomeState{Children: 0, Energy: 10.0, FoundFood: false}

	for i := 0; i < 1000; i++ {
		gene := G[index]
		switch gene[0] {
		case 'A': // No Op
			logWithFields(&g).Debug("No Op")
		case 'B': // Spawn Child
			if g.Energy > 5.0 {
				g.Children += 1
				logWithFields(&g).Debug("Spawn Child succeeded")
			} else {
				logWithFields(&g).Debug("Spawn Child failed")
			}

			g.Energy -= 5.0
		case 'C': // Locate Food
			g.FoundFood = true
			logWithFields(&g).Debug("Located Food")
		case 'D': // Eat Food
			if g.FoundFood {
				g.Energy += 10.0
			}
		default:
			logWithFields(&g).Debug("Unexpected")
		}

		if gene != "C" {
			g.FoundFood = false
		}

		g.Energy = g.Energy - 1.0

		// Wasted
		if g.Energy <= 0.0 {
			logWithFields(&g).Debugf("Died on iteration %d", i)
			break
		}

		// Iterate through the genome, looping at the end
		index = (index + 1) % len(G)
	}

	return 1.0 - (float64(g.Children) / 1000.0), nil
}

func (G Genome) Mutate(rng *rand.Rand) {
	eaopt.MutUniformString(G, corpus, 1, rng)
}

func (G Genome) Crossover(Y eaopt.Genome, rng *rand.Rand) {
	eaopt.CrossGNXString(G, Y.(Genome), 2, rng)
}

func MakeStrings(rng *rand.Rand) eaopt.Genome {
	return Genome(eaopt.InitUnifString(genomeLength, corpus, rng))
}

func (G Genome) Clone() eaopt.Genome {
	var XX = make(Genome, len(G))
	copy(XX, G)
	return XX
}

func evolveGenomes(len uint64) {
	genomeLength = uint(len)

	var ga, err = eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		fmt.Println(err)
		return
	}

	ga.NGenerations = 20
	ga.NPops = 10
	ga.MigFrequency = 5
	ga.Migrator = eaopt.MigRing{NMigrants: 5}
	ga.ParallelEval = false

	winner := ""
	mutex := sync.Mutex{}

	// Periodically update progress, or when a new champion is found
	ga.Callback = func(ga *eaopt.GA) {
		mutex.Lock()
		defer mutex.Unlock()

		if ga.Generations%5 == 0 {
			fmt.Printf("%d)\n", ga.Generations)
		}

		var buffer bytes.Buffer
		for _, letter := range ga.HallOfFame[0].Genome.(Genome) {
			buffer.WriteString(letter)
		}

		if winner != buffer.String() {
			winner = buffer.String()
			fmt.Printf("%d) Result -> %s (%d children)\n", ga.Generations, buffer.String(), uint((1.0-ga.HallOfFame[0].Fitness)*1000.0))
		}
	}

	// Run the GA
	ga.Minimize(MakeStrings)
}

func parseGenomeString(genome string) {
	codons := strings.Split(genome, "")
	for i := 0; i < len(codons); i++ {
		switch codons[i] {
		case "A":
			fmt.Println("Nop")
		case "B":
			fmt.Println("Spawn Child")
		case "C":
			fmt.Println("Locate Food")
		case "D":
			fmt.Println("Eat Food")
		}
	}
}

func evaluateSingleGenome(genomeString string) {
	g := Genome(strings.Split(genomeString, ""))
	fitness, _ := g.Evaluate()
	fmt.Println(fitness)
}

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("\tga parse <genome> - Parses the genome into a textual representation\n")
	fmt.Printf("\tga evolve <genomeLength> - Evolves genomes with the given fixed length\n")
	fmt.Printf("\tga <genome> - Evaluates the fitness of the given genome\n")
}

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		usage()
		os.Exit(1)
	} else if args[0] == "evolve" {
		len := args[1]
		uintLen, _ := strconv.ParseUint(len, 10, 32)
		evolveGenomes(uintLen)
	} else if args[0] == "parse" {
		parseGenomeString(args[1])
	} else {
		log.SetLevel(log.DebugLevel)
		evaluateSingleGenome(args[0])
	}
}
