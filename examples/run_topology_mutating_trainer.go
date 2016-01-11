package main

import (
	"github.com/couchbaselabs/logg"
	ng "github.com/maxxk/neurgo"
	nv "github.com/maxxk/neurvolve"
)

func RunTopologyMutatingTrainer() bool {

	ng.SeedRandom()

	// training set
	examples := ng.XnorTrainingSamples()

	// create netwwork with topology capable of solving XNOR
	cortex := ng.BasicCortex()

	// verify it can not yet solve the training set (since training would be useless in that case)
	verified := cortex.Verify(examples)
	if verified {
		panic("neural net already trained, nothing to do")
	}

	shc := &nv.StochasticHillClimber{
		FitnessThreshold:           ng.FITNESS_THRESHOLD,
		MaxIterationsBeforeRestart: 20000,
		MaxAttempts:                10,
		WeightSaturationRange:      []float64{-10000, 10000},
	}

	tmt := &nv.TopologyMutatingTrainer{
		MaxAttempts:                100,
		MaxIterationsBeforeRestart: 5,
		StochasticHillClimber:      shc,
	}
	cortexTrained, succeeded := tmt.TrainExamples(cortex, examples)
	if succeeded {
		logg.LogTo("MAIN", "Successfully trained net: %v", ng.JsonString(cortexTrained))

		// verify it can now solve the training set
		verified = cortexTrained.Verify(examples)
		if !verified {
			logg.LogTo("MAIN", "Failed to verify neural net")
			succeeded = false
		}

	}

	if !succeeded {
		logg.LogTo("MAIN", "Failed to train neural net")
	}

	return succeeded

}

func MultiRunTopologyMutatingTrainer() bool {

	logg.LogKeys["MULTI_RUN"] = true

	numSuccess := 0
	for i := 0; i < 100; i++ {
		logg.LogTo("MULTI_RUN", "Running trainer, iteration: %v", i)
		success := RunTopologyMutatingTrainer()
		if success {
			logg.LogTo("MULTI_RUN", "iteration %v succeeded", i)
			numSuccess += 1
		} else {
			logg.LogTo("MULTI_RUN", "iteration %v failed", i)
		}
	}

	logg.LogTo("MULTI_RUN", "%v/100 runs succeeded", numSuccess)

	return numSuccess == 100

}
