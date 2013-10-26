package neurvolve

import (
	"fmt"
	ng "github.com/tleyden/neurgo"
	"log"
)

type TopologyMutatingTrainer struct {
	MaxIterationsBeforeRestart int
	MaxAttempts                int
	NumOutputLayerNodes        int
	StochasticHillClimber      *StochasticHillClimber
}

func (tmt *TopologyMutatingTrainer) Train(cortex *ng.Cortex, scape Scape) (fittestCortex *ng.Cortex, succeeded bool) {

	ng.SeedRandom()

	shc := tmt.StochasticHillClimber

	includeNonTopological := false
	mutators := CortexMutatorsNonRecurrent(includeNonTopological)

	originalCortex := cortex.Copy()

	currentCortex := cortex
	currentCortex.RenderSVGFile("/Users/traun/tmp/current.svg")

	// Apply NN to problem and save fitness
	fitness := scape.Fitness(currentCortex)

	if fitness > shc.FitnessThreshold {
		succeeded = true
		return
	}

	for i := 0; ; i++ {

		log.Printf("before mutate.  i/max: %d/%d", i, tmt.MaxAttempts)

		// before we mutate the cortex, we need to init it,
		// otherwise things like Outsplice will fail because
		// there are no DataChan's.
		currentCortex.Init()

		// mutate the network
		randInt := RandomIntInRange(0, len(mutators))
		mutator := mutators[randInt]
		ok, _ := mutator(currentCortex)
		if !ok {
			log.Printf("mutate didn't work, retrying...")
			continue
		}

		isValid := currentCortex.Validate()
		if !isValid {
			log.Panicf("Cortex did not validate")
		}

		filenameJson := fmt.Sprintf("cortex-%v.json", i)
		currentCortex.MarshalJSONToFile(filenameJson)
		filenameSvg := fmt.Sprintf("cortex-%v.svg", i)
		currentCortex.RenderSVGFile(filenameSvg)
		log.Printf("after mutate. cortex written to: %v and %v", filenameSvg, filenameJson)

		log.Printf("run stochastic hill climber")

		// memetic step: call stochastic hill climber and see if it can solve it
		fittestCortex, succeeded = shc.Train(currentCortex, scape)
		log.Printf("stochastic hill climber finished.  succeeded: %v", succeeded)

		if succeeded {
			succeeded = true
			break
		}

		if i >= tmt.MaxAttempts {
			succeeded = false
			break
		}

		if ng.IntModuloProper(i, tmt.MaxIterationsBeforeRestart) {
			log.Printf("** restart.  i/max: %d/%d", i, tmt.MaxAttempts)

			currentCortex = originalCortex.Copy()
			isValid := currentCortex.Validate()
			if !isValid {
				currentCortex.Repair() // TODO: remove workaround
				isValid = currentCortex.Validate()
				if !isValid {
					log.Panicf("Cortex could not be repaired")
				}
			}

		}

	}

	return

}

func (tmt *TopologyMutatingTrainer) TrainExamples(cortex *ng.Cortex, examples []*ng.TrainingSample) (fittestCortex *ng.Cortex, succeeded bool) {

	trainingSampleScape := &TrainingSampleScape{
		examples: examples,
	}
	return tmt.Train(cortex, trainingSampleScape)

}
