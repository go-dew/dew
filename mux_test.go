package dew_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-dew/dew"
)

func TestMux_BasicCommand(t *testing.T) {
	mux := dew.New()
	mux.Register(new(userHandler))
	mux.Register(new(postHandler))
	ctx := dew.NewContext(context.Background(), mux)

	createUser, err := dew.Dispatch(ctx, &createUser{Name: "john"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createUser.Result != "user created" {
		t.Fatalf("unexpected result: %s", createUser.Result)
	}

	createPost, err := dew.Dispatch(ctx, &createPost{Title: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createPost.Result != "post created" {
		t.Fatalf("unexpected result: %s", createPost.Result)
	}
}

func TestMux_DispatchError(t *testing.T) {
	t.Run("BusNotFound", func(t *testing.T) {
		ctx := context.Background()
		err := dew.DispatchMulti(ctx, dew.NewAction(&createUser{Name: "john"}))
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "bus not found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("ResolveError", func(t *testing.T) {
		mux := dew.New()
		ctx := dew.NewContext(context.Background(), mux)
		err := dew.DispatchMulti(ctx, dew.NewAction(&createUser{Name: "john"}))
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "handler not found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMux_QueryError(t *testing.T) {
	t.Run("BusNotFound", func(t *testing.T) {
		_, err := dew.Query(context.Background(), &findUser{ID: 1})
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "bus not found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("ResolveError", func(t *testing.T) {
		mux := dew.New()
		ctx := dew.NewContext(context.Background(), mux)
		_, err := dew.Query(ctx, &findUser{ID: 1})
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "handler not found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMux_QueryAsyncError(t *testing.T) {
	t.Run("BusNotFound", func(t *testing.T) {
		err := dew.QueryAsync(context.Background(), dew.NewQuery(&findUser{ID: 1}))
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "bus not found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("ResolveError", func(t *testing.T) {
		mux := dew.New()
		ctx := dew.NewContext(context.Background(), mux)
		err := dew.QueryAsync(ctx, dew.NewQuery(&findUser{ID: 1}))
		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "handler not found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMux_ValueTypeHandler(t *testing.T) {
	var userHandler userHandler

	mux := dew.New()
	mux.Register(userHandler)
	ctx := dew.NewContext(context.Background(), mux)

	createUser := &createUser{Name: "john"}
	testRunDispatch(t, ctx, dew.NewAction(createUser))
	if createUser.Result != "user created" {
		t.Fatalf("unexpected result: %s", createUser.Result)
	}
}

func TestMux_HandlerNotFound(t *testing.T) {
	mux := dew.New()
	ctx := dew.NewContext(context.Background(), mux)

	action := dew.NewAction(&createUser{Name: "john"})
	err := dew.DispatchMulti(ctx, action)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestMux_Query(t *testing.T) {
	mux := dew.New()
	mux.Register(new(userHandler))
	ctx := dew.NewContext(context.Background(), mux)

	// Test successful query
	result := testRunQuery(t, ctx, &findUser{ID: 1})
	if result.Result != "john" {
		t.Fatalf("unexpected result: %s", result.Result)
	}

	// Test query error
	_, err := dew.Query(ctx, &findUser{ID: 2})
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
	if !errors.Is(err, errUserNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMux_QueryAsync(t *testing.T) {
	mux := dew.New()

	var queryCount atomic.Int32

	mux.UseQuery(func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			queryCount.Add(1)
			return next.Handle(ctx)
		})
	})

	mux.Group(func(mux dew.Bus) {
		mux.Register(dew.HandlerFunc[findUser](
			func(ctx context.Context, query *findUser) error {
				time.Sleep(100 * time.Millisecond)
				query.Result = fmt.Sprintf("user-%d", query.ID)
				return nil
			},
		))
	})

	mux.Register(dew.HandlerFunc[findPost](
		func(ctx context.Context, query *findPost) error {
			time.Sleep(100 * time.Millisecond)
			query.Result = fmt.Sprintf("post-%d", query.ID)
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	commands := dew.Commands{
		dew.NewQuery(&findUser{ID: 1}),
		dew.NewQuery(&findUser{ID: 2}),
		dew.NewQuery(&findUser{ID: 3}),
		dew.NewQuery(&findPost{ID: 1}),
		dew.NewQuery(&findPost{ID: 2}),
		dew.NewQuery(&findPost{ID: 3}),
	}

	// count time
	now := time.Now()

	// query
	err := dew.QueryAsync(ctx, commands...)
	if err != nil {
		t.Fatal(err)
	}

	d := time.Since(now)
	if !(80*time.Millisecond <= d && d <= 120*time.Millisecond) {
		t.Fatalf("unexpected time: %v", time.Since(now))
	}

	if queryCount.Load() != 1 {
		t.Fatalf("unexpected query count: %d", queryCount.Load())
	}

	for _, query := range commands {
		switch query := query.Command().(type) {
		case *findPost:
			if query.Result != fmt.Sprintf("post-%d", query.ID) {
				t.Fatalf("unexpected result: %s", query.Result)
			}
		case *findUser:
			if query.Result != fmt.Sprintf("user-%d", query.ID) {
				t.Fatalf("unexpected result: %s", query.Result)
			}
		default:
			t.Fatalf("unexpected query type: %T", query)
		}
	}

	// empty queries
	err = dew.QueryAsync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMux_QueryAsync_Error(t *testing.T) {
	mux := dew.New()

	var (
		errUserNotFound = errors.New("user not found")
		errPostNotFound = errors.New("post not found")
	)

	mux.Register(dew.HandlerFunc[findUser](
		func(ctx context.Context, query *findUser) error {
			return errUserNotFound
		},
	))

	mux.Register(dew.HandlerFunc[findPost](
		func(ctx context.Context, query *findPost) error {
			return errPostNotFound
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	commands := dew.Commands{
		dew.NewQuery(&findUser{ID: 1}),
		dew.NewQuery(&findPost{ID: 1}),
	}

	// query
	err := dew.QueryAsync(ctx, commands...)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
	if !errors.Is(err, errUserNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !errors.Is(err, errPostNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMux_Reentrant(t *testing.T) {
	mux := dew.New()
	mux.Register(new(userHandler))
	mux.Register(new(postHandler))

	type findUserPost struct {
		ID     int
		Result struct {
			User string
			Post string
		}
	}

	mux.Register(dew.HandlerFunc[findUserPost](
		func(ctx context.Context, query *findUserPost) error {
			findUserQuery, err := dew.Query(ctx, &findUser{ID: query.ID})
			if err != nil {
				return err
			}
			postQuery, err := dew.Query(ctx, &findPost{ID: query.ID})
			if err != nil {
				return err
			}
			query.Result.User = findUserQuery.Result
			query.Result.Post = postQuery.Result
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	query, err := dew.Query(ctx, &findUserPost{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query.Result.User != "john" {
		t.Fatalf("unexpected result: %s", query.Result.User)
	}
	if query.Result.Post != "hello" {
		t.Fatalf("unexpected result: %s", query.Result.Post)
	}
}

type ctxKey struct {
	name string
}

func TestMux_Middlewares(t *testing.T) {
	mux := dew.New()
	mux.Use(dew.ALL, func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			return next.Handle(ctx.WithValue(ctxKey{"mw"}, "[all]"))
		})
	})
	mux.Use(dew.ACTION, func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			return next.Handle(ctx.WithValue(ctxKey{"mw2"}, "[action]"))
		})
	})
	mux.Use(dew.QUERY, func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			return next.Handle(ctx.WithValue(ctxKey{"mw2"}, "[query]"))
		})
	})
	mux.Register(dew.HandlerFunc[createUser](
		func(ctx context.Context, command *createUser) error {
			command.Result = ctx.Value(ctxKey{"mw"}).(string) + ctx.Value(ctxKey{"mw2"}).(string)
			return nil
		},
	))
	mux.Register(dew.HandlerFunc[findUser](
		func(ctx context.Context, query *findUser) error {
			query.Result = ctx.Value(ctxKey{"mw"}).(string) + ctx.Value(ctxKey{"mw2"}).(string)
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	command := &createUser{Name: "test"}
	testRunDispatch(t, ctx, dew.NewAction(command))
	if command.Result != "[all][action]" {
		t.Fatalf("unexpected result: %s", command.Result)
	}

	query := &findUser{ID: 1}
	result, err := dew.Query(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if result.Result != "[all][query]" {
		t.Fatalf("unexpected result: %s", result.Result)
	}

	// dispatch no action

	if err := dew.DispatchMulti(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMux_DispatchMiddlewares(t *testing.T) {
	mux := dew.New()
	var dispatchCount atomic.Int32

	mux.UseDispatch(func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			dispatchCount.Add(1)
			return next.Handle(ctx)
		})
	})
	mux.Register(dew.HandlerFunc[createUser](
		func(ctx context.Context, command *createUser) error {
			command.Result = command.Name
			return nil
		},
	))
	mux.Register(dew.HandlerFunc[findUser](
		func(ctx context.Context, query *findUser) error {
			query.Result = fmt.Sprintf("user-%d", query.ID)
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	createUsers := []*createUser{
		{Name: "test"},
		{Name: "john"},
	}

	// query
	findUser, err := dew.Query(ctx, &findUser{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if findUser.Result != "user-1" {
		t.Fatalf("unexpected result: %s", findUser.Result)
	}

	// check if dispatch middleware is called only once
	if dispatchCount.Load() != 0 {
		t.Fatalf("unexpected middleware call count: %d", dispatchCount.Load())
	}

	// multiple commands
	if err := dew.DispatchMulti(ctx,
		dew.NewAction(createUsers[0]),
		dew.NewAction(createUsers[1]),
	); err != nil {
		t.Fatal(err)
	}

	// check if dispatch middleware is called only once
	if dispatchCount.Load() != 1 {
		t.Fatalf("unexpected middleware call count: %d", dispatchCount.Load())
	}
	for _, cmd := range createUsers {
		if cmd.Result != cmd.Name {
			t.Fatalf("unexpected result: %s", cmd.Result)
		}
	}
}

func TestMux_QueryMiddlewares(t *testing.T) {
	mux := dew.New()
	var dispatchCount atomic.Int32

	mux.UseQuery(func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			dispatchCount.Add(1)
			return next.Handle(ctx)
		})
	})
	mux.Register(dew.HandlerFunc[findUser](
		func(ctx context.Context, query *findUser) error {
			query.Result = fmt.Sprintf("user-%d", query.ID)
			return nil
		},
	))
	mux.Register(dew.HandlerFunc[createUser](
		func(ctx context.Context, command *createUser) error {
			command.Result = command.Name
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	// multiple commands
	createUser := &createUser{Name: "test"}
	if err := dew.DispatchMulti(ctx, dew.NewAction(createUser)); err != nil {
		t.Fatal(err)
	}

	// check if dispatch middleware is called only once
	if dispatchCount.Load() != 0 {
		t.Fatalf("unexpected middleware call count: %d", dispatchCount.Load())
	}

	if createUser.Result != createUser.Name {
		t.Fatalf("unexpected result: %s", createUser.Result)
	}

	// query
	findUser, err := dew.Query(ctx, &findUser{ID: 1})
	if err != nil {
		t.Fatal(err)
	}

	// check if dispatch middleware is called only once
	if dispatchCount.Load() != 1 {
		t.Fatalf("unexpected middleware call count: %d", dispatchCount.Load())
	}

	if findUser.Result != "user-1" {
		t.Fatalf("unexpected result: %s", findUser.Result)
	}
}

func TestMux_Groups(t *testing.T) {
	mux := dew.New()
	mux.Use(dew.ALL, func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			return next.Handle(ctx.WithValue(ctxKey{"global"}, "[global]"))
		})
	})

	mux.Group(func(mux dew.Bus) {
		mux.Use(dew.ACTION, func(next dew.Middleware) dew.Middleware {
			return dew.MiddlewareFunc(func(ctx dew.Context) error {
				return next.Handle(ctx.WithValue(ctxKey{"local"}, "[user-action]"))
			})
		})
		mux.Register(dew.HandlerFunc[createUser](
			func(ctx context.Context, command *createUser) error {
				command.Result = ctx.Value(ctxKey{"global"}).(string) + ctx.Value(ctxKey{"local"}).(string) + command.Name
				return nil
			},
		))
	})

	mux.Group(func(mux dew.Bus) {
		mux.Use(dew.ACTION, func(next dew.Middleware) dew.Middleware {
			return dew.MiddlewareFunc(func(ctx dew.Context) error {
				return next.Handle(ctx.WithValue(ctxKey{"local"}, "[post-action]"))
			})
		})
		mux.Register(dew.HandlerFunc[createPost](
			func(ctx context.Context, command *createPost) error {
				command.Result = ctx.Value(ctxKey{"global"}).(string) + ctx.Value(ctxKey{"local"}).(string) + command.Title
				return nil
			},
		))
	})

	mux.Register(dew.HandlerFunc[updateUser](
		func(ctx context.Context, command *updateUser) error {
			command.Result = ctx.Value(ctxKey{"global"}).(string)
			if ctx.Value(ctxKey{"local"}) != nil {
				command.Result += ctx.Value(ctxKey{"local"}).(string)
			}
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	createUser := &createUser{Name: "john"}
	testRunDispatch(t, ctx, dew.NewAction(createUser))
	if createUser.Result != "[global][user-action]john" {
		t.Fatalf("unexpected result: %s", createUser.Result)
	}

	createPost := &createPost{Title: "hello"}
	testRunDispatch(t, ctx, dew.NewAction(createPost))
	if createPost.Result != "[global][post-action]hello" {
		t.Fatalf("unexpected result: %s", createPost.Result)
	}

	updateUser := &updateUser{}
	testRunDispatch(t, ctx, dew.NewAction(updateUser))
	if updateUser.Result != "[global]" {
		t.Fatalf("unexpected result: %s", updateUser.Result)
	}
}

func TestMux_GroupsQuery(t *testing.T) {
	mux := dew.New()
	mux.Use(dew.ALL, func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			return next.Handle(ctx.WithValue(ctxKey{"global"}, "[global]"))
		})
	})

	mux.Group(func(mux dew.Bus) {
		mux.Use(dew.QUERY, func(next dew.Middleware) dew.Middleware {
			return dew.MiddlewareFunc(func(ctx dew.Context) error {
				return next.Handle(ctx.WithValue(ctxKey{"local"}, "[local1]"))
			})
		})
		mux.Register(dew.HandlerFunc[findUser](
			func(ctx context.Context, command *findUser) error {
				command.Result = ctx.Value(ctxKey{"global"}).(string) + ctx.Value(ctxKey{"local"}).(string) + "john"
				return nil
			},
		))
	})

	mux.Group(func(mux dew.Bus) {
		mux.Use(dew.QUERY, func(next dew.Middleware) dew.Middleware {
			return dew.MiddlewareFunc(func(ctx dew.Context) error {
				return next.Handle(ctx.WithValue(ctxKey{"local"}, "[local2]"))
			})
		})
		mux.Register(dew.HandlerFunc[findPost](
			func(ctx context.Context, command *findPost) error {
				command.Result = ctx.Value(ctxKey{"global"}).(string) + ctx.Value(ctxKey{"local"}).(string) + "post"
				return nil
			},
		))
	})

	type findTagQuery struct {
		Result string
	}

	mux.Register(dew.HandlerFunc[findTagQuery](
		func(ctx context.Context, command *findTagQuery) error {
			command.Result = ctx.Value(ctxKey{"global"}).(string)
			if ctx.Value(ctxKey{"local"}) != nil {
				command.Result += ctx.Value(ctxKey{"local"}).(string)
			}
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	findUser := testRunQuery(t, ctx, &findUser{ID: 1})
	if findUser.Result != "[global][local1]john" {
		t.Fatalf("unexpected result: %s", findUser.Result)
	}

	findPost := testRunQuery(t, ctx, &findPost{ID: 1})
	if findPost.Result != "[global][local2]post" {
		t.Fatalf("unexpected result: %s", findPost.Result)
	}

	findTag := testRunQuery(t, ctx, &findTagQuery{})
	if findTag.Result != "[global]" {
		t.Fatalf("unexpected result: %s", findTag.Result)
	}

}

func TestMux_ErrorHandling(t *testing.T) {
	mux := dew.New()
	mux.Register(new(userHandler))

	ctx := dew.NewContext(context.Background(), mux)

	createUser := &createUser{Name: ""}
	err := dew.DispatchMulti(ctx, dew.NewAction(createUser))
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
	if !errors.Is(err, errNameRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMux_Validation(t *testing.T) {
	mux := dew.New()
	mux.Register(new(postHandler))

	ctx := dew.NewContext(context.Background(), mux)

	err := dew.DispatchMulti(ctx, dew.NewAction(&createPost{Title: ""}))
	if err == nil {
		t.Fatal("expected a validation error, but got nil")
	}
	if !errors.Is(err, dew.ErrValidationFailed) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMux_BusContext(t *testing.T) {
	mux := dew.New()

	mux.UseDispatch(func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			bus := dew.MustFromContext(ctx.Context())
			if bus != mux {
				t.Fatal("expected bus not found")
			}
			if ctx.Command() != nil {
				t.Errorf("command should be nil")
			}
			return next.Handle(ctx)
		})
	})

	mux.Use(dew.ALL, func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			cmd := ctx.Command().(*createUser)
			if cmd.Name != "john" {
				t.Fatalf("unexpected command: %v", cmd)
			}
			return next.Handle(ctx.WithValue("key", "value"))
		})
	})

	mux.Register(dew.HandlerFunc[createUser](
		func(ctx context.Context, command *createUser) error {
			if ctx.Value("key") != "value" {
				t.Fatal("expected value not found")
			}
			bus := dew.MustFromContext(ctx)
			if bus != mux {
				t.Fatal("expected bus not found")
			}
			return nil
		},
	))

	ctx := dew.NewContext(context.Background(), mux)

	testRunDispatch(t, ctx, dew.NewAction(&createUser{Name: "john"}))
}

func BenchmarkMux(b *testing.B) {

	mux1 := dew.New()
	mux1.Register(new(userHandler))
	mux1.Register(new(postHandler))
	ctx1 := dew.NewContext(context.Background(), mux1)

	b.Run("query", func(b *testing.B) {

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = dew.Query(ctx1, &findUser{ID: 1})
		}
	})

	b.Run("dispatch", func(b *testing.B) {

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dew.DispatchMulti(ctx1, dew.NewAction(&createPost{Title: "john"}))
		}
	})

	mux2 := dew.New()
	mux2.Use(dew.ALL, func(next dew.Middleware) dew.Middleware {
		return dew.MiddlewareFunc(func(ctx dew.Context) error {
			return next.Handle(ctx)
		})
	})
	mux2.Group(func(mux dew.Bus) {
		mux.Use(dew.ALL, func(next dew.Middleware) dew.Middleware {
			return dew.MiddlewareFunc(func(ctx dew.Context) error {
				return next.Handle(ctx)
			})
		})
		mux.Register(new(userHandler))
		mux.Register(new(postHandler))
	})
	ctx2 := dew.NewContext(context.Background(), mux2)

	b.Run("query-with-middleware", func(b *testing.B) {

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = dew.Query(ctx2, &findUser{ID: 1})
		}
	})

	b.Run("dispatch-with-middleware", func(b *testing.B) {

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dew.DispatchMulti(ctx2, dew.NewAction(&createPost{Title: "john"}))
		}
	})
}

func testRunQuery[T dew.QueryAction](t *testing.T, ctx context.Context, query *T) *T {
	t.Helper()
	result, err := dew.Query[T](ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func testRunDispatch(t *testing.T, ctx context.Context, commands ...dew.CommandHandler[dew.Action]) {
	t.Helper()
	err := dew.DispatchMulti(ctx, commands...)
	if err != nil {
		t.Fatal(err)
	}
}

// -------------------------------
// commands (actions and queries)

type createUser struct {
	Name   string
	Result string
}

func (c createUser) Validate(_ context.Context) error { return nil }

type updateUser struct {
	Name   string
	Result string
}

func (c updateUser) Validate(_ context.Context) error { return nil }

type createPost struct {
	Title  string
	Result string
}

func (c createPost) Validate(_ context.Context) error {
	if c.Title == "" {
		return errors.New("title is required")
	}
	return nil
}

type findUser struct {
	ID     int
	Result string
}

type findPost struct {
	ID     int
	Result string
}

// ---------
// handlers

var (
	errNameRequired = errors.New("name is required")
	errUserNotFound = errors.New("user not found")
)

type userHandler struct{}

func (h *userHandler) CreateUser(_ context.Context, command *createUser) error {
	if command.Name == "" {
		return errNameRequired
	}
	command.Result = "user created"
	return nil
}

type postHandler struct{}

func (h *postHandler) CreatePost(_ context.Context, command *createPost) error {
	command.Result = "post created"
	return nil
}

func (h *postHandler) FindPost(_ context.Context, query *findPost) error {
	query.Result = "hello"
	return nil
}

func (*userHandler) FindUser(_ context.Context, query *findUser) error {
	if query.ID == 1 {
		query.Result = "john"
		return nil
	} else if query.ID == 2 {
		return errUserNotFound
	}
	return nil
}
