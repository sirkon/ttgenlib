package generator

import (
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/sirkon/errors"
	"github.com/sirkon/gogh"
	"github.com/sirkon/jsonexec"
	"github.com/sirkon/message"
	"github.com/sirkon/ttgenlib/internal/ordmap"
	"golang.org/x/tools/go/packages"
)

// Generator a facility for code processing and generation.
type Generator struct {
	path       string
	fset       *token.FileSet
	pkg        *packages.Package
	pkgs       map[string]*packages.Package
	mockLookup MockLookup

	nomock      []doNotMock
	m           *gogh.Module[*gogh.Imports]
	mockerNames func(tn *types.TypeName) (filename string, typename string)

	preTest func(r *goRenderer)
	ctxInit func(r *goRenderer)
	msgr    LoggingRenderer
}

func newGenerator(
	pkg string,
	mockLookup MockLookup,
	msgsRenderer LoggingRenderer,
	opts ...Option,
) (*Generator, error) {
	var goList struct {
		ImportPath string
	}
	if err := jsonexec.Run(&goList, "go", "list", "--json", pkg); err != nil {
		return nil, errors.Wrapf(err, "get package '%s' info", pkg)
	}

	cfg := &packages.Config{
		Mode:    PackageLoadMode,
		Context: nil,
		Logf: func(format string, args ...interface{}) {
			message.Infof(format, args...)
		},
		Tests: false,
	}
	res, err := packages.Load(cfg, goList.ImportPath)
	if err != nil {
		return nil, errors.Wrap(err, "parse package")
	}

	g := &Generator{
		path:       goList.ImportPath,
		pkg:        nil,
		mockLookup: mockLookup,
		pkgs:       map[string]*packages.Package{},
		nomock:     []doNotMock{contextNoMock},
		mockerNames: func(tn *types.TypeName) (filename string, typename string) {
			filename = gogh.Underscored(tn.Name(), "mocker", "test") + ".go"
			typename = gogh.Private(tn.Name(), "mocker")

			return filename, typename
		},
		preTest: func(r *goRenderer) {},
		ctxInit: func(r *goRenderer) {
			r.Imports().Add("context").Ref("ctx")
			r.L(`ctx := $ctx.Background()`)
		},
		msgr: msgsRenderer,
	}
	for _, pg := range res {
		if pg.PkgPath == goList.ImportPath {
			g.pkg = pg
			g.fset = g.pkg.Fset
		}
		g.pkgs[pg.PkgPath] = pg
	}

	for _, opt := range opts {
		if err := opt(g, optionRestriction{}); err != nil {
			return nil, errors.Wrap(err, "apply an options")
		}
	}

	m, err := gogh.New(
		gogh.FancyFmt,
		func(r *gogh.Imports) *gogh.Imports {
			return r
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "init code renderer for the module")
	}
	g.m = m

	return g, nil
}

func (g *Generator) generate(p *goPackage, r *goRenderer, f types.Object) error {
	s := f.Type().(*types.Signature)

	typeMocks, err := g.getMocksOfType(s)
	if err != nil {
		return errors.Wrap(err, "get mocks of type")
	}

	if len(typeMocks) > 0 {
		if err := g.generateTypeMocker(p, s, typeMocks); err != nil {
			return errors.Wrap(err, "generate mocker for the type")
		}
	}

	paramMocks, err := g.getMocksOfArguments(s, f)
	if err != nil {
		return errors.Wrap(err, "get mocks for arguments")
	}

	g.generateTest(r, f, len(typeMocks) > 0, paramMocks)

	return nil
}

func (g *Generator) generateTest(
	r *goRenderer,
	f types.Object,
	hasMocksInType bool,
	amocks []MockLookupResult,
) {
	s := f.Type().(*types.Signature)
	var mtype *types.Named
	if hasMocksInType {
		mtype = s.Recv().Type().(*types.Pointer).Elem().(*types.Named)
		_, mockertype := g.mockerNames(mtype.Obj())
		r.Let("mockertype", mockertype)
	}

	r.Imports().Add("testing").Ref("tst")

	if mtype != nil {
		r.L(`func Test${0}${1}(t *${tst}.T) {`, mtype.Obj().Name(), f.Name())
	} else {
		r.L(`func Test${0}(t *${tst}.T) {`, f.Name())
	}

	argfields, resfields, errcheck := g.renderTestStructure(r, hasMocksInType, mtype, amocks, s)

	r.N()
	r.L(`    tests := []test{}`)
	r.L(`    for _, tt := range tests {`)
	r.L(`        tt := tt`)
	r.L(`        t.Run(tt.name, func(t *$tst.T) {`)
	if hasMocksInType || len(amocks) > 0 {
		r.Imports().Add(gomockPath).Ref("gomock")
		r.L(`            ctrl := $gomock.NewController(t)`)
	}
	g.preTest(r)
	if hasContextArg(s) {
		g.ctxInit(r)
	}
	if hasMocksInType {
		r.L(`            m := new${mockertype|P}(ctrl)`, mtype.Obj().Name())
		r.L(`            x := m.$0()`, mtype.Obj().Name())
	} else if mtype != nil {
		r.L(`            var x *$type // User change required, it is unclear how to create it properly'.`)
	}
	if len(amocks) > 0 {
		r.L(`            amocks := argMocks{`)
		for _, amock := range amocks {
			r.L(`            $0: $1(ctrl),`, amock.Name, strings.TrimSuffix(r.Type(amock.Named), amock.Named.Obj().Name())+amock.Constructor.Name())
		}
		r.L(`            }`)
	}

	// Must call setup with proper parameters.
	if hasMocksInType || len(amocks) > 0 {
		sp := &gogh.Commas{}
		sp.Add("ctrl")
		sp.Add("&tt")
		if hasMocksInType {
			sp.Add("m")
		}
		if len(amocks) > 0 {
			sp.Add("&amocks")
		}
		if mtype != nil {
			sp.Add("x")
		}

		r.L(`tt.setup($0)`, sp)
	}

	// Now, render call and its handling.

	// Set up call params.
	cp := &gogh.Commas{}
outer:
	for i := 0; i < s.Params().Len(); i++ {
		p := s.Params().At(i)
		if isContext(p.Type()) {
			// All context.Context params will be stuffed with the same ctx.
			// It is possible to have two or more at once, at least nothing
			// can prevent some random coder from making so. To be resolved
			// manually then.
			cp.Add("ctx")
			continue
		}

		for _, amock := range amocks {
			if amock.Name == p.Name() {
				cp.Add("amocks." + amock.Name)
				continue outer
			}
		}

		name := argfields.MustGet(p.Name())
		cp.Add("tt." + name)
	}

	var recvPrefix string
	if s.Recv() != nil {
		recvPrefix = "x."
	}

	if s.Results().Len() == 0 {
		r.L(`$0$1($2)`, recvPrefix, f.Name(), cp)
	} else {
		rv := &gogh.Commas{}
		var results []string
		for i := 0; i < s.Results().Len(); i++ {
			isLast := i == s.Results().Len()-1

			if isLast && isError(s.Results().At(i).Type()) {
				rv.Add("err")
				continue
			}

			gotname := "got" + strings.TrimPrefix(resfields.MustGet(i), "want")
			rv.Add(gotname)
			results = append(results, gotname)
		}

		r.L(`$0 := $1$2($3)`, rv, recvPrefix, f.Name(), cp)
		if isErrored(s) {
			r.L(`switch {`)
			r.L(`case err != nil && (tt.wantErr || tt.$0 != nil):`, errcheck)
			r.L(`    if tt.$0 != nil {`, errcheck)
			r.L(`        if cerr := tt.$0(err); cerr != nil {`, errcheck)
			g.msgr.InvalidError(r, "cerr")
			r.L(`            return`)
			r.L(`        }`)
			r.L(`    }`)
			g.msgr.ExpectedError(r)
			r.L(`    return`)
			r.L(`case err != nil && !tt.wantErr:`)
			g.msgr.UnexpectedError(r)
			r.L(`    return`)
			r.L(`case err == nil && tt.wantErr:`)
			g.msgr.ErrorWasExpected(r)
			r.L(`    return`)
			r.L(`case err == nil && !tt.wantErr:`)
			r.L(`}`)
		}

		for i := 0; i < s.Results().Len(); i++ {
			rv := s.Results().At(i)
			if i == s.Results().Len()-1 && isError(rv.Type()) {
				continue
			}

			wantname := resfields.MustGet(i)
			gotname := results[i]
			r.Imports().Add(deepequalPath).Ref("de")
			r.L(`if !$de.Equal(tt.$0, $1) {`, wantname, gotname)
			if rv.Name() != "" {
				r.L(
					`    $de.SideBySide(t, "the return value for $0", tt.$1, $2)`,
					rv.Name(),
					wantname,
					gotname,
				)
			} else {
				r.L(
					`    $de.SideBySide(t, "the return value index $0 ($1)", tt.$2, $3)`,
					i,
					r.Type(rv.Type()),
					wantname,
					gotname,
				)
			}
			r.L(`}`)
		}
	}

	r.L(`        })`)
	r.L(`    }`)

	r.L(`}`)
}

func (g *Generator) renderTestStructure(
	r *goRenderer,
	hasMocksInType bool,
	mtype *types.Named,
	amocks []MockLookupResult,
	s *types.Signature,
) (
	argfields *ordmap.OrderedMap[string, string],
	resfields *ordmap.OrderedMap[int, string],
	errcheck string,
) {
	r = r.Scope()

	if len(amocks) > 0 {
		r.L(`    type argMocks struct{`)
		for _, amock := range amocks {
			r.L(`        $0 *$1`, amock.Name, r.Type(amock.Named))
		}
		r.L(`    }`)
		r.N()
	}

	r.L(`    type test struct{`)
	r.Uniq("name")

	r.L(`       name string`)

	// Render setup function arguments.
	rr := r.Scope()
	var setupArgs gogh.Params
	if hasMocksInType {
		setupArgs.Add(rr.Uniq("ctrl"), r.S("*$gomock.Controller"))
	}
	rr.Uniq("row")
	setupArgs.Add("row", "*test")
	if hasMocksInType {
		setupArgs.Add(r.Uniq("m"), r.S("*${mockertype}"))
	}
	if len(amocks) > 0 {
		setupArgs.Add(rr.Uniq("amocks"), "*argMocks")
	}
	if s.Recv() != nil {
		setupArgs.Add(
			rr.Uniq(gogh.Private(r.Type(mtype))),
			"*"+r.Type(mtype),
		)
	}

	if len(amocks) > 0 || hasMocksInType {
		r.L(`        setup func($0)`, &setupArgs)
	}

	// Render fields refered to function non-interface arguments.
	r.N()
	argfields = ordmap.New[string, string]()
outer:
	for i := 0; i < s.Params().Len(); i++ {
		param := s.Params().At(i)
		if isContext(param.Type()) {
			continue outer
		}

		for _, amock := range amocks {
			if amock.Name == param.Name() {
				continue outer
			}
		}

		argfield := r.Uniq(param.Name(), "arg")
		argfields.Set(param.Name(), argfield)
		r.L(`        $0 $1`, argfield, r.Type(param.Type()))
	}

	// Render fields for expected return values and error check.
	r.N()
	resfields = ordmap.New[int, string]()
	wantErr := r.Uniq("wantErr")
	errCheck := r.Uniq("errCheck")
	for i := 0; i < s.Results().Len(); i++ {
		if i == s.Results().Len()-1 && isError(s.Results().At(i).Type()) {
			r.L(`        $0 bool`, wantErr)
			r.L(`        $0 func(err error) error`, errCheck)
			continue
		}

		res := s.Results().At(i)

		var name string
		if res.Name() != "" {
			name = r.Uniq(gogh.Private("want", res.Name()))
		} else {
			name = r.Uniq("want", strconv.Itoa(i+1))
		}
		resfields.Set(i, name)

		r.L(`        $0 $1`, name, r.Type(s.Results().At(i).Type()))
	}

	r.L(`    }`)

	return argfields, resfields, errCheck
}

func (g *Generator) generateTypeMocker(p *goPackage, s *types.Signature, mocks []MockLookupResult) error {
	tn := s.Recv().Type().(*types.Pointer).Elem().(*types.Named).Obj()

	filename, typename := g.mockerNames(tn)
	fn := strings.TrimSuffix(filename, ".go") + ".go"

	r := p.Go(fn, gogh.Shy)

	r.Let("type", tn.Name())
	r.Let("mockertype", typename)

	r.Imports().Add("sync").Ref("sync")
	r.Imports().Add(gomockPath).Ref("gomock")

	r.Uniq(tn.Name())
	var fieldNames []string
	for _, mock := range mocks {
		fieldNames = append(fieldNames, r.Uniq(mock.Name, "mock"))
	}
	wn := r.Uniq("waiter")

	r.L(`// Creates new mocker instance for ${type}.`)
	r.L(`func new${mockertype|P}(ctrl *$gomock.Controller) *${mockertype} {`)
	r.L(`    return &${mockertype}{`)
	for i, mock := range mocks {
		var cname string
		tname := r.Type(mock.Named)
		if strings.Contains(tname, ".") {
			cname = strings.Replace(tname, ".", ".New", 1)
		} else {
			cname = "New" + tname
		}
		r.L(`        $0: $1(ctrl),`, fieldNames[i], cname)
	}
	r.N()
	r.L(`    }`)
	r.L(`}`)
	r.N()
	r.L(`// ${mockertype} repository for mockers of type ${type}.`)
	r.L(`type ${mockertype} struct{`)
	for i, mock := range mocks {
		r.L(`    $0 *$1`, fieldNames[i], r.Type(mock.Named))
	}
	r.L(`        $0 $sync.WaitGroup`, wn)
	r.L(`}`)
	r.N()
	r.L(`// waiters sets expected count of background processes to wait before finish the test.`)
	r.L(`func (m *${mockertype}) waiters(i int) {`)
	r.L(`    m.$0.Add(i)`, wn)
	r.L(`}`)
	r.N()
	r.L(`// end is called in a background process to reduce waiting count which is set by waiters call.`)
	r.L(`// It should be bound to the last call made in each background process to wait for. Something like:`)
	r.L(`//`)
	r.L(`//    m.Mock.EXPECT().….Do(m.end)`)
	r.L(`func (m *${mockertype}) end(...any) {`)
	r.L(`    m.$0.Done()`, wn)
	r.L(`}`)
	r.N()
	r.L(`// wait for background processes to stop.`)
	r.L(`func (m *${mockertype}) wait() {`)
	r.L(`    m.$0.Wait()`, wn)
	r.L(`}`)
	r.N()
	r.L(`// ${0|p} creates $0 instance with mocks.`, tn.Name())
	r.L(`func (m *${mockertype}) ${type}() *$type {`)
	r.L(`    // User defined.`)
	r.L(`}`)
	r.N()

	return nil
}

func (g *Generator) getMocksOfArguments(s *types.Signature, f types.Object) (res []MockLookupResult, _ error) {
	for i := 0; i < s.Params().Len(); i++ {
		p := s.Params().At(i)

		vn, ok := p.Type().(*types.Named)
		if !ok || !underlyingTypeIs[*types.Interface](vn) {
			g.infoParamNotInterfaceOmit(p.Pos(), f.Name())
			continue
		}

		if g.shouldNotBeMocked(vn) {
			continue
		}

		if _, ok := vn.Underlying().(*types.Interface); !ok {
			g.infoParamNotInterfaceOmit(p.Pos(), f.Name())
			continue
		}

		mockData, err := g.mockLookup(g, vn)
		if err != nil {
			return nil, errors.Wrapf(err, "look for a mock for type %s", vn)
		}

		mockData.Name = p.Name()
		res = append(res, mockData)
	}

	return res, nil
}

func (g *Generator) getMocksOfType(s *types.Signature) (res []MockLookupResult, _ error) {
	if s.Recv() == nil {
		return nil, nil
	}

	var t *types.Struct
	if vv, ok := s.Recv().Type().(*types.Pointer); ok {
		t = vv.Elem().(*types.Named).Underlying().(*types.Struct)
	} else {
		t = s.Recv().Type().(*types.Named).Underlying().(*types.Struct)
	}
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)

		if f.Anonymous() {
			message.Warningf("%s embedded fields are not supported", g.fset.Position(f.Pos()))
			continue
		}

		vn, ok := f.Type().(*types.Named)
		if !ok || !underlyingTypeIs[*types.Interface](vn) {
			g.infoFieldNotInterfaceOmit(f.Pos(), f.Name())
			continue
		}

		if g.shouldNotBeMocked(vn) {
			continue
		}

		mockData, err := g.mockLookup(g, vn)
		if err != nil {
			return res, errors.Wrapf(err, "look for a mock for type %s", vn)
		}

		mockData.Name = f.Name()
		res = append(res, mockData)
	}

	return res, nil
}
