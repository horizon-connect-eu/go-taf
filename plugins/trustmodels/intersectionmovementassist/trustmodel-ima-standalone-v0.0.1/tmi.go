package trustmodel_ima_standalone_v0_0_1

import (
	"fmt"
	"github.com/vs-uulm/go-subjectivelogic/pkg/subjectivelogic"
	"github.com/vs-uulm/go-taf/pkg/core"
	internaltrustmodelstructure "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelstructure"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"github.com/vs-uulm/taf-tlee-interface/pkg/trustmodelstructure"
	"hash/fnv"
	"regexp"
	"sort"
	"strings"
)

type TrustModelInstance struct {
	id       string
	version  int
	template TrustModelTemplate

	sourceID      string
	sourceOpinion subjectivelogic.QueryableOpinion            // Opinion V_ego -> V_sourceID
	objects       map[string]subjectivelogic.QueryableOpinion // X : Opinion V_ego -> C_sourceID_{X}

	currentStructure   trustmodelstructure.TrustGraphStructure
	currentValues      map[string][]trustmodelstructure.TrustRelationship
	currentFingerprint uint32
	rtls               map[string]subjectivelogic.QueryableOpinion
	staticRTL          subjectivelogic.QueryableOpinion
}

func (e *TrustModelInstance) ID() string {
	return e.id
}

func (e *TrustModelInstance) Version() int {
	return e.version
}

func (e *TrustModelInstance) Fingerprint() uint32 {
	return e.currentFingerprint
}

func (e *TrustModelInstance) Template() core.TrustModelTemplate {
	return e.template
}

func (e *TrustModelInstance) Update(update core.Update) bool {
	oldVersion := e.Version()
	switch update := update.(type) {
	case trustmodelupdate.RefreshCPM:
		topologyIsModified := e.processTopologyUpdate(update.Objects())
		if topologyIsModified {
			e.updateStructure()
			e.updateFingerprint()
			e.incrementVersion()
			e.updateValues()
		}
	case trustmodelupdate.UpdateAtomicTrustOpinion:
		trustee := update.Trustee()
		if strings.HasPrefix(trustee, "V_") {
			id, err := parseVehicleIdentifier(trustee)
			if err == nil && id == e.sourceID {

				e.sourceOpinion = update.Opinion()
				e.updateValues()
				e.incrementVersion()
			}
		} else if strings.HasPrefix(trustee, "C_") {
			_, objID, err := parseObjectIdentifier(trustee)
			if err == nil {
				if _, ok := e.objects[objID]; ok {
					e.objects[objID] = update.Opinion()
					e.updateValues()
					e.incrementVersion()
				}
			}
		}
	default:
		//ignore
	}
	return oldVersion != e.Version() //when version has changed, indicate to run TLEE
}

func (e *TrustModelInstance) incrementVersion() int {
	e.version = e.version + 1
	return e.version
}

/*
processTopologyUpdate reflects changes in the internal topology based upon the latest objects received in an update.
In case there are topology changes due to the update, the function returns true. Otherwise, if the topology is not affected, it returns false.
*/
func (e *TrustModelInstance) processTopologyUpdate(latestObjects []string) bool {
	topologyChanged := false

	addedObjects := make(map[string]struct{})
	removedObjects := make(map[string]struct{}) //old objects will be placed here and will be deleted in case they are still used

	for obj := range e.objects {
		if obj != e.sourceID { //never remove the observation of vehicle on itself (C_{X}_{X})
			removedObjects[obj] = struct{}{}
		}
	}

	for _, object := range latestObjects {
		if _, ok := e.objects[object]; ok {
			//object existed before, so remove from the set of missing objects
			delete(removedObjects, object)
		} else {
			//object is not yet known, so needs to be added
			addedObjects[object] = struct{}{}
		}
	}

	//For new objects: Add to topology with full uncertainty
	if len(addedObjects) > 0 {
		topologyChanged = true
		for object := range addedObjects {
			e.objects[object] = &FullUncertainty
		}

	}
	/* Currently, we keep disappearing IDs. The following code would drop them.
	if len(removedObjects) > 0 {
		topologyChanged = true
		for object, _ := range removedObjects {
			delete(e.objects, object)
		}
	}
	*/
	return topologyChanged
}

/*
updateFingerprint calculates the current fingerprint for the TMI.
Therefore, it takes the all the dynamic nodes, concatenates their
sorted string identifiers and calculates a hash value.
*/
func (e *TrustModelInstance) updateFingerprint() {
	nodes := make([]string, 0)
	for object := range e.objects {
		nodes = append(nodes, object)
	}

	sort.Strings(nodes)
	stringFingerprint := strings.Join(nodes, "")

	algorithm := fnv.New32a()
	_, err := algorithm.Write([]byte(stringFingerprint))
	if err == nil {
		e.currentFingerprint = algorithm.Sum32()
	}
}

/*
updateStructure updates the internally kept structure according to the latest topology.
*/
func (e *TrustModelInstance) updateStructure() {
	//Objects (observations) that originate from the sender vehicle
	objects := make([]string, 0)
	//Direct edges from the ego node to all others
	egoTargets := make([]string, 0)
	for object := range e.objects {
		objects = append(objects, objectIdentifier(object, e.sourceID))
		egoTargets = append(egoTargets, objectIdentifier(object, e.sourceID))
	}
	egoTargets = append(egoTargets, vehicleIdentifier(e.sourceID))

	e.currentStructure = internaltrustmodelstructure.NewTrustGraphDTO(trustmodelstructure.CumulativeFusion, []trustmodelstructure.AdjacencyListEntry{
		internaltrustmodelstructure.NewAdjacencyEntryDTO(vehicleIdentifier("ego"), egoTargets),
		internaltrustmodelstructure.NewAdjacencyEntryDTO(vehicleIdentifier(e.sourceID), objects),
	})
}

/*
updateValues updates the internally kept values according to the latest state. Will also dynamically set RTL map to fixed RTL.
*/
func (e *TrustModelInstance) updateValues() {
	values := make(map[string][]trustmodelstructure.TrustRelationship)
	rtls := make(map[string]subjectivelogic.QueryableOpinion)

	for obj, opinion := range e.objects {

		ego := vehicleIdentifier("ego")
		source := vehicleIdentifier(e.sourceID)
		observation := objectIdentifier(obj, e.sourceID)
		scope := observation

		//set values
		values[scope] = []trustmodelstructure.TrustRelationship{
			//full belief between V_* and C_*_*
			internaltrustmodelstructure.NewTrustRelationshipDTO(source, observation, &FullBelief),
			//opinion from V_ego on C_*_*
			internaltrustmodelstructure.NewTrustRelationshipDTO(ego, observation, opinion),
			//opinion from V_y on C_y_*
			internaltrustmodelstructure.NewTrustRelationshipDTO(ego, source, e.sourceOpinion),
		}

		//set RTL
		rtls[observation] = &RTL
	}

	e.currentValues = values
	e.rtls = rtls
}

func (e *TrustModelInstance) Initialize(params map[string]interface{}) {
	//If a source ID has been defined, use it; otherwise, use ID of TMI
	sourceId, exists := params["SourceId"]
	if !exists {
		e.sourceID = e.id
	} else {
		e.sourceID = sourceId.(string)
	}
	e.version = 0
	e.currentFingerprint = 0
	e.rtls = map[string]subjectivelogic.QueryableOpinion{}
	e.sourceOpinion = &FullUncertainty
	e.objects[e.sourceID] = &FullUncertainty

	e.updateStructure()
	e.updateFingerprint()
	e.updateValues()
	return
}

func (e *TrustModelInstance) Cleanup() {
	//nothing to do here (yet)
	return
}

func (e *TrustModelInstance) Structure() trustmodelstructure.TrustGraphStructure {
	return e.currentStructure
}

func (e *TrustModelInstance) Values() map[string][]trustmodelstructure.TrustRelationship {

	/*
		//This code is a quick-n-dirty fix to set V-ego => V_X to full belief and V_X to C_X_? to the value of V-ego => V_X
			modifiedValues := make(map[string][]trustmodelstructure.TrustRelationship)
			for k, v := range e.currentValues {
				rels := make([]trustmodelstructure.TrustRelationship, 0)

				var egoToVehicle subjectivelogic.QueryableOpinion
				for _, rel := range v {
					if rel.Source() == "V_ego" && strings.HasPrefix(rel.Destination(), "V_") {
						egoToVehicle = rel.Opinion()
					}
				}

				for _, rel := range v {
					if rel.Source() == "V_ego" && strings.HasPrefix(rel.Destination(), "V_") {
						rels = append(rels, internaltrustmodelstructure.NewTrustRelationshipDTO(rel.Source(), rel.Destination(), &FullBelief))
					} else if strings.HasPrefix(rel.Source(), "V_") && strings.HasPrefix(rel.Destination(), "C_") {
						rels = append(rels, internaltrustmodelstructure.NewTrustRelationshipDTO(rel.Source(), rel.Destination(), egoToVehicle))
					} else {
						rels = append(rels, rel)
					}
				}
				modifiedValues[k] = rels
			}
			return modifiedValues
	*/
	return e.currentValues
}

func (e *TrustModelInstance) RTLs() map[string]subjectivelogic.QueryableOpinion {
	return e.rtls
}

/*
vehicleIdentifier is a helper function to turn a plain identifier into an identifier for vehicles used in the structure.
*/
func vehicleIdentifier(id string) string {
	return fmt.Sprintf("V_%s", id)
}

/*
objectIdentifier is a helper function to turn a plain identifier into an identifier for objects/observations used in the structure.
*/
func objectIdentifier(id string, source string) string {
	return fmt.Sprintf("C_%s_%s", source, id)
}

/*
parseObjectIdentifier is a helper function to extract plain identifiers from an object identifier string.
*/
func parseObjectIdentifier(str string) (string, string, error) {
	pattern := regexp.MustCompile(`^C_(\d+)_(\d+)$`)
	res := pattern.FindStringSubmatch(str)
	if res != nil && len(res) == 3 {
		return res[1], res[2], nil
	} else {
		return "", "", fmt.Errorf("Invalid object identifier '" + str + "'")
	}
}

/*
parseVehicleIdentifier is a helper function to extract plain identifiers from a vehicle identifier string.
*/
func parseVehicleIdentifier(str string) (string, error) {
	pattern := regexp.MustCompile(`^V_(\d+|ego).*$`)
	res := pattern.FindStringSubmatch(str)
	if res != nil && len(res) == 2 {
		return res[1], nil
	} else {
		return "", fmt.Errorf("Invalid vehicle identifier '" + str + "'")
	}
}

func (e *TrustModelInstance) String() string {
	return core.TMIAsString(e)
}
