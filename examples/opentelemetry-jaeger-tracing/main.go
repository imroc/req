package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"log"
	"opentelemetry-jaeger-tracing/github"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "github-query"

var githubClient *github.Client

func traceProvider() (*trace.TracerProvider, error) {
	// Create the Jaeger exporter
	ep := os.Getenv("JAEGER_ENDPOINT")
	if ep == "" {
		ep = "http://localhost:14268/api/traces"
	}
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(ep)))
	if err != nil {
		return nil, err
	}

	// Record information about this application in a Resource.
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "test"),
		),
	)

	// Create the TraceProvider.
	tp := trace.NewTracerProvider(
		// Always be sure to batch in production.
		trace.WithBatcher(exp),
		// Record information about this application in a Resource.
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	return tp, nil
}

// QueryUser queries information for specified GitHub user, and display a
// brief introduction which includes name, blog, and the most popular repo.
func QueryUser(username string) error {
	ctx, span := otel.Tracer("query").Start(context.Background(), "QueryUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("query.username", username),
	)
	profile, err := githubClient.GetUserProfile(ctx, username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetAttributes(
		attribute.String("query.name", profile.Name),
		attribute.String("result.blog", profile.Blog),
	)
	repo, err := findMostPopularRepo(ctx, username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetAttributes(
		attribute.String("popular.repo.name", repo.Name),
		attribute.Int("popular.repo.star", repo.Star),
	)
	fmt.Printf("The most popular repo of %s (%s) is %s, with %d stars\n", profile.Name, profile.Blog, repo.Name, repo.Star)
	return nil
}

func findMostPopularRepo(ctx context.Context, username string) (repo *github.Repo, err error) {
	ctx, span := otel.Tracer("query").Start(ctx, "findMostPopularRepo")
	defer span.End()

	for page := 1; ; page++ {
		var repos []*github.Repo
		repos, err = githubClient.ListUserRepo(ctx, username, page)
		if err != nil {
			return
		}
		if len(repos) == 0 {
			break
		}
		if repo == nil {
			repo = repos[0]
		}
		for _, rp := range repos[1:] {
			if rp.Star >= repo.Star {
				repo = rp
			}
		}
		if len(repos) == 100 {
			continue
		}
		break
	}

	if repo == nil {
		err = fmt.Errorf("no repo found for %s", username)
	}
	return
}

func main() {
	tp, err := traceProvider()
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(tp)

	githubClient = github.NewClient()
	if os.Getenv("DEBUG") == "on" {
		githubClient.SetDebug(true)
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		githubClient.LoginWithToken(token)
	}
	githubClient.SetTracer(otel.Tracer("github"))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigs
		fmt.Printf("Caught %s, shutting down\n", sig)
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	for {
		var name string
		fmt.Printf("Please give a github username: ")
		_, err := fmt.Fscanf(os.Stdin, "%s\n", &name)
		if err != nil {
			panic(err)
		}
		err = QueryUser(name)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
