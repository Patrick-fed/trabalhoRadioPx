package voice

import "testing"

// B1 / RF03: o voiceGuard deve permitir apenas um falante por canal - o
// segundo StartTransmission e rejeitado (ErrChannelBusy), so o atual speaker
// pode parar, e a parada (ou o release no disconnect) libera o canal.
func TestVoiceService_SingleSpeakerGuard(t *testing.T) {
	vs := NewVoiceService()

	if err := vs.StartTransmission("u1", "ch1"); err != nil {
		t.Fatalf("first StartTransmission: %v", err)
	}
	if !vs.IsChannelBusy("ch1") {
		t.Error("expected channel busy after start")
	}
	if vs.CanTransmit("u2", "ch1") {
		t.Error("u2 must not transmit while u1 is speaking")
	}
	if !vs.CanTransmit("u1", "ch1") {
		t.Error("u1 (current speaker) must be able to keep transmitting")
	}

	if err := vs.StartTransmission("u2", "ch1"); err != ErrChannelBusy {
		t.Errorf("second speaker: got %v, want ErrChannelBusy", err)
	}
	if vs.GetCurrentSpeaker("ch1") != "u1" {
		t.Errorf("speaker must remain u1, got %q", vs.GetCurrentSpeaker("ch1"))
	}

	// Reenvio do proprio speaker (retry de PTT) nao pode falhar.
	if err := vs.StartTransmission("u1", "ch1"); err != nil {
		t.Errorf("same speaker restart: %v", err)
	}

	// Somente o speaker atual pode liberar o canal.
	if err := vs.StopTransmission("u2", "ch1"); err != ErrNotSpeaker {
		t.Errorf("stop by non-speaker: got %v, want ErrNotSpeaker", err)
	}
	if !vs.IsChannelBusy("ch1") {
		t.Error("failed stop must not release the channel")
	}

	if err := vs.StopTransmission("u1", "ch1"); err != nil {
		t.Fatalf("stop by speaker: %v", err)
	}
	if vs.IsChannelBusy("ch1") {
		t.Error("expected channel released after stop")
	}
	if !vs.CanTransmit("u2", "ch1") {
		t.Error("u2 must be free to transmit after release")
	}
	if err := vs.StopTransmission("u1", "ch1"); err != ErrNotSpeaker {
		t.Errorf("double stop: got %v, want ErrNotSpeaker", err)
	}
	if err := vs.StopTransmission("u1", "unknown"); err != ErrChannelNotFound {
		t.Errorf("stop on unknown channel: got %v, want ErrChannelNotFound", err)
	}
}

func TestVoiceService_ActiveChannelsCount(t *testing.T) {
	vs := NewVoiceService()

	if n := vs.ActiveChannelsCount(); n != 0 {
		t.Fatalf("fresh service: got %d active, want 0", n)
	}
	vs.StartTransmission("u1", "ch1")
	vs.StartTransmission("u2", "ch2")
	if n := vs.ActiveChannelsCount(); n != 2 {
		t.Errorf("two transmitting channels: got %d, want 2", n)
	}
	vs.StopTransmission("u1", "ch1")
	if n := vs.ActiveChannelsCount(); n != 1 {
		t.Errorf("after stopping ch1: got %d, want 1", n)
	}
}
