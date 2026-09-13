package minimal

import "testing"

func TestMinimal(t *testing.T) {
	m := New()

	err := m.Subln("Hello from Subln")
	if err != nil {
		t.Fatal(err)
	}

	err = m.Infoln("Hello from Infoln")
	if err != nil {
		t.Fatal(err)
	}

	err = m.Successln("Hello from Successln")
	if err != nil {
		t.Fatal(err)
	}

	err = m.Warnln("Hello from Warnln")
	if err != nil {
		t.Fatal(err)
	}

	err = m.Errorln("Hello from Errorln")
	if err != nil {
		t.Fatal(err)
	}
}
