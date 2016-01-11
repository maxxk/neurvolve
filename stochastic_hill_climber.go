package neurvolve

import (
	"github.com/couchbaselabs/logg"
	ng "github.com/maxxk/neurgo"
	"math"
	"math/rand"
)

type StochasticHillClimber struct {
	FitnessThreshold           float64
	MaxIterationsBeforeRestart int
	MaxAttempts                int
	WeightSaturationRange      []float64
}

func (shc *StochasticHillClimber) Train(cortex *ng.Cortex, scape Scape) (resultNeuralNet *ng.Cortex, fitness float64, succeeded bool) {

	shc.validate()

	numAttempts := 0

	fittestNeuralNet := cortex
	resultNeuralNet = cortex

	// Apply NN to problem and save fitness
	fitness = scape.Fitness(fittestNeuralNet)
	logg.LogTo("MAIN", "Initial fitness: %v", fitness)

	if fitness > shc.FitnessThreshold {
		succeeded = true
		return
	}

	for i := 0; ; i++ {

		// Save the genotype
		candidateNeuralNet := fittestNeuralNet.Copy()

		// Perturb synaptic weights and biases
		PerturbParameters(candidateNeuralNet, shc.WeightSaturationRange)

		// Re-Apply NN to problem
		candidateFitness := scape.Fitness(candidateNeuralNet)
		logg.LogTo("DEBUG", "candidate fitness: %v", fitness)

		// If fitness of perturbed NN is higher, discard original NN and keep new
		// If fitness of original is higher, discard perturbed and keep old.

		if candidateFitness > fitness {
			logg.LogTo("MAIN", "i: %v candidateFitness: %v > fitness: %v", i, candidateFitness, fitness)
			i = 0
			fittestNeuralNet = candidateNeuralNet
			resultNeuralNet = candidateNeuralNet
			fitness = candidateFitness

		}

		if candidateFitness > shc.FitnessThreshold {
			logg.LogTo("MAIN", "candidateFitness: %v > Threshold.  Success at i=%v", candidateFitness, i)
			succeeded = true
			break
		}

		if ng.IntModuloProper(i, shc.MaxIterationsBeforeRestart) {
			logg.LogTo("MAIN", "** restart hill climber.  fitness: %f i/max: %d/%d", fitness, numAttempts, shc.MaxAttempts)
			numAttempts += 1
			i = 0
			shc.resetParametersToRandom(fittestNeuralNet)
			ng.SeedRandom()
		}

		if numAttempts >= shc.MaxAttempts {
			succeeded = false
			break
		}

	}

	return

}

func (shc *StochasticHillClimber) TrainExamples(cortex *ng.Cortex, examples []*ng.TrainingSample) (fittestNeuralNet *ng.Cortex, fitness float64, succeeded bool) {

	trainingSampleScape := &TrainingSampleScape{
		examples: examples,
	}
	return shc.Train(cortex, trainingSampleScape)

}

// 1. Each neuron in the neural net (weight or bias) will be chosen for perturbation
//    with a probability of 1/sqrt(nn_size)
// 2. Within the chosen neuron, the weights which will be perturbed will be chosen
//    with probability of 1/sqrt(parameters_size)
// 3. The intensity of the parameter perturbation will chosen with uniform distribution
//    of -pi and pi
func PerturbParameters(cortex *ng.Cortex, saturationBounds []float64) {

	// pick the neurons to perturb (at least one)
	neurons := chooseNeuronsToPerturb(cortex)

	for _, neuron := range neurons {
		logg.LogTo("DEBUG", "Going to perturb neuron: %v", neuron.NodeId.UUID)
		perturbNeuron(neuron, saturationBounds)
	}

}

func (shc *StochasticHillClimber) resetParametersToRandom(cortex *ng.Cortex) {

	neurons := cortex.Neurons
	for _, neuronNode := range neurons {
		for _, cxn := range neuronNode.Inbound {
			cxn.Weights = ng.RandomWeights(len(cxn.Weights))
		}
		neuronNode.Bias = ng.RandomBias()
	}

}

func chooseNeuronsToPerturb(cortex *ng.Cortex) []*ng.Neuron {

	neuronsToPerturb := make([]*ng.Neuron, 0)

	// choose some random neurons to perturb.  we need at least one, so
	// keep looping until we've chosen at least one
	didChooseNeuron := false
	for {

		probability := nodePerturbProbability(cortex)
		neurons := cortex.Neurons
		for _, neuronNode := range neurons {
			if rand.Float64() < probability {
				neuronsToPerturb = append(neuronsToPerturb, neuronNode)
				didChooseNeuron = true
			}
		}

		if didChooseNeuron {
			break
		}

	}
	return neuronsToPerturb

}

func nodePerturbProbability(cortex *ng.Cortex) float64 {
	neurons := cortex.Neurons
	numNeurons := len(neurons)
	return 1 / math.Sqrt(float64(numNeurons))
}

func perturbNeuron(neuron *ng.Neuron, saturationBounds []float64) {

	probability := parameterPerturbProbability(neuron)

	// keep trying until we've perturbed at least one parameter
	for {
		didPerturbWeight := false
		for _, cxn := range neuron.Inbound {
			didPerturbWeight = possiblyPerturbConnection(cxn, probability, saturationBounds)
		}

		didPerturbBias := possiblyPerturbBias(neuron, probability, saturationBounds)

		// did we perturb anything?  if so, we're done
		if didPerturbWeight || didPerturbBias {
			break
		}

	}

}

func parameterPerturbProbability(neuron *ng.Neuron) float64 {
	numWeights := 0
	for _, connection := range neuron.Inbound {
		numWeights += len(connection.Weights)
	}
	return 1 / math.Sqrt(float64(numWeights))
}

func possiblyPerturbConnection(cxn *ng.InboundConnection, probability float64, saturationBounds []float64) bool {

	didPerturb := false
	for j, weight := range cxn.Weights {
		if rand.Float64() < probability {
			perturbedWeight := perturbParameter(weight, saturationBounds)
			logg.LogTo("DEBUG", "weight %v -> %v", weight, perturbedWeight)
			cxn.Weights[j] = perturbedWeight
			didPerturb = true
		}
	}
	return didPerturb

}

func possiblyPerturbBias(neuron *ng.Neuron, probability float64, saturationBounds []float64) bool {
	didPerturb := false
	if rand.Float64() < probability {
		bias := neuron.Bias
		perturbedBias := perturbParameter(bias, saturationBounds)
		neuron.Bias = perturbedBias
		logg.LogTo("DEBUG", "bias %v -> %v", bias, perturbedBias)
		didPerturb = true
	}
	return didPerturb
}

func (shc *StochasticHillClimber) validate() {
	if len(shc.WeightSaturationRange) == 0 {
		logg.LogPanic("Invalid (empty) WeightSaturationRange")
	}
}
