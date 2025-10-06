package repositories

import (
	"crypto/md5"
	"fmt"
	"github.com/SHOWROOM-inc/recursive_mock_gen/models"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CodeRepository interface {
	LoadInterfaces(rootDir string) (models.Cache, error)
}

func NewCodeRepository() CodeRepository {
	return &codeRepository{}
}

type codeRepository struct {
}

func (r *codeRepository) LoadInterfaces(rootDir string) (models.Cache, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Dir: rootDir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	dst := models.Cache{}

	for _, pkg := range pkgs {
		if pkg.PkgPath == "" || pkg.Types == nil || len(pkg.Syntax) == 0 {
			continue
		}

		for i, file := range pkg.Syntax {
			fileName := pkg.CompiledGoFiles[i]
			if r.isExcludeFileName(fileName) {
				continue
			}

			for _, decl := range file.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok || gd.Tok != token.TYPE {
					continue
				}

				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, ok := ts.Type.(*ast.InterfaceType); !ok {
						continue
					}
					obj := pkg.Types.Scope().Lookup(ts.Name.Name)
					if obj == nil {
						continue
					}
					t := obj.Type().Underlying()
					iface, ok := t.(*types.Interface)
					if !ok {
						continue
					}
					iface = iface.Complete()

					entry := models.Entry{
						PkgPath:       pkg.PkgPath,
						PkgName:       pkg.Name,
						InterfaceName: obj.Name(),
						FilePath:      r.rel(fileName),
						Hash:          r.genHash(r.genCanonicalInterfaceString(pkg, obj.Name(), iface)),
					}
					dst.Append(entry)
				}
			}
		}

	}

	return dst, nil
}

// isExcludeFileName ファイル名が生成の対象外であればtrue, そうでなければfalseを返します。
func (r *codeRepository) isExcludeFileName(fileName string) bool {
	if strings.HasSuffix(fileName, "_mock.go") {
		return true
	}

	if strings.HasSuffix(fileName, "_test.go") {
		return true
	}

	return false
}

func (r *codeRepository) rel(p string) string {
	wd, _ := os.Getwd()
	if r, err := filepath.Rel(wd, p); err == nil {
		return r
	}
	return p
}

// genCanonicalInterfaceString 与えられたインタフェイスを正規化した文字列にします。
func (r *codeRepository) genCanonicalInterfaceString(pkg *packages.Package, name string, iface *types.Interface) string {
	var parts []string
	parts = append(parts, "package="+pkg.PkgPath, "interface="+name)

	// メソッドは名前順にする
	n := iface.NumMethods()
	methods := make([]*types.Func, 0, n)
	for i := 0; i < n; i++ {
		methods = append(methods, iface.Method(i))
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].Name() < methods[j].Name() })

	qual := func(other *types.Package) string {
		// 完全修飾（インポートパス）で文字列化
		if other == nil {
			return ""
		}
		return other.Path()
	}

	for _, m := range methods {
		s := types.TypeString(m.Type(), qual) // シグネチャ（ジェネリクス対応）
		parts = append(parts, fmt.Sprintf("method=%s%s", m.Name(), s))
	}
	// 埋め込みインターフェイス由来メソッドも含めて Complete() 済みなのでOK
	return strings.Join(parts, ";")
}

func (r *codeRepository) genHash(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
