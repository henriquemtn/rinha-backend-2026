# Mudancas principais da API

Este documento resume o que mudou na API durante a refatoracao, comparando a implementacao inicial com a estrutura atual.

## Antes

A API estava concentrada dentro da pasta `api/`, com quase toda a logica no pacote `main`.

Principais caracteristicas:

- `api/main.go` subia o servidor e tambem controlava carregamento de recursos.
- `api/handlers.go` tinha o roteamento, readiness e endpoint `/fraud-score`.
- `api/vector.go`, `api/payload_parser.go`, `api/score.go`, `api/resources.go` e `api/ivf.go` ficavam todos no mesmo pacote `main`.
- As referencias eram carregadas em runtime a partir de `references.json.gz`.
- O indice IVF era construido pela propria API, com configuracao por variaveis como `IVF_K`, `IVF_TRAIN_SAMPLES`, `IVF_ITER` e `N_PROBE`.
- O score fazia fallback para full scan quando o IVF ainda nao estava pronto.
- O Dockerfile compilava apenas um binario da pasta `api`.
- O compose ainda apontava para caminhos antigos da pasta `api`.

Essa versao funcionava como uma implementacao inicial, mas misturava responsabilidades e deixava a API carregando/trabalhando mais no startup.

## Depois

A API foi reorganizada para seguir um layout mais proximo de projetos reais e robustos, com comandos e pacotes internos separados.

Estrutura atual:

```text
cmd/
  api/
  indexer/
internal/
  consts/
  ivf/
  quantize/
  simd/
  vector/
container/
  Dockerfile
  docker-compose.yml
  nginx.conf
tests/
  smoke.js
  test.js
  test-data.json
```

Principais mudancas:

- O modulo Go saiu de `api/` e foi para a raiz do repositorio.
- A API agora fica em `cmd/api`.
- O gerador de indice fica em `cmd/indexer`.
- A logica de vetor foi isolada em `internal/vector`.
- A logica de busca IVF foi isolada em `internal/ivf`.
- A quantizacao ficou em `internal/quantize`.
- As constantes compartilhadas ficaram em `internal/consts`.
- A distancia otimizada ficou em `internal/simd`, com fallback para ambientes sem CGO.
- O indice `ivf.bin` agora e gerado durante o build da imagem Docker.
- Em runtime, a API apenas carrega `normalization.json`, `mcc_risk.json` e `ivf.bin`.
- O endpoint `/ready` continua retornando `200` quando a API esta pronta.
- O endpoint `/fraud-score` continua recebendo a transacao e retornando `approved` e `fraud_score`.

## Mudanca importante no fluxo de build

Antes:

```text
API sobe -> carrega resources -> monta/usa indice em runtime
```

Depois:

```text
Docker build -> roda indexer -> gera dataset/ivf.bin -> imagem final sobe API pronta para carregar o indice
```

Isso reduz o trabalho que a API precisa fazer quando o container inicia.

## Docker e Nginx

O Dockerfile foi alinhado ao modelo:

- usa `golang:1.24-alpine` como builder;
- instala `build-base` para permitir CGO;
- compila `cmd/indexer`;
- compila `cmd/api`;
- roda o indexer durante o build;
- copia apenas o binario da API e o dataset necessario para a imagem final.

O `docker-compose.yml` tambem foi alinhado ao modelo:

- usa duas instancias da API;
- usa sockets Unix em `/run/sock/api1.sock` e `/run/sock/api2.sock`;
- usa Nginx como proxy na porta `9999`;
- define `GOMAXPROCS`, `GOMEMLIMIT`, `N_PROBE_FAST` e `N_PROBE_FULL`;
- usa rede dedicada `rinha`.

O `nginx.conf` foi ajustado para:

- balancear entre `api1.sock` e `api2.sock`;
- desabilitar buffering;
- usar keepalive no upstream;
- manter timeouts curtos;
- trabalhar com payload pequeno, adequado ao endpoint da rinha.

## Testes

Foram adicionados testes k6:

- `tests/smoke.js`: teste rapido para confirmar que a API responde `200` e devolve JSON valido.
- `tests/test.js`: teste de carga com dataset, validando latencia, erros HTTP e acertos/erros de deteccao.

O `smoke.js` serve para validar se a stack esta viva.

O `test.js` serve para medir comportamento sob carga e gerar o resumo em:

```text
tests/results.json
```

## Resumo

A mudanca principal foi sair de uma API monolitica e crua, concentrada em `api/`, para uma estrutura mais organizada:

- comandos em `cmd/`;
- logica reutilizavel em `internal/`;
- indice IVF pre-gerado no build;
- runtime mais simples;
- Docker e Nginx alinhados ao projeto modelo;
- testes k6 adicionados para smoke e benchmark local.
