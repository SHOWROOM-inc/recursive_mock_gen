package models

import "path"

type Entry struct {
	PkgPath       string `json:"pkg_path"`
	PkgName       string `json:"pkg_name"`
	InterfaceName string `json:"interface_name"`
	FilePath      string `json:"file_path"`
	Hash          string `json:"hash"`
}

func (e Entry) OutPackageName() string {
	return "mock_" + e.PkgName
}

func (e Entry) OutFilePath(basePath string) string {
	return path.Join(basePath, e.FilePath)
}

type Cache map[string]Entry

func (c Cache) Append(e Entry) {
	key := e.PkgName + "." + e.InterfaceName
	c[key] = e
}

// Diff 引数newSrcと比較して、変更されたエントリと削除されたエントリを比較します
func (c Cache) Diff(newSrc Cache) (changed Cache, deleted Cache) {
	changed = Cache{}
	deleted = Cache{}

	// newSrcに新規追加されたモノ、変更されたモノをchangedに入れる
	for key, srcEntry := range newSrc {
		myEntry, ok := c[key]
		if !ok {
			changed.Append(srcEntry)
			continue
		}
		if myEntry.Hash != srcEntry.Hash {
			changed.Append(srcEntry)
			continue
		}
	}

	// newSrcでは削除されたモノをdeleted入れる
	for key, entry := range c {
		_, ok := newSrc[key]
		if !ok {
			deleted.Append(entry)
		}
	}

	return
}
