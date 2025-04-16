package go_routeros

import "testing"

func TestNormalizeToCommandLineByPrint(t *testing.T) {
	tests := []struct {
		name     string
		menu     string
		args     []string
		expected string
	}{
		{
			"Print simples com um filtro",
			"/ip/firewall/address-list/print",
			[]string{"?address=192.168.0.1"},
			"/ip/firewall/address-list/print where address=192.168.0.1",
		},
		{
			"Print com dois filtros",
			"/ip/firewall/address-list/print",
			[]string{"?address=192.168.0.1", "?list=aviso"},
			"/ip/firewall/address-list/print where address=192.168.0.1 and list=aviso",
		},
		{
			"Print com OR",
			"/ip/firewall/address-list/print",
			[]string{"?list=aviso", "?list=block", "?#|"},
			"/ip/firewall/address-list/print where list=aviso or list=block",
		},
		{
			"Remove com AND",
			"/ip/firewall/address-list/remove",
			[]string{"?list=aviso", "?disabled=yes", "?#&"},
			"/ip/firewall/address-list/remove [find list=aviso and disabled=yes]",
		},
		{
			"Print com OR agrupado e AND externo",
			"/ip/firewall/address-list/print",
			[]string{"?list=aviso", "?list=block", "?#|", "?disabled=no", "?#&"},
			"/ip/firewall/address-list/print where (list=aviso or list=block) and disabled=no",
		},
		{
			"Print com OR agrupado e AND externo",
			"/ip/firewall/address-list/print",
			[]string{"?list=aviso", "?list=block", "?#|", "?disabled=no", "?#&"},
			"/ip/firewall/address-list/print where (list=aviso or list=block) and disabled=no",
		},
		{
			"Set com OR agrupado e AND externo",
			"/ip/firewall/address-list/set",
			[]string{"=address=1.1.1.2", "?list=aviso", "?list=block", "?#|", "?disabled=no", "?#&"},
			"/ip/firewall/address-list/set address=1.1.1.2 [find (list=aviso or list=block) and disabled=no]",
		},
		{
			"Set com OR agrupado e AND externo",
			"/ip/firewall/address-list/add",
			[]string{"=address=1.1.1.2", "=address=1.1.1.2", "=list=aviso"},
			"/ip/firewall/address-list/add address=1.1.1.2 list=aviso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeToCommandLine(tt.menu, tt.args...)
			if result != tt.expected {
				t.Errorf("expected: %s, got: %s", tt.expected, result)
			}
		})
	}
}
