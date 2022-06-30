package logging

//go:generate sh -c "mockgen -package logging -self_package github.com/imroc/req/v3/internal/logging -destination mock_connection_tracer_test.go github.com/imroc/req/v3/internal/logging ConnectionTracer"
//go:generate sh -c "mockgen -package logging -self_package github.com/imroc/req/v3/internal/logging -destination mock_tracer_test.go github.com/imroc/req/v3/internal/logging Tracer"
