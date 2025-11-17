load('ext://restart_process', 'docker_build_with_restart')

# DB
k8s_yaml('k8s/dev/db.yaml')
k8s_resource('db')

# Auth
local_resource(
  'compile-auth',
  'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/service ./cmd/auth',
  deps=['cmd/auth/', 'internal/auth/', 'pkg/', 'go.mod', 'go.sum'],
  ignore=['build/']
)

docker_build_with_restart(
  'registry:12030/xeodocs-backend-auth',
  '.',
  entrypoint=['/service'],
  dockerfile='dockerfiles/Dockerfile.auth',
  live_update=[sync('./build', '/')]
)

k8s_yaml('k8s/dev/auth.yaml')
k8s_resource('auth', resource_deps=['compile-auth', 'db'])

# Gateway
local_resource(
  'compile-gateway',
  'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/service ./cmd/gateway',
  deps=['cmd/gateway/', 'pkg/', 'go.mod', 'go.sum'],
  ignore=['build/']
)

docker_build_with_restart(
  'registry:12030/xeodocs-backend-gateway',
  '.',
  entrypoint=['/service'],
  dockerfile='dockerfiles/Dockerfile.gateway',
  live_update=[sync('./build', '/')]
)

k8s_yaml('k8s/dev/gateway.yaml')
k8s_resource('gateway', port_forwards='12020:12020', resource_deps=['compile-gateway', 'auth'])
