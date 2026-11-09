package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/kackerx/sqlmux/internal/db"
	"github.com/kackerx/sqlmux/internal/i18n"
)

// A connection whose first catalog is on its way spins on its node and says
// Loading… under it, as DBeaver does; the catalog in, the tables are there and
// the spinning stops. A reload (R) of one already loaded spins on the node alone
// (§7.8「转圈」).
func TestCatalogLoading(t *testing.T) {
	a := sized(160, 45, "nerd")
	c := readyConn("slow")
	a.sess.Conns = append(a.sess.Conns, c)
	node := func() node {
		ns, _ := a.treeNodes()
		for _, n := range ns {
			if n.kind == nodeConn && n.conn == c {
				return n
			}
		}
		t.Fatal("no node of the connection")
		return node{}
	}
	if a.spinning() {
		t.Fatal("nothing spins before the load")
	}

	load := a.loadCatalog(c)
	if load == nil || c.loading != 1 {
		t.Fatalf("loading %d", c.loading)
	}
	if !a.spinning() || node().Aside == "" {
		t.Errorf("the node spins: aside %q, spinning %v", node().Aside, a.spinning())
	}
	if tree := treeTexts(a); !strings.Contains(tree, i18n.T("加载中…")) {
		t.Errorf("no Loading… under it:\n%s", tree)
	}

	a.Update(catalogMsg{conn: c, schemas: []string{"public"}, current: "public", tables: []db.Table{{Schema: "public", Name: "orders"}}})
	if c.loading != 0 || node().Aside != "" || a.spinning() {
		t.Errorf("loaded: loading %d, aside %q, spinning %v", c.loading, node().Aside, a.spinning())
	}
	if tree := treeTexts(a); strings.Contains(tree, i18n.T("加载中…")) || !strings.Contains(tree, "orders") {
		t.Errorf("after the load:\n%s", tree)
	}

	a.loadCatalog(c) // R: what it has stays, the node spins
	if tree := treeTexts(a); strings.Contains(tree, i18n.T("加载中…")) || !strings.Contains(tree, "orders") || node().Aside == "" {
		t.Errorf("reloading:\n%s", tree)
	}
	a.Update(catalogMsg{conn: c, err: errors.New("timeout")})
	if c.loading != 0 || node().Aside != "" {
		t.Errorf("a failed load stops spinning: %d", c.loading)
	}
}
