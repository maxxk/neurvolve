package neurvolve

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/go.assert"
	ng "github.com/tleyden/neurgo"
	"log"
	"testing"
)

func verifyWeightsModified(neuron, neuronCopy *ng.Neuron) bool {
	foundModifiedWeight := false

	// make sure the weights have been modified for at least
	// one of the inbound connections
	originalInboundMap := neuron.InboundUUIDMap()
	copyInboundMap := neuronCopy.InboundUUIDMap()

	for uuid, connection := range originalInboundMap {
		connectionCopy := copyInboundMap[uuid]
		for i, weight := range connection.Weights {
			weightCopy := connectionCopy.Weights[i]
			if weight != weightCopy {
				foundModifiedWeight = true
				break
			}
		}
	}
	return foundModifiedWeight

}

func TestNeuronAddInlinkRecurrent(t *testing.T) {

	madeNonRecurrentInlink := false
	madeRecurrentInlink := false

	for i := 0; i < 100; i++ {
		xnorCortex := ng.XnorCortex()
		neuron := xnorCortex.NeuronUUIDMap()["output-neuron"]
		inboundConnection := NeuronAddInlinkRecurrent(neuron)
		if neuron.IsInboundConnectionRecurrent(inboundConnection) {

			log.Printf("added inboundConnection: %v", inboundConnection)

			// the first time we make a nonRecurrentInlink,
			// test the network out
			if madeRecurrentInlink == false {

				// workaround for bug:
				// 1. neuron is Init() when only has X inbound cxn
				// 2. new (recursive) inbound cxn added (X+1)
				// 3. dataChan only has X buffer
				// 4. dataChan needs X+1 buffer
				// 5. code blocks when trying to send recurrent 0val
				// real fix:
				// 1. when adding a new inbound, recreate
				//    dataChan
				// 2. go fix up any outbounds that point to datachan
				// 3. either that or re-create entire network
				jsonString := fmt.Sprintf("%v", xnorCortex.String())
				jsonBytes := []byte(jsonString)
				cortex := &ng.Cortex{}
				err := json.Unmarshal(jsonBytes, cortex)
				if err != nil {
					log.Fatal(err)
				}
				assert.True(t, err == nil)

				// make sure the network actually works
				examples := ng.XnorTrainingSamples()
				fitness := cortex.Fitness(examples)
				assert.True(t, fitness >= 0)

			}

			madeRecurrentInlink = true
		} else {

			// the first time we make a nonRecurrentInlink,
			// test the network out
			if madeNonRecurrentInlink == false {
				// make sure the network doesn't totally break
				examples := ng.XnorTrainingSamples()
				fitness := xnorCortex.Fitness(examples)
				assert.True(t, fitness >= 0)
			}

			madeNonRecurrentInlink = true

		}

	}

	assert.True(t, madeNonRecurrentInlink)
	assert.True(t, madeRecurrentInlink)

}

func TestNeuronAddInlinkNonRecurrent(t *testing.T) {

	madeNonRecurrentInlink := false
	madeRecurrentInlink := false

	// since it's stochastic, repeat the operation many times and make
	// sure that it always produces expected behavior
	for i := 0; i < 100; i++ {

		xnorCortex := ng.XnorCortex()
		neuron := xnorCortex.NeuronUUIDMap()["output-neuron"]
		hiddenNeuron1 := xnorCortex.NeuronUUIDMap()["hidden-neuron1"]
		targetLayerIndex := hiddenNeuron1.NodeId.LayerIndex

		// add a new neuron at the same layer index as the hidden neurons
		hiddenNeuron3 := &ng.Neuron{
			ActivationFunction: ng.EncodableSigmoid(),
			NodeId:             ng.NewNeuronId("hidden-neuron3", targetLayerIndex),
			Bias:               -30,
		}

		shouldReInit := false
		hiddenNeuron3.Init(shouldReInit)
		xnorCortex.Neurons = append(xnorCortex.Neurons, hiddenNeuron3)

		inboundConnection := NeuronAddInlinkNonRecurrent(neuron)
		log.Printf("new inbound: %v", inboundConnection)
		if neuron.IsInboundConnectionRecurrent(inboundConnection) {
			madeRecurrentInlink = true
		} else {
			madeNonRecurrentInlink = true
		}
	}

	assert.True(t, madeNonRecurrentInlink)
	assert.False(t, madeRecurrentInlink)

}

func TestNeuronMutateWeights(t *testing.T) {

	xnorCortex := ng.XnorCortex()
	neuron := xnorCortex.NeuronUUIDMap()["output-neuron"]
	assert.True(t, neuron != nil)
	neuronCopy := neuron.Copy()

	foundModifiedWeight := false
	for i := 0; i < 100; i++ {

		didMutateWeights := NeuronMutateWeights(neuron)
		if didMutateWeights == true {

			foundModifiedWeight = verifyWeightsModified(neuron, neuronCopy)

		}

		if foundModifiedWeight == true {
			break
		}

	}

	assert.True(t, foundModifiedWeight == true)

}

func TestNeuronResetWeights(t *testing.T) {

	xnorCortex := ng.XnorCortex()
	neuron := xnorCortex.NeuronUUIDMap()["output-neuron"]
	assert.True(t, neuron != nil)
	neuronCopy := neuron.Copy()

	foundModifiedWeight := false
	for i := 0; i < 100; i++ {

		NeuronResetWeights(neuron)
		foundModifiedWeight = verifyWeightsModified(neuron, neuronCopy)

		if foundModifiedWeight == true {
			break
		}

	}

	assert.True(t, foundModifiedWeight == true)

}

func TestNeuronMutateActivation(t *testing.T) {

	ng.SeedRandom()
	neuron := &ng.Neuron{
		ActivationFunction: ng.EncodableSigmoid(),
		NodeId:             ng.NewNeuronId("neuron", 0.25),
		Bias:               10,
	}
	NeuronMutateActivation(neuron)
	assert.True(t, neuron.ActivationFunction != nil)
	assert.True(t, neuron.ActivationFunction.Name != ng.EncodableSigmoid().Name)

}

func TestNeuronRemoveBias(t *testing.T) {

	neuron := &ng.Neuron{
		ActivationFunction: ng.EncodableSigmoid(),
		NodeId:             ng.NewNeuronId("neuron", 0.25),
		Bias:               10,
	}
	shouldReInit := false
	neuron.Init(shouldReInit)
	NeuronRemoveBias(neuron)
	assert.True(t, neuron.Bias == 0)

}

func TestNeuronAddBias(t *testing.T) {

	// basic case where there is no bias

	neuron := &ng.Neuron{
		ActivationFunction: ng.EncodableSigmoid(),
		NodeId:             ng.NewNeuronId("neuron", 0.25),
	}
	shouldReInit := false
	neuron.Init(shouldReInit)

	NeuronAddBias(neuron)
	assert.True(t, neuron.Bias != 0)

	// make sure it treats 0 bias as not having a bias

	neuron = &ng.Neuron{
		ActivationFunction: ng.EncodableSigmoid(),
		NodeId:             ng.NewNeuronId("neuron", 0.25),
		Bias:               0,
	}
	neuron.Init(shouldReInit)

	NeuronAddBias(neuron)
	assert.True(t, neuron.Bias != 0)

	// make sure it doesn't add a bias if there is an existing one

	neuron = &ng.Neuron{
		ActivationFunction: ng.EncodableSigmoid(),
		NodeId:             ng.NewNeuronId("neuron", 0.25),
		Bias:               10,
	}
	neuron.Init(shouldReInit)
	NeuronAddBias(neuron)
	assert.True(t, neuron.Bias == 10)

}

func TestAddBias(t *testing.T) {

	// xnortCortex := ng.XnorCortex()

}