package main

import (
	"github.com/couchbaselabs/logg"
	ng "github.com/maxxk/neurgo"
	nv "github.com/maxxk/neurvolve"
	"math"
)

func RunStochasticHillClimber() {

	ng.SeedRandom()

	// training set -- todo: examples := ng.XnorTrainingSamples()
	examples := ng.XnorTrainingSamples()

	// create netwwork with topology capable of solving XNOR
	cortex := ng.XnorCortexUntrained()

	// verify it can not yet solve the training set (since training would be useless in that case)
	verified := cortex.Verify(examples)
	if verified {
		panic("neural net already trained, nothing to do")
	}

	shc := &nv.StochasticHillClimber{
		FitnessThreshold:           ng.FITNESS_THRESHOLD,
		MaxIterationsBeforeRestart: 2000,
		MaxAttempts:                2000,
		WeightSaturationRange:      []float64{-100 * math.Pi, 100 * math.Pi},
	}
	cortexTrained, succeeded := shc.TrainExamples(cortex, examples)
	if !succeeded {
		panic("could not train neural net")
	}

	// verify it can now solve the training set
	verified = cortexTrained.Verify(examples)
	if !verified {
		panic("could not verify neural net")
	}

	logg.LogTo("DEBUG", "trained cortex: %v", cortexTrained)

	logg.Log("done")

}
