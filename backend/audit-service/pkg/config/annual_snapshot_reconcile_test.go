package config

import "testing"

func TestAnnualSnapshotReconcileOnStartDefaultsOn(t *testing.T) {
	t.Setenv(AnnualSnapshotReconcileOnStartEnv, "")
	cfg, err := Load(writeConfig(t, "jwt:\n  secret: x\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AnnualSnapshotReconcile.OnStart {
		t.Fatal("start-up reconcile is off by default, want on")
	}
}

func TestAnnualSnapshotReconcileOnStartEnvTurnsOff(t *testing.T) {
	t.Setenv(AnnualSnapshotReconcileOnStartEnv, "false")
	cfg, err := Load(writeConfig(t, "jwt:\n  secret: x\nannual_snapshot_reconcile:\n  on_start: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnnualSnapshotReconcile.OnStart {
		t.Fatalf("%s=false did not override the file", AnnualSnapshotReconcileOnStartEnv)
	}
}

func TestAnnualSnapshotReconcileOnStartFileOff(t *testing.T) {
	t.Setenv(AnnualSnapshotReconcileOnStartEnv, "")
	cfg, err := Load(writeConfig(t, "jwt:\n  secret: x\nannual_snapshot_reconcile:\n  on_start: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnnualSnapshotReconcile.OnStart {
		t.Fatal("annual_snapshot_reconcile.on_start: false in the file was ignored")
	}
}
