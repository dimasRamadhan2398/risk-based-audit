package department

import (
	"sort"
	"strings"

	"master-service/models"
	"master-service/pkg/orgmode"

	"github.com/google/uuid"
)

// Index is an in-memory view of a flexible tenant's department tree. Paths are
// derived from it, never stored (docs/org-structure-model-plan.md 4.3).
type Index struct {
	byID     map[uuid.UUID]*models.Department
	children map[uuid.UUID][]uuid.UUID
	levels   map[int]models.OrgLevel
}

// NewIndex builds the index. depts must carry ParentDepartmentID.
func NewIndex(depts []models.Department, levels []models.OrgLevel) *Index {
	ix := &Index{
		byID:     make(map[uuid.UUID]*models.Department, len(depts)),
		children: map[uuid.UUID][]uuid.UUID{},
		levels:   make(map[int]models.OrgLevel, len(levels)),
	}
	for i := range depts {
		d := depts[i]
		ix.byID[d.ID] = &d
	}
	for _, d := range ix.byID {
		if d.ParentDepartmentID != nil {
			if _, ok := ix.byID[*d.ParentDepartmentID]; ok {
				ix.children[*d.ParentDepartmentID] = append(ix.children[*d.ParentDepartmentID], d.ID)
			}
		}
	}
	for k := range ix.children {
		ids := ix.children[k]
		sort.Slice(ids, func(i, j int) bool { return ix.less(ids[i], ids[j]) })
	}
	for _, l := range levels {
		ix.levels[l.Level] = l
	}
	return ix
}

func (ix *Index) less(a, b uuid.UUID) bool {
	da, db := ix.byID[a], ix.byID[b]
	na, nb := strings.ToLower(da.DepartmentName), strings.ToLower(db.DepartmentName)
	if na != nb {
		return na < nb
	}
	return a.String() < b.String()
}

// Get returns the indexed department.
func (ix *Index) Get(id uuid.UUID) (*models.Department, bool) {
	d, ok := ix.byID[id]
	return d, ok
}

// Put adds or replaces a department (after a write inside the transaction).
func (ix *Index) Put(d models.Department) {
	if old, ok := ix.byID[d.ID]; ok && old.ParentDepartmentID != nil {
		p := *old.ParentDepartmentID
		kids := ix.children[p][:0:0]
		for _, c := range ix.children[p] {
			if c != d.ID {
				kids = append(kids, c)
			}
		}
		ix.children[p] = kids
	}
	ix.byID[d.ID] = &d
	if d.ParentDepartmentID != nil {
		if _, ok := ix.byID[*d.ParentDepartmentID]; ok {
			p := *d.ParentDepartmentID
			ix.children[p] = append(ix.children[p], d.ID)
			sort.Slice(ix.children[p], func(i, j int) bool { return ix.less(ix.children[p][i], ix.children[p][j]) })
		}
	}
}

// Level returns the org level with the given number.
func (ix *Index) Level(n int) (models.OrgLevel, bool) {
	l, ok := ix.levels[n]
	return l, ok
}

// ancestors returns the chain from the root down to id's parent. The walk is
// bounded so corrupt data (a cycle) cannot loop forever.
func (ix *Index) ancestors(id uuid.UUID) []*models.Department {
	var chain []*models.Department
	d, ok := ix.byID[id]
	for steps := 0; ok && d.ParentDepartmentID != nil && steps <= 4*orgmode.MaxLevels; steps++ {
		p, found := ix.byID[*d.ParentDepartmentID]
		if !found {
			break
		}
		chain = append(chain, p)
		d = p
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

// Depth is the number of ancestors + 1.
func (ix *Index) Depth(id uuid.UUID) int {
	return len(ix.ancestors(id)) + 1
}

// Height is the number of levels in id's subtree (1 for a leaf).
func (ix *Index) Height(id uuid.UUID) int {
	return ix.height(id, 0)
}

func (ix *Index) height(id uuid.UUID, depth int) int {
	h := 1
	if depth > 4*orgmode.MaxLevels { // corrupt data (cycle): stop
		return h
	}
	for _, c := range ix.children[id] {
		if ch := ix.height(c, depth+1) + 1; ch > h {
			h = ch
		}
	}
	return h
}

// Children returns the direct children of id.
func (ix *Index) Children(id uuid.UUID) []uuid.UUID {
	return ix.children[id]
}

// Descendants returns every department below id (breadth first).
func (ix *Index) Descendants(id uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{id: true}
	queue := append([]uuid.UUID(nil), ix.children[id]...)
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
		queue = append(queue, ix.children[c]...)
	}
	return out
}

func (ix *Index) seg(d *models.Department) models.DepartmentPathSeg {
	s := models.DepartmentPathSeg{ID: d.ID, Name: d.DepartmentName, Level: d.Level}
	if l, ok := ix.levels[d.Level]; ok {
		s.LevelName = l.Name
		s.NameKey = l.NameKey
	}
	return s
}

// Path is the derived path, root first, id itself last.
func (ix *Index) Path(id uuid.UUID) []models.DepartmentPathSeg {
	d, ok := ix.byID[id]
	if !ok {
		return nil
	}
	anc := ix.ancestors(id)
	path := make([]models.DepartmentPathSeg, 0, len(anc)+1)
	for _, a := range anc {
		path = append(path, ix.seg(a))
	}
	return append(path, ix.seg(d))
}

// Decorate fills the response-only fields of d from the index.
func (ix *Index) Decorate(d *models.Department) {
	if l, ok := ix.levels[d.Level]; ok {
		d.LevelName = l.Name
		d.NameKey = l.NameKey
	} else {
		d.LevelName = ""
		d.NameKey = nil
	}
	if _, ok := ix.byID[d.ID]; ok {
		d.Path = ix.Path(d.ID)
	} else {
		d.Path = []models.DepartmentPathSeg{ix.seg(d)}
	}
}

// Decorated returns a decorated copy of the indexed department.
func (ix *Index) Decorated(id uuid.UUID) *models.Department {
	d, ok := ix.byID[id]
	if !ok {
		return nil
	}
	cp := *d
	ix.Decorate(&cp)
	return &cp
}

// Tree returns the forest of the indexed departments. A department whose
// parent is not indexed is a root.
func (ix *Index) Tree() []*models.DepartmentTreeNode {
	var roots []uuid.UUID
	for id, d := range ix.byID {
		if d.ParentDepartmentID == nil {
			roots = append(roots, id)
			continue
		}
		if _, ok := ix.byID[*d.ParentDepartmentID]; !ok {
			roots = append(roots, id)
		}
	}
	sort.Slice(roots, func(i, j int) bool { return ix.less(roots[i], roots[j]) })
	out := make([]*models.DepartmentTreeNode, 0, len(roots))
	for _, r := range roots {
		out = append(out, ix.node(r, 0))
	}
	return out
}

func (ix *Index) node(id uuid.UUID, depth int) *models.DepartmentTreeNode {
	n := &models.DepartmentTreeNode{Department: *ix.Decorated(id), Children: []*models.DepartmentTreeNode{}}
	if depth > 4*orgmode.MaxLevels {
		return n
	}
	for _, c := range ix.children[id] {
		n.Children = append(n.Children, ix.node(c, depth+1))
	}
	return n
}
