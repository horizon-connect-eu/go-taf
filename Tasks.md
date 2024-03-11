# TAF Demonstrator & Scalability Test

## High-Level Description

The demonstrater is a simplified architecture of the current TAF architecture that takes advantage of the concepts to be used for the actual implementation.
It is based upon multiple components that pass entries via channels and use these entries to update an internal state.
After certain updates, the state is used for an internal computation.


## External Events

The external events are strings and have the following structure:

`<ID as int>,<value as int>,<A/B as String>`



## Event Source Component

The first component takes these events and turns them into internal structs.
And these structs are then send to both downstream components (A/B).
In the actual TAF, this would be something like the V2X message listener.


## TSM/TMM Replacements

The TSM/TMM replacements are components that receive events from the event source and now have to decide whether this event is relevant, and if so, what to do with it.
In our case, there are two instances of this component (A and B).
Only if the event is associated with this component (having A or B as the third parameter), it will forward the event to the next downstream component (TAM replacement).

## TAM Replacement

The TAM replacement has several jobs: 
(1) It takes events from the TSM/TMM replacements and process them. Therefore, it takes the event, and extracts the ID. The TAM also maintains a map of entries (*states map*) in which the key represents the ID, and the value is a slice of the ten latest entries received for this ID. So upon receiving a new event, this component needs to find the correct entry for the ID and adjust the content of the slice by adding the value which is part of the event. 
(2) After an entry has been updated, the component runs a pseudo-TLEE calculation. In our case, this means that the component provides a function that takes a slice as an input and returns the sum of the slice. Now after an update, the component calls this function with the ten latest entries associated with the ID that has been updated. The result should then be stored in a second map (*result map*) that contains the ID as a key and the latest computation result a as value.
(3) Another function allows to query results for a given ID  from the result map. 



## Steps 

### 1. Event Source

Create an event source component that generates 5 random events per second.
The range of IDs should include 0-99 as values.
Create them as structs first, and then turn them into proper structs.
Also create the A and B components, that will print out the values upon arrival.


### 2. State Updates

Change the A and B components so that will forward matching event to the TAM replacement.
In turn, the TAM replacement should do the update operation of the states map according to the description above.


### 3. Computations

Modify the TAM replacement so that after any update, the pseudo-TLEE calculation is executed as well. 
Store the output in the result map.

### 4. Querying

Add yet another component that uses the command line interface. Here, the user can enter numbers between 0 and 99. When submitted, this component gets the latest result for the entry with this number as ID from the TAM replacement.

#### 5. Sharding

Change the TAM replacement functionality so that state handling is parallelized to a configurable amount of shards.