package models

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestEntry_OutPackageName(t *testing.T) {
    e := Entry{PkgName: "sample"}
    assert.Equal(t, "mock_sample", e.OutPackageName())
}

func TestEntry_OutFilePath(t *testing.T) {
    tests := []struct {
        name     string
        base     string
        filePath string
        want     string
    }{
        {name: "simple join", base: "base", filePath: "dir/file.go", want: "base/dir/file.go"},
        {name: "no base", base: "", filePath: "file.go", want: "/file.go"[1:]}, // path.Join("", x) == x
        {name: "nested", base: "root/base", filePath: "a/b/c.go", want: "root/base/a/b/c.go"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e := Entry{FilePath: tt.filePath}
            got := e.OutFilePath(tt.base)
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestCache_Append(t *testing.T) {
    c := Cache{}
    e := Entry{PkgName: "pkg", InterfaceName: "IF", Hash: "h"}
    c.Append(e)

    key := e.PkgName + "." + e.InterfaceName
    val, ok := c[key]
    assert.True(t, ok)
    assert.Equal(t, e, val)
}

func TestCache_Diff(t *testing.T) {
    t.Run("added entry appears in changed", func(t *testing.T) {
        old := Cache{}
        eNew := Entry{PkgName: "p", InterfaceName: "I", Hash: "h1"}
        newer := Cache{}
        newer.Append(eNew)

        changed, deleted := old.Diff(newer)
        assert.Len(t, changed, 1)
        assert.Contains(t, changed, "p.I")
        assert.Empty(t, deleted)
    })

    t.Run("hash changed appears in changed", func(t *testing.T) {
        old := Cache{}
        eOld := Entry{PkgName: "p", InterfaceName: "I", Hash: "old"}
        old.Append(eOld)
        newer := Cache{}
        eNew := Entry{PkgName: "p", InterfaceName: "I", Hash: "new"}
        newer.Append(eNew)

        changed, deleted := old.Diff(newer)
        assert.Len(t, changed, 1)
        assert.Equal(t, eNew, changed["p.I"])
        assert.Empty(t, deleted)
    })

    t.Run("unchanged produces empty diffs", func(t *testing.T) {
        old := Cache{}
        e := Entry{PkgName: "p", InterfaceName: "I", Hash: "same"}
        old.Append(e)
        newer := Cache{}
        newer.Append(e)

        changed, deleted := old.Diff(newer)
        assert.Empty(t, changed)
        assert.Empty(t, deleted)
    })

    t.Run("deleted entry appears in deleted", func(t *testing.T) {
        old := Cache{}
        e := Entry{PkgName: "p", InterfaceName: "I", Hash: "h"}
        old.Append(e)
        newer := Cache{}

        changed, deleted := old.Diff(newer)
        assert.Empty(t, changed)
        assert.Len(t, deleted, 1)
        assert.Equal(t, e, deleted["p.I"])
    })

    t.Run("mixed added/changed/deleted", func(t *testing.T) {
        old := Cache{}
        eA := Entry{PkgName: "a", InterfaceName: "I", Hash: "x"} // unchanged
        eB := Entry{PkgName: "b", InterfaceName: "I", Hash: "y"} // will change
        eD := Entry{PkgName: "d", InterfaceName: "I", Hash: "del"} // will be deleted
        old.Append(eA)
        old.Append(eB)
        old.Append(eD)

        newer := Cache{}
        newer.Append(eA)                                                      // unchanged
        newer.Append(Entry{PkgName: "b", InterfaceName: "I", Hash: "y2"}) // changed
        newer.Append(Entry{PkgName: "c", InterfaceName: "I", Hash: "z"})  // added

        changed, deleted := old.Diff(newer)

        // changed should contain b.I with new hash and c.I
        assert.Len(t, changed, 2)
        assert.Equal(t, Entry{PkgName: "b", InterfaceName: "I", Hash: "y2"}, changed["b.I"])
        assert.Equal(t, Entry{PkgName: "c", InterfaceName: "I", Hash: "z"}, changed["c.I"])

        // deleted should contain d.I
        assert.Len(t, deleted, 1)
        assert.Equal(t, eD, deleted["d.I"])
    })
}
