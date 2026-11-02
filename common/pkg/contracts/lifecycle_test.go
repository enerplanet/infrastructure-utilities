package contracts

import "testing"

func TestCanTransition_runsBuildingsBeforeTheModel(t *testing.T) {
	for _, step := range [][2]string{{StatusQueue, StatusBuildings}, {StatusBuildings, StatusRunning}, {StatusBuildings, StatusFailed}} {
		if !CanTransition(step[0], step[1]) {
			t.Errorf("%s -> %s refused", step[0], step[1])
		}
	}
	if CanTransition(StatusBuildings, StatusCompleted) {
		t.Errorf("%s -> %s allowed before the model ran", StatusBuildings, StatusCompleted)
	}
}
