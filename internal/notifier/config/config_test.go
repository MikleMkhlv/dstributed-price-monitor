package notifier_test

import (
	notifier "dstributed-price-monitor/internal/notifier/config"
	"encoding/json"
	"testing"
)

func TestCreateNotifierConfig(t *testing.T) {
	cfg, err := notifier.LoadNotifierConfig("../../../conf_notifier_test.yaml")
	if err != nil {
		t.Errorf("createNotifierConfig: error load notifier config. %v", err)
		return
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Errorf("createNotifierConfig: error marhal cfg struct")
		return
	}

	t.Logf("%s", string(data))
	if len(data) == 0 {
		t.Error("conf is empty")
		return
	}
}
